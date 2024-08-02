package common

import (
	"encoding/binary"
	"google.golang.org/protobuf/proto"
	"io"
)

/*
	each message sent over the network should implement this interface
	If a new message type needs to be added: first define it in a proto file, generate the go protobuf files and then implement the three methods
*/

type Serializable interface {
	Marshal(io.Writer) error
	Unmarshal(io.Reader) error
	New() Serializable
}

/*
	A struct that allocates a unique uint8 for each message type. When you define a new proto message type, add the message to here
*/

type MessageCode struct {
	ClientBatchRpc uint8
	StatusRPC      uint8
	PrepareRequest uint8
	PromiseReply   uint8
	ProposeRequest uint8
	AcceptReply    uint8
}

/*
	A static function which assigns a unique uint8 to each message type. Update this function when you define new message types
*/

func GetRPCCodes() MessageCode {
	return MessageCode{
		ClientBatchRpc: 1,
		StatusRPC:      2,
		PrepareRequest: 3,
		PromiseReply:   4,
		ProposeRequest: 5,
		AcceptReply:    6,
	}
}

func marshalMessage(wire io.Writer, m proto.Message) error {
	data, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	lengthWritten := len(data)
	var b [8]byte
	bs := b[:8]
	binary.LittleEndian.PutUint64(bs, uint64(lengthWritten))
	_, err = wire.Write(bs)
	if err != nil {
		return err
	}
	_, err = wire.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func unmarshalMessage(wire io.Reader, m proto.Message) error {
	var b [8]byte
	bs := b[:8]
	_, err := io.ReadFull(wire, bs)
	if err != nil {
		return err
	}
	numBytes := binary.LittleEndian.Uint64(bs)
	data := make([]byte, numBytes)
	length, err := io.ReadFull(wire, data)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(data[:length], m)
	if err != nil {
		return err
	}
	return nil
}

// ClientBatch wrapper

func (t *ClientBatch) Marshal(wire io.Writer) error {
	return marshalMessage(wire, t)
}

func (t *ClientBatch) Unmarshal(wire io.Reader) error {
	return unmarshalMessage(wire, t)
}

func (t *ClientBatch) New() Serializable {
	return new(ClientBatch)
}

// Status wrapper

func (t *Status) Marshal(wire io.Writer) error {
	return marshalMessage(wire, t)
}

func (t *Status) Unmarshal(wire io.Reader) error {
	return unmarshalMessage(wire, t)
}

func (t *Status) New() Serializable {
	return new(Status)
}

// PrepareRequest wrapper

func (t *PrepareRequest) Marshal(wire io.Writer) error {
	return marshalMessage(wire, t)
}

func (t *PrepareRequest) Unmarshal(wire io.Reader) error {
	return unmarshalMessage(wire, t)
}

func (t *PrepareRequest) New() Serializable {
	return new(PrepareRequest)
}

// PromiseReply wrapper

func (t *PromiseReply) Marshal(wire io.Writer) error {
	return marshalMessage(wire, t)
}

func (t *PromiseReply) Unmarshal(wire io.Reader) error {
	return unmarshalMessage(wire, t)
}

func (t *PromiseReply) New() Serializable {
	return new(PromiseReply)
}

// ProposeRequest wrapper

func (t *ProposeRequest) Marshal(wire io.Writer) error {
	return marshalMessage(wire, t)
}

func (t *ProposeRequest) Unmarshal(wire io.Reader) error {
	return unmarshalMessage(wire, t)
}

func (t *ProposeRequest) New() Serializable {
	return new(ProposeRequest)
}

// AcceptReply wrapper

func (t *AcceptReply) Marshal(wire io.Writer) error {
	return marshalMessage(wire, t)
}

func (t *AcceptReply) Unmarshal(wire io.Reader) error {
	return unmarshalMessage(wire, t)
}

func (t *AcceptReply) New() Serializable {
	return new(AcceptReply)
}
