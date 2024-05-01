package tcpChannel

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

type ConnectionState struct {
	address net.Addr
	state   NET_EVENT
	err     error
}
type ProtoListenerInterface interface {
	//AddProtocol(protocol ProtoInterface) error
	StartProtocol(protocol ProtoInterface, networkQueueSize int, timeoutQueueSize int, localCommQueueSize int) error
	WaitForProtocolsToEnd(closeConnections bool)
	RemoveProtocol(id APP_PROTO_ID)
}
type protoWrapper struct {
	queue                   chan *NetworkEvent
	timeoutChannel          chan int
	localCommunicationQueue chan *LocalCommunicationEvent
	proto                   ProtoInterface
	messageHandlers         map[MessageHandlerID]MESSAGE_HANDLER_TYPE
	timerHandlers           map[int]*timerArgs
	timersId                int
	protoListener           *ProtoListener
}
type ProtocolAPI interface {
	//StartProtocols() error
	RegisterNetworkMessageHandler(handlerId MessageHandlerID, funcHandler MESSAGE_HANDLER_TYPE) error
	RegisterTimeout(duration time.Duration, data interface{}, funcToExecute TimerHandlerFunc) (int, error)
	SendLocalEvent(destProto APP_PROTO_ID, data interface{}, funcToExecute LocalProtoComHandlerFunc) error
	CancelTimer(timerId int) (bool, error)
	RegisterPeriodicTimeout(duration time.Duration, data interface{}, funcToExecute TimerHandlerFunc) (int, error)
	NetworkInterface() ChannelInterface
	SelfProtocol() ProtoInterface
}

type CustomPair[F any, S any] struct {
	First  F
	Second S
}
type timerArgs struct {
	protoId       APP_PROTO_ID
	data          interface{}
	funcHandler   TimerHandlerFunc
	timer         *time.Timer
	periodicTimer *time.Ticker
}
type ProtoListener struct {
	protocols      map[APP_PROTO_ID]*protoWrapper
	waitChannel    chan APP_PROTO_ID
	channel        ChannelInterface
	order          binary.ByteOrder
	ConnectionType CONNECTION_TYPE
}

/*********************** CLIENT METHODS ***************************/
func (p *protoWrapper) RegisterNetworkMessageHandler(handlerId MessageHandlerID, funcHandler MESSAGE_HANDLER_TYPE) error {
	var err error
	if p.messageHandlers[handlerId] == nil {
		p.messageHandlers[handlerId] = funcHandler
		err = nil
	} else {
		err = ELEMENT_EXISTS_ALREADY
	}
	return err
}

// address string, port int, connectionType gobabelUtils.CONNECTION_TYPE
func NewProtocolListener(address string, port int, connectionType CONNECTION_TYPE, order binary.ByteOrder) ProtoListenerInterface {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ch := NewTCPChannel(address, port, connectionType)
	protoL := &ProtoListener{
		protocols:      make(map[APP_PROTO_ID]*protoWrapper),
		waitChannel:    make(chan APP_PROTO_ID),
		channel:        ch,
		order:          order,
		ConnectionType: connectionType,
	}
	ch.SetProtoLister(protoL)
	return protoL
}
func (l *ProtoListener) WaitForProtocolsToEnd(closeConnections bool) {
	var protoID APP_PROTO_ID
	for {
		protoID = <-l.waitChannel
		delete(l.protocols, protoID)
		if len(l.protocols) == 0 {
			if closeConnections {
				l.channel.CloseConnections()
			}
			return
		}
	}
}

func (proto *protoWrapper) RegisterTimeout(duration time.Duration, data interface{}, funcToExecute TimerHandlerFunc) (int, error) {
	proto.timersId++
	aux := proto.timersId
	t := time.AfterFunc(duration, func() {
		proto.timeoutChannel <- aux
	})
	auxT := &timerArgs{
		protoId:     proto.proto.ProtocolUniqueId(),
		data:        data,
		funcHandler: funcToExecute,
		timer:       t,
	}
	proto.timerHandlers[proto.timersId] = auxT
	return aux, UNKNOWN_PROTOCOL
}

