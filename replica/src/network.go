package src

import (
	"baxos/common"
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

/*
	given an int32 id, connect to that node
	nodeType should be one of client and replica
*/

func (rp *Replica) ConnectToNode(id int32, address string, nodeType string) {
	if rp.debugOn {
		rp.debug("Connecting to "+strconv.Itoa(int(id))+" in address "+address, 3)
	}

	var b [4]byte
	bs := b[:4]

	for true {
		conn, err := net.Dial("tcp", address)
		if err == nil {
			if nodeType == "client" {
				rp.outgoingClientWriters[id] = bufio.NewWriter(conn)
				binary.LittleEndian.PutUint16(bs, uint16(rp.name))
				_, err := conn.Write(bs)
				if err != nil {
					panic("Error while connecting to client " + strconv.Itoa(int(id)))
				}
			} else if nodeType == "replica" {
				rp.outgoingReplicaWriters[id] = bufio.NewWriter(conn)
				binary.LittleEndian.PutUint16(bs, uint16(rp.name))
				_, err := conn.Write(bs)
				if err != nil {
					panic("Error while connecting to replica " + strconv.Itoa(int(id)))
				}
			} else {
				panic("Unknown node id")
			}
			if rp.debugOn {
				rp.debug("Established outgoing connection to "+strconv.Itoa(int(id)), 3)
			}
			break
		} else {
			if rp.debugOn {
				rp.debug("Error while connecting to "+strconv.Itoa(int(id))+" "+err.Error(), 3)
			}

		}
	}
}

/*
	Connect to all replicas on bootstrap
*/

func (rp *Replica) ConnectBootStrap() {

	for name, address := range rp.replicaAddrList {
		rp.ConnectToNode(name, address, "replica")
	}
}

/*
	listen to a given connection reader. Upon receiving any message, put it into the central incoming buffer
*/

func (rp *Replica) connectionListener(reader *bufio.Reader, id int32) {

	var msgType uint8
	var err error = nil

	for true {
		if msgType, err = reader.ReadByte(); err != nil {
			if rp.debugOn {
				rp.debug("Error while reading message code: connection broken from "+strconv.Itoa(int(id))+fmt.Sprintf(" %v", err.Error()), 3)
			}
			return
		}
		if rpair, present := rp.rpcTable[msgType]; present {
			obj := rpair.Obj.New()
			if err = obj.Unmarshal(reader); err != nil {
				if rp.debugOn {
					rp.debug("Error while unmarshalling from "+strconv.Itoa(int(id))+fmt.Sprintf(" %v", err.Error()), 3)
				}
				return
			}
			rp.incomingChan <- &common.RPCPair{
				Code: msgType,
				Obj:  obj,
			}
			if rp.debugOn {
				rp.debug("Pushed a message from "+strconv.Itoa(int(id)), 0)
			}
		} else {
			if rp.debugOn {
				rp.debug("Error received unknown message type from "+strconv.Itoa(int(id)), 3)
			}
			return
		}
	}
}

/*
	listen on the replica port for new connections from clients, and all replicas
*/

func (rp *Replica) WaitForConnections() {
	go func() {
		var b [4]byte
		bs := b[:4]
		Listener, err_ := net.Listen("tcp", rp.listenAddress)

		if err_ != nil {
			panic("Error while listening to incoming connections")
		}

		if rp.debugOn {
			rp.debug("Listening to incoming connections in "+rp.listenAddress, 3)
		}
		for true {
			conn, err := Listener.Accept()
			if err != nil {
				panic(err.Error() + fmt.Sprintf("%v", err.Error()))
			}
			if _, err := io.ReadFull(conn, bs); err != nil {
				panic(err.Error() + fmt.Sprintf("%v", err.Error()))
			}
			id := int32(binary.LittleEndian.Uint16(bs))
			if rp.debugOn {
				rp.debug("Received incoming connection from "+strconv.Itoa(int(id)), 3)
			}
			nodeType := rp.getNodeType(id)
			if nodeType == "client" {
				rp.incomingClientReaders[id] = bufio.NewReader(conn)
				go rp.connectionListener(rp.incomingClientReaders[id], id)
				if rp.debugOn {
					rp.debug("Started listening to client "+strconv.Itoa(int(id)), 3)
				}
				rp.ConnectToNode(id, rp.clientAddrList[id], "client")

			} else if nodeType == "replica" {
				rp.incomingReplicaReaders[id] = bufio.NewReader(conn)
				go rp.connectionListener(rp.incomingReplicaReaders[id], id)
				if rp.debugOn {
					rp.debug("Started listening to replica "+strconv.Itoa(int(id)), 3)
				}
			} else {
				panic("should not happen")
			}
		}
	}()
}

/*
	this is the main execution thread that listens to all the incoming messages
	It listens to incoming messages from the incomingChan, and invokes the appropriate handler depending on the message type
*/

func (rp *Replica) Run() {

	for true {
		select {
		case _ = <-rp.baxosConsensus.wakeupChan:
			rp.proposeAfterBackingOff()
			break
		case instance := <-rp.baxosConsensus.timeOutChan:
			rp.randomBackOff(instance)
			break
		case replicaMessage := <-rp.incomingChan:
			if rp.debugOn {
				rp.debug("Received replica message", 0)
			}
			switch replicaMessage.Code {

			case rp.messageCodes.StatusRPC:
				statusMessage := replicaMessage.Obj.(*common.Status)
				if rp.debugOn {
					rp.debug("Status message from "+fmt.Sprintf("%#v", statusMessage.Sender), 3)
				}
				rp.handleStatus(statusMessage)
				break

			case rp.messageCodes.ClientBatchRpc:
				clientBatch := replicaMessage.Obj.(*common.ClientBatch)
				if rp.debugOn {
					rp.debug("Client batch message from "+fmt.Sprintf("%#v", clientBatch.Sender), 0)
				}
				rp.handleClientBatch(clientBatch)
				break

			default:
				if rp.debugOn {
					rp.debug("Baxos consensus message ", 0)
				}
				rp.handleBaxosConsensus(replicaMessage.Obj, replicaMessage.Code)
				break

			}
			break
		}

	}
}

/*
	Write a message to the wire, first the message type is written and then the actual message
*/

func (rp *Replica) internalSendMessage(peer int32, rpcPair *common.RPCPair) {
	peerType := rp.getNodeType(peer)
	if peerType == "replica" {
		w := rp.outgoingReplicaWriters[peer]
		if w == nil {
			panic("replica not found" + strconv.Itoa(int(peer)))
		}
		rp.outgoingReplicaWriterMutexs[peer].Lock()
		err := w.WriteByte(rpcPair.Code)
		if err != nil {
			if rp.debugOn {
				rp.debug("Error writing message code byte:"+err.Error(), 0)
			}
			rp.outgoingReplicaWriterMutexs[peer].Unlock()
			return
		}
		err = rpcPair.Obj.Marshal(w)
		if err != nil {
			if rp.debugOn {
				rp.debug("Error while marshalling:"+err.Error(), 0)
			}
			rp.outgoingReplicaWriterMutexs[peer].Unlock()
			return
		}
		err = w.Flush()
		if err != nil {
			if rp.debugOn {
				rp.debug("Error while flushing:"+err.Error(), 0)
			}
			rp.outgoingReplicaWriterMutexs[peer].Unlock()
			return
		}
		rp.outgoingReplicaWriterMutexs[peer].Unlock()
		if rp.debugOn {
			rp.debug("Internal sent message to "+strconv.Itoa(int(peer)), 0)
		}
	} else if peerType == "client" {
		w := rp.outgoingClientWriters[peer]
		if w == nil {
			panic("client not found " + strconv.Itoa(int(peer)))
		}
		rp.outgoingClientWriterMutexs[peer].Lock()
		err := w.WriteByte(rpcPair.Code)
		if err != nil {
			if rp.debugOn {
				rp.debug("Error writing message code byte:"+err.Error(), 0)
			}
			rp.outgoingClientWriterMutexs[peer].Unlock()
			return
		}
		err = rpcPair.Obj.Marshal(w)
		if err != nil {
			if rp.debugOn {
				rp.debug("Error while marshalling:"+err.Error(), 0)
			}
			rp.outgoingClientWriterMutexs[peer].Unlock()
			return
		}
		err = w.Flush()
		if err != nil {
			if rp.debugOn {
				rp.debug("Error while flushing:"+err.Error(), 0)
			}
			rp.outgoingClientWriterMutexs[peer].Unlock()
			return
		}
		rp.outgoingClientWriterMutexs[peer].Unlock()
		if rp.debugOn {
			rp.debug("Internal sent message to "+strconv.Itoa(int(peer)), 0)
		}
	} else {
		panic("Unknown id from node name " + strconv.Itoa(int(peer)))
	}
}

/*
	send message
*/

func (rp *Replica) sendMessage(peer int32, rpcPair common.RPCPair) {
	rp.internalSendMessage(peer, &rpcPair)
}
