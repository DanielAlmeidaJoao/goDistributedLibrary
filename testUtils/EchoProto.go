package testUtils

import (
	"fmt"
	protoListener "github.com/DanielAlmeidaJoao/goDistributedLibrary/tcpChannel"
	"log"
)

type ProtoEcho struct {
	id               protoListener.APP_PROTO_ID
	ChannelInterface protoListener.ChannelInterface
	ServerAddr       *protoListener.CustomConnection
	Counter          int
	listener         protoListener.ProtoListenerInterface
	MessagesSent     map[int]*EchoMessage
}

/*

	ProtocolUniqueId() gobabelUtils.APP_PROTO_ID
	OnStart(channelInterface ChannelInterface)
	OnMessageArrival(customCon *customConnection, source, destProto gobabelUtils.APP_PROTO_ID, msg []byte, channelInterface ChannelInterface)
	ConnectionUp(customCon *customConnection, channelInterface ChannelInterface)
	ConnectionDown(customCon *customConnection, channelInterface ChannelInterface)

*/

func (p *ProtoEcho) ProtocolUniqueId() protoListener.APP_PROTO_ID {
	return p.id
}

func (p *ProtoEcho) OnStart(channelInterface protoListener.ChannelInterface) {
	p.ChannelInterface = channelInterface
	log.Println("PROTOCOL STARTED: ", p.ProtocolUniqueId())
}
func (p *ProtoEcho) OnMessageArrival(customCon *protoListener.CustomConnection, source, destProto protoListener.APP_PROTO_ID, msg []byte, channelInterface protoListener.ChannelInterface) {
	str := string(msg)
	fmt.Println("------------ RECEIVED MESSAGE FROM: -----------------", customCon.GetConnectionKey(), str)
}
func (p *ProtoEcho) ConnectionUp(customCon *protoListener.CustomConnection, channelInterface protoListener.ChannelInterface) {
	log.Printf("CONNECTION IS UP. FROM <%s>\n", customCon.GetConnectionKey())
	p.ServerAddr = customCon
}
func (p *ProtoEcho) ConnectionDown(customCon *protoListener.CustomConnection, channelInterface protoListener.ChannelInterface) {
	log.Printf("CONNECTION IS DOW. TO/FROM <%s>\n", customCon.GetConnectionKey())
}
func (p *ProtoEcho) HandleMessage(customConn *protoListener.CustomConnection, protoSource protoListener.APP_PROTO_ID, data *protoListener.CustomReader) {
	msg := DeserializeData(data)
	log.Println("RECEIVED MESSAGE IS:", msg.Data, msg.Count)
	pair := &protoListener.CustomPair[string, *EchoMessage]{
		First:  customConn.GetConnectionKey(),
		Second: msg,
	}
	err2 := p.listener.RegisterLocalCommunication(p.ProtocolUniqueId(), 50, pair, p.Proto2GoingToReply) //registar no server
	if err2 != nil {
		log.Println("ERROR SENDING DATA TO ANOTHER PROTO", err2)
	}
}
func sameMsgs(m1, m2 *EchoMessage) bool {
	return m1.Count == m2.Count && m1.Data == m2.Data
}
func (p *ProtoEcho) ClientHandleMessage(customConn *protoListener.CustomConnection, protoSource protoListener.APP_PROTO_ID, data *protoListener.CustomReader) {
	msg := DeserializeData(data)
	m := p.MessagesSent[int(msg.Count)]
	log.Println("CLIENT RECEIVED THE SAME MESSAGE ? ", sameMsgs(m, msg), m.Count)
}
func (p *ProtoEcho) HandleMessage2(customConn *protoListener.CustomConnection, protoSource protoListener.APP_PROTO_ID, data *protoListener.CustomReader) {
	fmt.Println("RECEIVED A RANDOM MESSAGE FROM ", customConn.GetConnectionKey())
	msg := DeserializeDataRandomMSG(data)
	println("RECEIVED RANDOM MESSAGE IS:", msg.Data, msg.Count)
	fmt.Println("RANDOM NUMBER IS :", msg.Count)
}

func (p *ProtoEcho) HandleLocalCommunication(sourceProto, destProto protoListener.APP_PROTO_ID, data interface{}) {
	log.Println(sourceProto, destProto, data)
}

func (p *ProtoEcho) Proto2GoingToReply(sourceProto, destProto protoListener.APP_PROTO_ID, data interface{}) {
	log.Println(p.id, "RECEVEID DATA FROM:", sourceProto, destProto, data)

	pair, ok := data.(*protoListener.CustomPair[string, *EchoMessage])

	if ok {
		result, er := p.ChannelInterface.SendAppData2(pair.First, 45, 45, pair.Second, 2)
		log.Println("SERVER REPLIED WITH:", result, er)
	}
}

func test() {
	msg := EchoMessage{}
	var netMsg protoListener.NetworkMessage = &msg
	msg2 := netMsg
	println(msg2)
}
func (p *ProtoEcho) SendMessage(address string, data string) (int, error) {
	return p.ChannelInterface.SendAppData(address, p.ProtocolUniqueId(), p.ProtocolUniqueId(), []byte(data))
}

func NewEchoProto(manager protoListener.ProtoListenerInterface) *ProtoEcho {
	return &ProtoEcho{
		id:           protoListener.APP_PROTO_ID(45),
		listener:     manager,
		MessagesSent: make(map[int]*EchoMessage),
	}
}
func NewEchoProto2(manager protoListener.ProtoListenerInterface) *ProtoEcho {
	return &ProtoEcho{
		id:           protoListener.APP_PROTO_ID(50),
		listener:     manager,
		MessagesSent: make(map[int]*EchoMessage),
	}
}
