package src

import (
	"bytes"
	"encoding/gob"
	"io"
	"strconv"
	"time"
)

/*
	RPC pair assigns a unique uint8 to each type of message defined in the proto files
*/

type RPCPair struct {
	Code uint8
	Obj  Serializable
}

/*
	Outgoing RPC assigns a rpc to its intended destination
*/

type OutgoingRPC struct {
	RpcPair *RPCPair
	Peer    int32
}

/*
	Returns the self ip:port
*/

func GetAddress(nodes []Instance, name int32) string {
	for i := 0; i < len(nodes); i++ {
		if nodes[i].Name == strconv.Itoa(int(name)) {
			return nodes[i].Address
		}
	}
	panic("should not happen")
}

/*
	Timer for triggering event upon timeout
*/

type TimerWithCancel struct {
	d time.Duration
	t *time.Timer
	c chan interface{}
	f func()
}

/*
	instantiate a new timer with cancel
*/

func NewTimerWithCancel(d time.Duration) *TimerWithCancel {
	t := &TimerWithCancel{}
	t.d = d
	t.c = make(chan interface{}, 5)
	return t
}

/*
	Start the timer
*/

func (t *TimerWithCancel) Start() {
	t.t = time.NewTimer(t.d)
	go func() {
		select {
		case <-t.t.C:
			t.f()
			return
		case <-t.c:
			return
		}
	}()
}

/*
	Set a function to call when timeout
*/

func (t *TimerWithCancel) SetTimeoutFuntion(f func()) {
	t.f = f
}

/*
Cancel timer
*/
func (t *TimerWithCancel) Cancel() {
	select {
	case t.c <- nil:
		// Success
		break
	default:
		//Unsuccessful
		break
	}

}

/*
	util function to get the size of a message
*/

func GetRealSizeOf(v interface{}) (int, error) {
	b := new(bytes.Buffer)
	if err := gob.NewEncoder(b).Encode(v); err != nil {
		return 0, err
	}
	return b.Len(), nil
}

/*
	each message sent over the network should implement this interface
	If a new message type needs to be added: first define it in a proto file, generate the go protobuf files using mage generate and then implement the three methods
*/

type Serializable interface {
	Marshal(io.Writer) error
	Unmarshal(io.Reader) error
	New() Serializable
}

/*
	A struct that allocates a unique uint8 for each message type. When you define a new proto message type, add the message to here
	raft messages are gRPC only, hence do not need a code
*/

type MessageCode struct {
	ClientBatchRpc uint8
	StatusRPC      uint8
	PaxosConsensus uint8
}

/*
	A static function which assigns a unique uint8 to each message type. Update this function when you define new message types
*/

func GetRPCCodes() MessageCode {
	return MessageCode{
		ClientBatchRpc: 1,
		StatusRPC:      2,
		PaxosConsensus: 3,
	}
}
