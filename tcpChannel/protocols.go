package tcpChannel

type ProtoInterface interface {
	ProtocolUniqueId() APP_PROTO_ID
	OnStart(protocolAPI ProtocolAPI)
	OnMessageArrival(customCon *CustomConnection, source, destProto APP_PROTO_ID, msg []byte, protocolAPI ProtocolAPI)
	ConnectionUp(customCon *CustomConnection, protocolAPI ProtocolAPI)
	ConnectionDown(customCon *CustomConnection, protocolAPI ProtocolAPI)
}

type ProtocolManager struct {
	id            APP_PROTO_ID
	eventQueue    <-chan *NetworkEvent
	handlerID     MessageHandlerID
	eventHandlers map[MessageHandlerID]any
	channel       ChannelInterface
}

func (p *ProtocolManager) connectionUp(networkEvent *NetworkEvent) {

}
func (p *ProtocolManager) connectionDown(networkEvent *NetworkEvent) {

}
func (p *ProtocolManager) sendMessage(ipAddress string, source, destProto APP_PROTO_ID, msg []byte) {

}
func (p *ProtocolManager) deliveryEvent(networkEvent *NetworkEvent) {

}

func (p *ProtocolManager) nextEventId() MessageHandlerID {
	p.handlerID++
	return p.handlerID
}
func (p *ProtocolManager) start() {
	for {
		event := <-p.eventQueue
		networkEventHandler := p.eventHandlers[event.MessageHandlerID]
		if networkEventHandler != nil {
			// TODO networkEventHandler(event)
		}
	}
}