// I DONT LIKE IT BECAUSE FOR EACH TIMER I NEED TO HAVE A GOROUTINE
func (protoW *protoWrapper) RegisterPeriodicTimeout(duration time.Duration, data interface{}, funcToExecute TimerHandlerFunc) (int, error) {
	ticker := time.NewTicker(duration)
	protoW.timersId++
	protoW.timerHandlers[protoW.timersId] = &timerArgs{
		protoId:       protoW.proto.ProtocolUniqueId(),
		data:          data,
		funcHandler:   funcToExecute,
		timer:         nil,
		periodicTimer: ticker,
	}
	aux := protoW.timersId
	go func() {
		for range ticker.C {
			protoW.timeoutChannel <- aux
		}
	}()

	return aux, nil
}

func (protoW *protoWrapper) SendLocalEvent(destProto APP_PROTO_ID, data interface{}, funcToExecute LocalProtoComHandlerFunc) error {
	proto := protoW.protoListener.protocols[destProto]
	if proto == nil {
		return UNKNOWN_PROTOCOL
	}
	proto.localCommunicationQueue <- NewLocalCommunicationEvent(protoW.proto.ProtocolUniqueId(), proto.proto, data, funcToExecute)
	return nil
}

func (protoW *protoWrapper) NetworkInterface() ChannelInterface {
	return protoW.protoListener.channel
}

func (protoW *protoWrapper) SelfProtocol() ProtoInterface {
	return protoW.proto
}

func (protoW *protoWrapper) CancelTimer(timerId int) (bool, error) {
	args := protoW.timerHandlers[timerId]
	if args != nil {
		delete(protoW.timerHandlers, timerId)
	}
	if args != nil {
		if args.timer == nil {
			args.periodicTimer.Stop()
		} else {
			args.timer.Stop()
		}
	}
	return true, nil
}

func (l *ProtoListener) AddProtocol(protocol ProtoInterface, networkQueueSize int, timeoutQueueSize int, localCommQueueSize int) error {
	//TODO make the constants dynamic
	if l.protocols[(protocol).ProtocolUniqueId()] != nil {
		return PROTOCOL_EXIST_ALREADY
	}
	if ALL_PROTO_ID == protocol.ProtocolUniqueId() {
		return RESERVED_PROTOCOL_ID
	}
	l.protocols[(protocol).ProtocolUniqueId()] = &protoWrapper{
		queue:                   make(chan *NetworkEvent, networkQueueSize),
		proto:                   protocol,
		timeoutChannel:          make(chan int, timeoutQueueSize),
		localCommunicationQueue: make(chan *LocalCommunicationEvent, localCommQueueSize),
		messageHandlers:         make(map[MessageHandlerID]MESSAGE_HANDLER_TYPE),
		timerHandlers:           make(map[int]*timerArgs),
		protoListener:           l,
	}
	return nil
}
func (l *ProtoListener) RemoveProtocol(id APP_PROTO_ID) {
	proto := l.protocols[id]
	if proto != nil {
		//TODO USE LOCKS HERE
		delete(l.protocols, id)
		close(proto.queue)
		close(proto.timeoutChannel)
		close(proto.localCommunicationQueue)
		l.waitChannel <- id
	}
}

// TODO should all protocols receive connection up event ??
// TODO should a protocol be registered after all the protocols have already started
func (l *ProtoListener) StartProtocols() error {
	if len(l.protocols) == 0 {
		log.Fatal(NO_PROTOCOLS_TO_RUN)
		return NO_PROTOCOLS_TO_RUN
	}
	for _, protoWrapper := range l.protocols {
		l.auxRunProtocol(protoWrapper)
	}

	return nil

}
func (l *ProtoListener) StartProtocol(protocol ProtoInterface, networkQueueSize int, timeoutQueueSize int, localCommQueueSize int) error {
	err := l.AddProtocol(protocol, networkQueueSize, timeoutQueueSize, localCommQueueSize)
	if len(l.protocols) == 1 {
		if ch, ok := l.channel.(*tcpChannel); ok {
			ch.Start()
		}
	}
	if err == nil {
		l.auxRunProtocol(l.protocols[protocol.ProtocolUniqueId()])
	}
	return err
}

