package main

import (
	"encoding/binary"
	"fmt"
	protoListener "github.com/DanielAlmeidaJoao/goDistributedLibrary/protocolLIstenerLogics"
	testUtils "github.com/DanielAlmeidaJoao/goDistributedLibrary/testUtils"
	"log"
	"time"
)

//go build ./...
/**
type ProtoInterface interface {
	ProtocolUniqueId() gobabelUtils.APP_PROTO_ID
	OnStart(channelInterface *channel.ChannelInterface)
	OnMessageArrival(from *net.Addr, source, destProto gobabelUtils.APP_PROTO_ID, msg []byte, channelInterface *channel.ChannelInterface)
	ConnectionUp(from *net.Addr, channelInterface *channel.ChannelInterface)
	ConnectionDown(from *net.Addr, channelInterface *channel.ChannelInterface)
}
*/

func TimerFunc1(proto protoListener.APP_PROTO_ID, message interface{}) {
	log.Println("TIMER TRIGERED!!!! !!!! !!!!")
	msg, ok := message.(*testUtils.EchoMessage)
	if ok {
		log.Println("message is: %s %d", msg.Data, msg.Count)
	} else {
		log.Println("DATA IS WRONG TYPE")
	}
}
func PeriodicTimerHandler(proto protoListener.APP_PROTO_ID, message interface{}) {
	localTime := time.Now()
	protocol, ok := message.(*testUtils.ProtoEcho)
	if ok {
		protocol.Counter++
		msg := testUtils.EchoMessage{
			Data:  localTime.String(),
			Count: int32(protocol.Counter),
		}
		protocol.MessagesSent[protocol.Counter] = &msg
		for _, value := range protocol.ChannelInterface.Connections() {
			result, er := value.SendData2(45, 45, &msg, 2)
			log.Println("MSG SENT AFTER TIMER: ", result, er)
		}

	} else {
		log.Fatal("FAILED TO CAST OBJECT")
	}
}
func main() {
	// go build ./...
	fmt.Println("SERVER STARTED")

	protocolsManager := protoListener.NewProtocolListener("localhost", 3002, protoListener.SERVER, binary.LittleEndian)
	proto := testUtils.NewEchoProto(protocolsManager)
	//protocolsManager.AddProtocol(proto)
	//err1 := protocolsManager.RegisterNetworkMessageHandler(gobabelUtils.MessageHandlerID(2), proto.HandleMessage)
	err2 := protocolsManager.RegisterNetworkMessageHandler(protoListener.MessageHandlerID(2), proto.ClientHandleMessage) //registar no server
	if err2 != nil {
		log.Println("ERROR REGISTERING MSG HANDLERS:", err2)
	}

	//err := protocolsManager.StartProtocols()
	err := protocolsManager.StartProtocol(proto)
	if err != nil {
		log.Fatal(err)
		panic(1)
	}

	time.Sleep(time.Second * 3)
	proto.ChannelInterface.OpenConnection("localhost", 3000, 45)
	proto.ChannelInterface.OpenConnection("localhost", 3000, 45)
	time.Sleep(time.Second * 5)

	_ = testUtils.EchoMessage{
		Data:  "SAY WHAAAAT ??",
		Count: 134,
	}
	//SendAppData2(hostAddress string, source, destProto gobabelUtils.APP_PROTO_ID, msg NetworkMessage, msgHandlerId gobabelUtils.MessageHandlerID) (int, error)
	//result, er := proto.ChannelInterface.SendAppData2(proto.ServerAddr.GetConnectionKey(), 45, 45, &msg, 2)
	//result, er = proto.ServerAddr.SendData2(45, 45, &msg, 2)

	//fmt.Println("RESULT IS: ", result, er)

	//protocolsManager.RegisterTimeout(proto.ProtocolUniqueId(), time.Second*5, &msg, TimerFunc1)
	protocolsManager.RegisterPeriodicTimeout(proto.ProtocolUniqueId(), time.Millisecond*200, proto, PeriodicTimerHandler)

	protocolsManager.WaitForProtocolsToEnd(false)
}
