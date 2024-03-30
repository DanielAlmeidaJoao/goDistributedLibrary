package protocolLIstenerLogics

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type ConnectionState struct {
	address net.Addr
	state   NET_EVENT
	err     error
}
type ProtoListenerInterface interface {
	//AddProtocol(protocol ProtoInterface) error
	StartProtocol(protoWrapper ProtoInterface) error
	WaitForProtocolsToEnd(closeConnections bool)
	//StartProtocols() error
	RegisterNetworkMessageHandler(handlerId MessageHandlerID, funcHandler MESSAGE_HANDLER_TYPE) error
	RegisterTimeout(sourceProto APP_PROTO_ID, duration time.Duration, data interface{}, funcToExecute TimerHandlerFunc) int
	RegisterLocalCommunication(sourceProto, destProto APP_PROTO_ID, data interface{}, funcToExecute LocalProtoComHandlerFunc) error
	CancelTimer(timerId int) bool
	RegisterPeriodicTimeout(sourceProto APP_PROTO_ID, duration time.Duration, data interface{}, funcToExecute TimerHandlerFunc) int
	RemoveProtocol(id APP_PROTO_ID)
}
type protoWrapper struct {
	queue                   chan *NetworkEvent
	timeoutChannel          chan int
	localCommunicationQueue chan *LocalCommunicationEvent
	proto                   ProtoInterface
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
	mutex           sync.Mutex
	protocols       map[APP_PROTO_ID]*protoWrapper
	waitChannel     chan APP_PROTO_ID
	channel         ChannelInterface
	messageHandlers map[MessageHandlerID]MESSAGE_HANDLER_TYPE
	timerHandlers   map[int]*timerArgs
	timersId        int
	order           binary.ByteOrder
	ConnectionType  CONNECTION_TYPE
}

/*********************** CLIENT METHODS ***************************/
func (p *ProtoListener) RegisterNetworkMessageHandler(handlerId MessageHandlerID, funcHandler MESSAGE_HANDLER_TYPE) error {
	var err error
	p.mutex.Lock()
	if p.messageHandlers[handlerId] == nil {
		p.messageHandlers[handlerId] = funcHandler
		err = nil
	} else {
		err = ELEMENT_EXISTS_ALREADY
	}
	p.mutex.Unlock()
	return err
}

