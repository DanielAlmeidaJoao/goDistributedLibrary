package main

import (
	"encoding/binary"
	"fmt"
	protoListener "github.com/DanielAlmeidaJoao/goDistributedLibrary/tcpChannel"
	testUtils "github.com/DanielAlmeidaJoao/goDistributedLibrary/testUtils"
)

/**
type ProtoInterface interface {
	ProtocolUniqueId() gobabelUtils.APP_PROTO_ID
	OnStart(channelInterface *channel.ChannelInterface)
	OnMessageArrival(from *net.Addr, source, destProto gobabelUtils.APP_PROTO_ID, msg []byte, channelInterface *channel.ChannelInterface)
	ConnectionUp(from *net.Addr, channelInterface *channel.ChannelInterface)
	ConnectionDown(from *net.Addr, channelInterface *channel.ChannelInterface)
}
*/

/** ****/

// type MESSAGE_HANDLER_TYPE func(from string, protoSource APP_PROTO_ID, data tcpChannel.NetworkMessage)
func main() {
	//go build ./...
	fmt.Println("SERVER STARTED")
	pp := protoListener.NewProtocolListener("localhost", 3000, protoListener.SERVER, binary.LittleEndian)
	proto := testUtils.NewEchoProto(pp)
	proto2 := testUtils.NewEchoProto2(pp)

	fmt.Println("SERVER STARTED2")

	err := pp.StartProtocol(proto, 200, 100, 20)
	fmt.Println(err)
	err = pp.StartProtocol(proto2, 200, 100, 20)
	fmt.Println(err)

	fmt.Println("SERVER STARTED3")
	//time.Sleep(time.Second * 3)
	//proto.ChannelInterface.OpenConnection("localhost", 3002, 45)
	//time.Sleep(time.Second * 5)
	//(handlerId gobabelUtils.MessageHandlerID, funcHandler MESSAGE_HANDLER_TYPE, deserializer MESSAGE_DESERIALIZER_TYPE)
	err1 := pp.RegisterNetworkMessageHandler(protoListener.MessageHandlerID(2), proto.HandleMessage)
	err2 := pp.RegisterNetworkMessageHandler(protoListener.MessageHandlerID(3), proto.HandleMessage2) //registar no server

	//err2 = pp.SendLocalEvent(proto.ProtocolUniqueId(), proto2.ProtocolUniqueId(), 656, proto2.HandleLocalCommunication) //registar no server
	//fmt.Println(err2)
	fmt.Println("ERROR REGISTERING MSG HANDLERS:", err1, err2)
	fmt.Println("SERVER STARTED4")
	//pp.StartProtocols()

	fmt.Println("SERVER STARTED5")

	pp.WaitForProtocolsToEnd(false)
}
