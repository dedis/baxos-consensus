package src

import (
	"baxos/common"
	"fmt"
)

/*
	Logic for prepare message, check if it is possible to promise for the specified instance
*/

func (rp *Replica) processPrepare(message *common.PrepareRequest) *common.PromiseReply {
	if message == nil {
		return nil
	}
	rp.createInstance(int(message.InstanceNumber))

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
		if rp.debugOn {
			rp.debug(fmt.Sprintf("ACCEPTOR: Instance %d already decided, hence sending a promise reply with the decided value", message.InstanceNumber), 1)
		}
		return &common.PromiseReply{
			InstanceNumber: message.InstanceNumber,
			Promise:        false,
			Decided:        true,
			DecidedValue:   &rp.baxosConsensus.replicatedLog[message.InstanceNumber].decidedValue,
			Sender:         int64(rp.name),
		}
	}

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].acceptor_bookkeeping.promisedBallot < message.PrepareBallot {
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].acceptor_bookkeeping.promisedBallot = message.PrepareBallot
		if rp.debugOn {
			rp.debug(fmt.Sprintf("ACCEPTOR: Prepare with ballot %d accepted, hence sending a promise reply for instance %d", message.PrepareBallot, message.InstanceNumber), 1)
		}
		return &common.PromiseReply{
			InstanceNumber:     message.InstanceNumber,
			Promise:            true,
			LastPromisedBallot: int64(message.PrepareBallot),
			LastAcceptedBallot: int64(rp.baxosConsensus.replicatedLog[message.InstanceNumber].acceptor_bookkeeping.acceptedBallot),
			LastAcceptedValue:  &rp.baxosConsensus.replicatedLog[message.InstanceNumber].acceptor_bookkeeping.acceptedValue,
			Sender:             int64(rp.name),
		}
	}
	return nil
}

/*
	Handler for prepare message
*/

func (rp *Replica) handlePrepare(message *common.PrepareRequest) {

	promiseReply := rp.processPrepare(message)

	if promiseReply != nil {
		rpcPair := common.RPCPair{
			Code: rp.messageCodes.PromiseReply,
			Obj:  promiseReply,
		}

		rp.sendMessage(int32(message.Sender), rpcPair)
		if rp.debugOn {
			rp.debug(fmt.Sprintf("ACCEPTOR: Sent a promise response to %d for instance %d", message.Sender, message.InstanceNumber), 0)
		}
	}
}

// logic for propose message

func (rp *Replica) processPropose(message *common.ProposeRequest) *common.AcceptReply {
	rp.createInstance(int(message.InstanceNumber))

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
		if rp.debugOn {
			rp.debug(fmt.Sprintf("ACCEPTOR: Instance %d already decided, hence sending a accept reply with the decided value", message.InstanceNumber), 1)
		}
		return &common.AcceptReply{
			InstanceNumber: message.InstanceNumber,
			Accept:         false,
			Decided:        true,
			DecidedValue:   &rp.baxosConsensus.replicatedLog[message.InstanceNumber].decidedValue,
			Sender:         int64(rp.name),
		}
	}

	if message.ProposeBallot >= rp.baxosConsensus.replicatedLog[message.InstanceNumber].acceptor_bookkeeping.promisedBallot {
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].acceptor_bookkeeping.acceptedBallot = message.ProposeBallot
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].acceptor_bookkeeping.acceptedValue = *message.ProposeValue
		if rp.debugOn {
			rp.debug(fmt.Sprintf("ACCEPTOR: Accepted propose with ballot %d for instance %d, hence sending a accept", message.ProposeBallot, message.InstanceNumber), 1)
		}
		return &common.AcceptReply{
			InstanceNumber: message.InstanceNumber,
			Accept:         true,
			AcceptBallot:   int64(message.ProposeBallot),
			Sender:         int64(rp.name),
		}
	} else {
		if rp.debugOn {
			rp.debug(fmt.Sprintf("ACCEPTOR: Propose rejected for instance %d with ballot %d", message.InstanceNumber, message.ProposeBallot), 1)
		}
		return nil
	}
}

/*
	handler for propose message
*/

func (rp *Replica) handlePropose(message *common.ProposeRequest) {

	// handle the propose slot
	acceptReply := rp.processPropose(message)

	if acceptReply != nil {
		rp.sendMessage(int32(message.Sender), common.RPCPair{
			Code: rp.messageCodes.AcceptReply,
			Obj:  acceptReply,
		})
		if rp.debugOn {
			rp.debug(fmt.Sprintf("ACCEPTOR: Sent accept message to %d for instance %d", message.Sender, message.InstanceNumber), 0)
		}
	}

	// handle the decided slot
	if message.DecideInfo != nil {
		rp.createInstance(int(message.DecideInfo.InstanceNumber))
		if !rp.baxosConsensus.replicatedLog[message.DecideInfo.InstanceNumber].decided {
			if rp.debugOn {
				rp.debug(fmt.Sprintf("ACCEPTOR: Instance %d decided using decided value in propose", message.DecideInfo.InstanceNumber), 2)
			}
			rp.baxosConsensus.replicatedLog[message.DecideInfo.InstanceNumber].decided = true
			rp.baxosConsensus.replicatedLog[message.DecideInfo.InstanceNumber].decidedValue = *message.DecideInfo.DecidedValue
			rp.updateSMR()
		}
	}
}
