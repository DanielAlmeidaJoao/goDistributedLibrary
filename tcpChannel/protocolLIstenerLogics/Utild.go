package protocolLIstenerLogics

import (
	list2 "container/list"
	"errors"
	"net"
)

var NOT_CONNECTED = errors.New("NOT_CONNECTED")
var PROTOCOL_EXIST_ALREADY = errors.New("PROTOCOL_EXIST_ALREADY")
var NO_PROTOCOLS_TO_RUN = errors.New("NO_PROTOCOLS_TO_RUN")
var ELEMENT_EXISTS_ALREADY = errors.New("ELEMENT_EXISTS_ALREADY")
var UNKNOWN_PROTOCOL = errors.New("UNKNOWN_PROTOCOL_ID")

type MessageHandlerID uint16 //protocol handler id
const NO_NETWORK_MESSAGE_HANDLER_ID MessageHandlerID = 0

type APP_PROTO_ID uint16

const ALL_PROTO_ID = 1

// todo
type NET_EVENT int

const (
	CONNECTION_UP NET_EVENT = iota
	CONNECTION_DOWN
	MESSAGE_RECEIVED
	ON_TIMER_TRIGGERED
)

type CONNECTION_TYPE int8

const (
	P2P CONNECTION_TYPE = iota
	CLIENT
	SERVER
)

type MSG_TYPE uint8

var UNEXPECTED_NETWORK_READ_DATA = errors.New("UNEXPECTED_NETWORK_READ_DATA")

const (
	LISTEN_ADDRESS_MSG MSG_TYPE = iota
	APP_MSG
)

type DataWrapper struct {
	From    net.Addr
	Data    []byte
	protoId int16
}

const MAX_BYTES_TO_READ = 2 * 1024

func Abort(err error) bool {
	if err == nil {
		return false
	} else {
		panic(err)
		return true
	}
}

type LocalCommunicationEvent struct {
	sourceProto APP_PROTO_ID
	destProto   APP_PROTO_ID
	data        interface{}
	funcHandler LocalProtoComHandlerFunc
}

func (l *LocalCommunicationEvent) ExecuteFunc() {
	l.funcHandler(l.sourceProto, l.destProto, l.data)
}

// type TimerEvent
type MsgHandlerFunc func(from net.Addr, data []byte, sourceProto APP_PROTO_ID, eventType NET_EVENT)

type TimerHandlerFunc func(sourceProto APP_PROTO_ID, data interface{})
type LocalProtoComHandlerFunc func(sourceProto, destProto APP_PROTO_ID, data interface{})

func NewLocalCommunicationEvent(sourceProto, destProto APP_PROTO_ID, data interface{}, funcHandler LocalProtoComHandlerFunc) *LocalCommunicationEvent {
	return &LocalCommunicationEvent{
		sourceProto: sourceProto,
		destProto:   destProto,
		data:        data,
		funcHandler: funcHandler,
	}
}

func GetFromList(value any, list list2.List) *list2.Element {
	if list.Len() == 0 {
		return nil
	}
	i := 0
	next := list.Front()
	for i < list.Len() {
		if next.Value == value {
			return next
		}
		next = next.Next()
		i++
	}
	return nil
}

func DeleteFromList(value any, list list2.List) any {
	ele := GetFromList(value, list)
	return list.Remove(ele)
}
