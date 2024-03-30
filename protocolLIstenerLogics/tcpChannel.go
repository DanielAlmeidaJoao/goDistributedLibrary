package protocolLIstenerLogics

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"syscall"
)

const HeaderSize = 11

type ChannelInterface interface {
	SendAppData(addressKey string, source, destProto APP_PROTO_ID, msg []byte) (int, error)
	SendAppData2(addressKey string, source, destProto APP_PROTO_ID, msg NetworkMessage, msgHandlerId MessageHandlerID) (int, error)

	CloseConnection(addressKey string)
	CloseConnections()
	OpenConnection(address string, port int, protoSource APP_PROTO_ID)
	IsConnected(address string) *CustomConnection
	Connections() []*CustomConnection
}

func NewTCPChannel(address string, port int, connectionType CONNECTION_TYPE) *tcpChannel {
	channel := &tcpChannel{
		connections:    make(map[string]*CustomConnection),
		address:        address,
		port:           port,
		connectionType: connectionType,
		started:        false,
	}
	return channel
}

type CustomConnection struct {
	conn             net.Conn
	remoteListenAddr net.Addr
	connectionKey    string
	isServer         bool
}

func (c *CustomConnection) SendData2(source, destProto APP_PROTO_ID, msg NetworkMessage, msgHandlerId MessageHandlerID) (int, error) {
	customWriter := NewCustomWriter3(binary.LittleEndian)
	//TODO ASSERT THAT THE BUFFER IS THE SAME AFTER BEING PASSED TO HEADER AND THEN TO THE SERIALIZEER
	writeHeaders(customWriter, source, destProto, APP_MSG, msgHandlerId)
	msg.SerializeData(customWriter)
	return auxWriteToNetworkChild(c, customWriter)
}

func (c *tcpChannel) CloseConnections() {
	for _, v := range c.connections {
		v.conn.Close()
	}
}
func (c *CustomConnection) SendData(source, destProto APP_PROTO_ID, msg []byte, msgHandlerId MessageHandlerID) (int, error) {
	customWriter := NewCustomWriter(HeaderSize+len(msg), binary.LittleEndian)
	//TODO ASSERT THAT THE BUFFER IS THE SAME AFTER BEING PASSED TO HEADER AND THEN TO THE SERIALIZEER
	writeHeaders(customWriter, source, destProto, APP_MSG, msgHandlerId)
	_, _ = customWriter.Write(msg)
	return auxWriteToNetworkChild(c, customWriter)
}

func (c *CustomConnection) RemoteAddress() net.Addr {
	return c.remoteListenAddr
}

func (c *CustomConnection) GetConnectionKey() string {
	return c.connectionKey
}

func (c *CustomConnection) Close() error {
	return c.conn.Close()
}

type tcpChannel struct {
	mutex          sync.Mutex
	connections    map[string]*CustomConnection //map[string]net.Conn  // connectionId, int_ip
	address        string
	port           int
	listener       net.Listener
	protoListener  ProtoListener
	connectionType CONNECTION_TYPE
	started        bool
}

func shutDown(c *tcpChannel) {
	c.listener.Close()
}
func (c *tcpChannel) SetProtoLister(p *ProtoListener) {
	c.protoListener = *p
}
func (c *tcpChannel) formatAddress() string {
	return fmt.Sprintf("%s:%d", c.address, c.port)
}
func (c *tcpChannel) Start() {
	if c.started {
		return
	}
	c.started = true
	listener, err := net.Listen("tcp", c.formatAddress())
	Abort(err)
	c.listener = listener
	go func() {
		for {
			conn, err := listener.Accept()
			Abort(err)
			customCon := &CustomConnection{
				conn: conn,
			}
			go c.readFromConnection(customCon, nil)
			//c.onConnected(conn, fmt.Sprintf("%s:%d", c.address, c.port), gobabelUtils.ALL_PROTO_ID, true)
		}
	}()
}

