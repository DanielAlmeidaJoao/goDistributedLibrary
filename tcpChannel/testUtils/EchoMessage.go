package testUtils

import protoListener "github.com/DanielAlmeidaJoao/goDistributedLibrary/protocolLIstenerLogics"

/*** ***/
type EchoMessage struct {
	Data  string
	Count int32
}

type RandomMessage struct {
	Data  string
	Count float32
}

func (receiver *EchoMessage) SerializeData(writer *protoListener.CustomWriter) {
	//TODO implement me
	_, _ = writer.WriteUInt32(uint32(receiver.Count))
	_, _ = writer.WriteString(receiver.Data)
}

func (receiver *RandomMessage) SerializeData(writer *protoListener.CustomWriter) {
	//TODO implement me
	_, _ = writer.WriteFloat32(receiver.Count)
	_, _ = writer.WriteString(receiver.Data)
}

func DeserializeData(reader *protoListener.CustomReader) *EchoMessage {
	count, _ := reader.ReadUint32()
	str := string(reader.ReadTheRest())
	aux := &EchoMessage{
		Data:  str,
		Count: int32(count),
	}
	return aux
}

func DeserializeDataRandomMSG(reader *protoListener.CustomReader) *RandomMessage {
	count, _ := reader.ReadFloat32()
	str := string(reader.ReadTheRest())
	aux := &RandomMessage{
		Data:  str,
		Count: count,
	}
	return aux
}