func (l *ProtoListener) auxRunProtocol(protoW *protoWrapper) {
	go func() {
		proto := protoW.proto
		proto.OnStart(protoW)
		log.Printf("PROTOCOL <%d> STARTED LISTENING TO EVENTS...\n", proto.ProtocolUniqueId())
		for {
			select {
			case networkEvent, ok := <-protoW.queue:
				if ok {
					//log.Printf("PROTOCOL <%d> RECEIVED AN EVENT. EVENT TYPE <%d>\n", proto.ProtocolUniqueId(), networkEvent.NET_EVENT)
					switch networkEvent.NET_EVENT {
					case CONNECTION_UP:
						proto.ConnectionUp(networkEvent.customConn, protoW)
					case CONNECTION_DOWN:
						proto.ConnectionDown(networkEvent.customConn, protoW)
					case MESSAGE_RECEIVED:
						if NO_NETWORK_MESSAGE_HANDLER_ID == networkEvent.MessageHandlerID {
							proto.OnMessageArrival(networkEvent.customConn, networkEvent.SourceProto, networkEvent.DestProto, networkEvent.Data, protoW)
						} else {
							messageHandler := protoW.messageHandlers[networkEvent.MessageHandlerID]
							if messageHandler == nil {
								log.Printf("RECEIVED A NETWORK MESSAGE TO AN INVALID MESSAGE HANDLER <%d>. DEST PROTO %d \n", networkEvent.MessageHandlerID, networkEvent.DestProto)
							} else {
								messageHandler(protoW, networkEvent.customConn, networkEvent.SourceProto, NewCustomReader(networkEvent.Data, l.order))
							}
						}
					default:
						(l.channel).CloseConnection(networkEvent.customConn.connectionKey)
						log.Fatal(fmt.Sprintf("RECEIVED AN EVENT NOT PART OF THE PROTOCOL. CONNECTION CLOSED %s", networkEvent.customConn.remoteListenAddr.String()))
					}
				} else {
					log.Println("PROTOCOL CHANNEL ENDED 1; ID: ", proto.ProtocolUniqueId())
					return
				}
			case timerId, ok := <-protoW.timeoutChannel:
				if ok {
					args := protoW.timerHandlers[timerId]
					if args != nil {
						if args.timer != nil {
							// it is a timeout, otherwise it is a periodic timer
							delete(protoW.timerHandlers, timerId)
						}
						args.funcHandler(timerId, args.protoId, args.data)
					}
				} else {
					log.Println("PROTOCOL CHANNEL ENDED 2; ID: ", proto.ProtocolUniqueId())
					return
				}
			case localEventCom, ok := <-protoW.localCommunicationQueue:
				if ok {
					localEventCom.ExecuteFunc()
				} else {
					log.Println("PROTOCOL CHANNEL ENDED 3; ID: ", proto.ProtocolUniqueId())
					return
				}
			}
		}
	}()
}

/*********************** CLIENT METHODS ***************************/

// TODO should all protocols receive connection up event ??
func (l *ProtoListener) DeliverEvent(event *NetworkEvent) {
	if event.DestProto == ALL_PROTO_ID {
		//log.Default().Println("GOING TO DELIVER AN EVENT TO ALL PROTOCOLS")
		for _, proto := range l.protocols {
			proto.queue <- event
		}
	} else {
		//protocol := l.protocols[event.SourceProto]
		protocol := l.protocols[event.DestProto]
		if protocol == nil {
			log.Println("RECEIVED EVENT FOR A NON EXISTENT PROTOCOL!")
		} else {
			//log.Println("GOING TO DELIVER AN EVENT TO THE PROTOCOL:", event.SourceProto)
			protocol.queue <- event
		}
	}
}