// address string, port int, connectionType gobabelUtils.CONNECTION_TYPE
func NewProtocolListener(address string, port int, connectionType CONNECTION_TYPE, order binary.ByteOrder) ProtoListenerInterface {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ch := NewTCPChannel(address, port, connectionType)
	protoL := &ProtoListener{
		protocols:       make(map[APP_PROTO_ID]*protoWrapper),
		timerHandlers:   make(map[int]*timerArgs),
		waitChannel:     make(chan APP_PROTO_ID),
		channel:         ch,
		order:           order,
		messageHandlers: make(map[MessageHandlerID]MESSAGE_HANDLER_TYPE),
		ConnectionType:  connectionType,
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
func (l *ProtoListener) RegisterTimeout(sourceProto APP_PROTO_ID, duration time.Duration, data interface{}, funcToExecute TimerHandlerFunc) int {
	aux := l.timersId
	t := time.AfterFunc(duration, func() {
		l.protocols[sourceProto].timeoutChannel <- aux
	})
	l.timerHandlers[l.timersId] = &timerArgs{
		protoId:     sourceProto,
		data:        data,
		funcHandler: funcToExecute,
		timer:       t,
	}
	l.timersId++
	return aux
}

// I DONT LIKE IT BECAUSE FOR EACH TIMER I NEED TO HAVE A GOROUTINE
func (l *ProtoListener) RegisterPeriodicTimeout(sourceProto APP_PROTO_ID, duration time.Duration, data interface{}, funcToExecute TimerHandlerFunc) int {
	ticker := time.NewTicker(duration)
	aux := l.timersId
	go func() {
		for range ticker.C {
			l.protocols[sourceProto].timeoutChannel <- aux
		}
	}()
	l.timerHandlers[l.timersId] = &timerArgs{
		protoId:       sourceProto,
		data:          data,
		funcHandler:   funcToExecute,
		timer:         nil,
		periodicTimer: ticker,
	}
	l.timersId++
	return aux
}

func (l *ProtoListener) RegisterLocalCommunication(sourceProto, destProto APP_PROTO_ID, data interface{}, funcToExecute LocalProtoComHandlerFunc) error {
	proto := l.protocols[destProto]
	if proto == nil {
		return UNKNOWN_PROTOCOL
	}
	proto.localCommunicationQueue <- NewLocalCommunicationEvent(sourceProto, destProto, data, funcToExecute)
	return nil
}

func (l *ProtoListener) CancelTimer(timerId int) bool {
	args := l.timerHandlers[timerId]
	if args != nil {
		if args.timer == nil {
			args.periodicTimer.Stop()
		} else {
			args.timer.Stop()
		}
	}
	return true
}

func (l *ProtoListener) AddProtocol(protocol ProtoInterface) error {
	//TODO make the constants dynamic
	if l.protocols[(protocol).ProtocolUniqueId()] != nil {
		return PROTOCOL_EXIST_ALREADY
	}
	l.protocols[(protocol).ProtocolUniqueId()] = &protoWrapper{
		queue:                   make(chan *NetworkEvent, 50),
		proto:                   protocol,
		timeoutChannel:          make(chan int, 100),
		localCommunicationQueue: make(chan *LocalCommunicationEvent, 20),
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
func (l *ProtoListener) StartProtocol(protoWrapper ProtoInterface) error {
	err := l.AddProtocol(protoWrapper)
	if len(l.protocols) == 1 {
		if ch, ok := l.channel.(*tcpChannel); ok {
			ch.Start()
		}
	}
	if err == nil {
		l.auxRunProtocol(l.protocols[protoWrapper.ProtocolUniqueId()])
	}
	return err
}

func (l *ProtoListener) auxRunProtocol(protoWrapper *protoWrapper) {
	go func() {
		proto := protoWrapper.proto
		proto.OnStart(l.channel)
		log.Printf("PROTOCOL <%d> STARTED LISTENING TO EVENTS...\n", proto.ProtocolUniqueId())
		for {
			select {
			case networkEvent, ok := <-protoWrapper.queue:
				if ok {
					log.Printf("PROTOCOL <%d> RECEIVED AN EVENT. EVENT TYPE <%d>\n", proto.ProtocolUniqueId(), networkEvent.NET_EVENT)
					switch networkEvent.NET_EVENT {
					case CONNECTION_UP:
						proto.ConnectionUp(networkEvent.customConn, l.channel)
					case CONNECTION_DOWN:
						proto.ConnectionDown(networkEvent.customConn, l.channel)
					case MESSAGE_RECEIVED:
						if NO_NETWORK_MESSAGE_HANDLER_ID == networkEvent.MessageHandlerID {
							proto.OnMessageArrival(networkEvent.customConn, networkEvent.SourceProto, networkEvent.DestProto, networkEvent.Data, l.channel)
						} else {
							messageHandler := l.messageHandlers[networkEvent.MessageHandlerID]
							messageHandler(networkEvent.customConn, networkEvent.SourceProto, NewCustomReader(networkEvent.Data, l.order))
						}
					default:
						(l.channel).CloseConnection(networkEvent.customConn.connectionKey)
						log.Fatal(fmt.Sprintf("RECEIVED AN EVENT NOT PART OF THE PROTOCOL. CONNECTION CLOSED %s", networkEvent.customConn.remoteListenAddr.String()))
					}
				} else {
					return
				}
			case timerId, ok := <-protoWrapper.timeoutChannel:
				if ok {
					args := l.timerHandlers[timerId]
					if args != nil {
						if args.timer != nil {
							// it is a timeout, otherwise it is a periodic timer
							delete(l.timerHandlers, timerId)
						}
						args.funcHandler(args.protoId, args.data)
					}
				} else {
					return
				}
			case localEventCom, ok := <-protoWrapper.localCommunicationQueue:
				if ok {
					localEventCom.ExecuteFunc()
				} else {
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
		log.Default().Println("GOING TO DELIVER AN EVENT TO ALL PROTOCOLS")
		for _, proto := range l.protocols {
			proto.queue <- event
		}
	} else {
		protocol := l.protocols[event.SourceProto]
		if protocol == nil {
			log.Println("RECEIVED EVENT FOR A NON EXISTENT PROTOCOL!")
		} else {
			log.Default().Println("GOING TO DELIVER AN EVENT TO THE PROTOCOL:", event.SourceProto)
			protocol.queue <- event
		}
	}
}
