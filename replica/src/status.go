package src

import (
	"baxos/common"
	"fmt"
	"time"
)

/*
	Handler for status message
		1. Invoke bootstrap / start consensus or printlog depending on the operation type
		2. Send a response back to the sender
*/

func (rp *Replica) handleStatus(message *common.Status) {
	fmt.Print("Status  " + fmt.Sprintf("%v", message) + " \n")
	if message.Type == 1 {
		if rp.serverStarted == false {
			rp.serverStarted = true
			rp.ConnectBootStrap()
			time.Sleep(2 * time.Second)
		}
	} else if message.Type == 2 {
		if rp.logPrinted == false {
			rp.logPrinted = true
			rp.printBaxosLogConsensus() // this is for consensus testing purposes
		}
	} else if message.Type == 3 {
		if rp.consensusStarted == false {
			rp.consensusStarted = true
			rp.lastProposedTime = time.Now()
			if rp.debugOn {
				rp.debug("started baxos consensus", 0)
			}
		}
	}
	if rp.debugOn {
		rp.debug("Sending status reply ", 0)
	}
	statusMessage := common.Status{
		Type: message.Type,
		Note: message.Note,
	}

	rpcPair := common.RPCPair{
		Code: rp.messageCodes.StatusRPC,
		Obj:  &statusMessage,
	}

	rp.sendMessage(int32(message.Sender), rpcPair)
	if rp.debugOn {
		rp.debug("Sent status ", 0)
	}
}