func (c *tcpChannel) toByteMSG(msgType MSG_TYPE, data []byte) {
	//TODO sum := make([]byte, 1+len(data))
	//append(sum, msgType...)
}
func (c *tcpChannel) onDisconnected(from string) {
	c.mutex.Lock()
	delete(c.connections, from)
	c.mutex.Unlock()
}
func (c *tcpChannel) auxAddCon(customCon *CustomConnection, conn net.Conn, listenAddress net.Addr, isServer bool) *CustomConnection {
	if customCon == nil {
		customCon = &CustomConnection{
			conn: conn,
		}
	}
	customCon.isServer = isServer //this entity is a server
	c.mutex.Lock()

	var addressKey string
	log.Println("LISTEN ADDRESS KEY AND REMOTE:", listenAddress.String(), conn.RemoteAddr().String(), conn.LocalAddr().String())
	if c.connections[listenAddress.String()] == nil {
		addressKey = listenAddress.String()
	} else if isServer {
		addressKey = conn.RemoteAddr().String()
	} else {
		log.Println(c.connections[listenAddress.String()].conn == conn)
		addressKey = conn.LocalAddr().String()
	}
	log.Println("ADDRESS KEY:", addressKey)

	c.connections[addressKey] = customCon
	c.mutex.Unlock()
	customCon.connectionKey = addressKey
	customCon.remoteListenAddr = listenAddress
	return customCon
}
func (c *tcpChannel) onConnected(conn net.Conn, listenAddress string, protoDest APP_PROTO_ID) {
	//send the local listen address to the remote server
	var customCon *CustomConnection
	customCon = c.auxAddCon(nil, conn, conn.RemoteAddr(), false)
	_, err := c.sendMessage(customCon.connectionKey, ALL_PROTO_ID, ALL_PROTO_ID, []byte(listenAddress), LISTEN_ADDRESS_MSG, NO_NETWORK_MESSAGE_HANDLER_ID)
	if err != nil {
		c.handleConnectionDown(customCon, &err)
		return
	}

	connectionUp := NewNetworkEvent(customCon, nil, protoDest, protoDest, CONNECTION_UP, NO_NETWORK_MESSAGE_HANDLER_ID)
	c.protoListener.DeliverEvent(connectionUp)
	//the protocols receive the connection when connection is up
	c.readFromConnection(customCon, conn.RemoteAddr())
	//empty string, the user only wants to be a client
	if strings.Compare("", listenAddress) == 0 {
		//TODO SEND LISTEN ADDRESS
		// LATER CHECK IF IT IS CLIENT | SERVER | P2P
	}
}

func (c *tcpChannel) CloseConnection(connectionId string) {
	conn := c.connections[connectionId]
	go func() {
		if conn != nil {
			c.mutex.Lock()
			delete(c.connections, connectionId)
			c.mutex.Unlock()
			conn.Close()
		}
	}()
}
func (c *tcpChannel) closeConnection2(conn net.Conn) {
	c.CloseConnection(conn.RemoteAddr().String())
}
func (c *tcpChannel) OpenConnection(address string, port int, protoSource APP_PROTO_ID) {
	go func() {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
		if err != nil {
			customCon := &CustomConnection{
				conn: conn, connectionKey: address, remoteListenAddr: nil,
			}
			c.handleConnectionDown(customCon, &err)
		} else {
			c.onConnected(conn, fmt.Sprintf("%s:%d", c.address, c.port), protoSource)
		}
	}()
}
func writeHeaders(buf *CustomWriter, source, destProto APP_PROTO_ID, msgType MSG_TYPE, msgHandlerId MessageHandlerID) {
	_ = binary.Write(buf, binary.LittleEndian, uint8(msgType))
	_ = binary.Write(buf, binary.LittleEndian, uint16(msgHandlerId))
	_ = binary.Write(buf, binary.LittleEndian, uint16(source))
	_ = binary.Write(buf, binary.LittleEndian, uint16(destProto))
	buf.SetOffSet(buf.OffSet() + 4)
}

//write order: msgType(uint8), sourceProto(uint16), destProto(uint16),dataLength(uint32),appData
func (t *tcpChannel) Connections() []*CustomConnection {
	cons := make([]*CustomConnection, len(t.connections))
	count := 0
	for _, v := range t.connections {
		cons[count] = v
		count++
	}
	return cons
}
func (c *tcpChannel) readFromConnection(customCon *CustomConnection, listenAddress net.Addr) {
	conn := customCon.conn
	for {
		var err error
		var msgCode uint8
		err = binary.Read(conn, binary.LittleEndian, &msgCode)
		if connectionDown(&err) || err != nil {
			c.handleConnectionDown(customCon, &err)
			return
		}
		var msgHandlerId uint16
		err = binary.Read(conn, binary.LittleEndian, &msgHandlerId)
		if connectionDown(&err) || err != nil {
			c.handleConnectionDown(customCon, &err)
			return
		}

		var sourceProto uint16
		err = binary.Read(conn, binary.LittleEndian, &sourceProto)
		if connectionDown(&err) {
			c.handleConnectionDown(customCon, &err)
			return
		}
		var destProto uint16
		err = binary.Read(conn, binary.LittleEndian, &destProto)
		if connectionDown(&err) {
			c.handleConnectionDown(customCon, &err)
			return
		}
		var dataLen uint32
		err = binary.Read(conn, binary.LittleEndian, &dataLen)
		if connectionDown(&err) {
			c.handleConnectionDown(customCon, &err)
			return
		}
		buffer := make([]byte, dataLen)
		var readData int
		readData, err = conn.Read(buffer)
		if connectionDown(&err) || err != nil {
			c.handleConnectionDown(customCon, &err)
			return
		}
		if readData != int(dataLen) {
			str := fmt.Sprintf("EXPECTED TO READ %d; BUT ONLY READ %d. CONN CLOSED", dataLen, readData)
			err = errors.New(str)
			fmt.Println(str)
			c.handleConnectionDown(customCon, &err)
			return
		}
		netEvent := MESSAGE_RECEIVED
		if listenAddress == nil {
			tcpAddr, err := net.ResolveTCPAddr("tcp", string(buffer))
			if err == nil {
				log.Println("THE CLIENT REMOTE ADDRESS IS: ", tcpAddr.String())
				c.auxAddCon(customCon, conn, tcpAddr, true)
				listenAddress = tcpAddr
				netEvent = CONNECTION_UP
			} else {
				_ = conn.Close()
				log.Fatal("FAILED TO PARSE RECEIVED ADDRESS OF THE SERVER:", err)
				return
			}
			buffer = nil
		}
		msgEvent := NewNetworkEvent(customCon, buffer, APP_PROTO_ID(sourceProto), APP_PROTO_ID(destProto), netEvent, MessageHandlerID(msgHandlerId))
		c.protoListener.DeliverEvent(msgEvent)
	}
}

func (c *tcpChannel) SendAppData(ipAddress string, source, destProto APP_PROTO_ID, msg []byte) (int, error) {
	return c.sendMessage(ipAddress, source, destProto, msg, APP_MSG, NO_NETWORK_MESSAGE_HANDLER_ID)
}
func (c *tcpChannel) SendAppData2(hostAddress string, source, destProto APP_PROTO_ID, msg NetworkMessage, msgHandlerId MessageHandlerID) (int, error) {
	customWriter := NewCustomWriter3(binary.LittleEndian)
	//TODO ASSERT THAT THE BUFFER IS THE SAME AFTER BEING PASSED TO HEADER AND THEN TO THE SERIALIZEER
	writeHeaders(customWriter, source, destProto, APP_MSG, msgHandlerId)
	msg.SerializeData(customWriter)
	return c.auxWriteToNetwork(hostAddress, customWriter)
}

func (c *tcpChannel) IsConnected(address string) *CustomConnection {
	return c.connections[address]
}

func auxWriteToNetworkChild(connection *CustomConnection, writer *CustomWriter) (int, error) {
	dataSize := uint32(writer.Len() - HeaderSize)
	aux := writer.OffSet()
	writer.SetOffSet(HeaderSize - 4)
	_, _ = writer.WriteUInt32(dataSize)
	writer.SetOffSet(aux)
	//println("SENT DATA IS ", base32.HexEncoding.EncodeToString(writer.Data()))
	return connection.conn.Write(writer.Data())
}

func (c *tcpChannel) auxWriteToNetwork(address string, writer *CustomWriter) (int, error) {
	customCon := c.connections[address]
	written := -1
	var err error
	if customCon != nil {
		return auxWriteToNetworkChild(customCon, writer)
		//written to logger service
	} else {
		err = errors.New(fmt.Sprintf("%s IS NOT CONNECTED", address))
	}
	return written, err
}

// TODO use a golang routine just to send messages
// write order: msgType(uint8), sourceProto(uint16), destProto(uint16),dataLength(uint32),appData
func (c *tcpChannel) sendMessage(hostAddress string, source, destProto APP_PROTO_ID, msg []byte, msgType MSG_TYPE, msgHandlerId MessageHandlerID) (int, error) {
	//I DONT LIKE IT:
	// using protoBuf to binary a struct with the data, msgType, ??
	//TODO ASSERT THAT THE BUFFER IS THE SAME AFTER BEING PASSED TO HEADER AND THEN TO THE SERIALIZEER
	customWriter := NewCustomWriter(HeaderSize+len(msg), binary.LittleEndian)
	writeHeaders(customWriter, source, destProto, msgType, msgHandlerId)
	customWriter.Write(msg)
	return c.auxWriteToNetwork(hostAddress, customWriter)
}
func (t *tcpChannel) handleConnectionDown(customCon *CustomConnection, err *error) {
	//_ = customCon.conn.Close()
	t.onDisconnected(customCon.connectionKey)
	msgEvent := NewNetworkEvent(customCon, nil, ALL_PROTO_ID, ALL_PROTO_ID, CONNECTION_DOWN, NO_NETWORK_MESSAGE_HANDLER_ID)
	t.protoListener.DeliverEvent(msgEvent)
}
func connectionDown(err *error) bool {
	return errors.Is(*err, io.EOF) || errors.Is(*err, syscall.EPIPE)
}

/******************* ******************* ******************* ******************* *******************/

type NetworkEvent struct {
	customConn       *CustomConnection
	Data             []byte
	SourceProto      APP_PROTO_ID
	DestProto        APP_PROTO_ID
	NET_EVENT        NET_EVENT
	MessageHandlerID MessageHandlerID
}

func NewNetworkEvent(customCon *CustomConnection, data []byte, sourceProto, destProto APP_PROTO_ID, eventType NET_EVENT, messageHandlerID MessageHandlerID) *NetworkEvent {
	return &NetworkEvent{
		customConn:       customCon,
		Data:             data,
		SourceProto:      sourceProto,
		DestProto:        destProto,
		NET_EVENT:        eventType,
		MessageHandlerID: messageHandlerID,
	}
}
