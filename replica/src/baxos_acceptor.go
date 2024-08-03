package src

import "baxos/common"

/*
	Logic for prepare message, check if it is possible to promise for the specified instance
*/

func (rp *Replica) processPrepare(message *common.PrepareRequest) *common.PromiseReply {
	rp.createInstance(int(message.InstanceNumber))

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
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
	}
}

// logic for propose message

func (rp *Replica) processPropose(message *common.ProposeRequest) *common.AcceptReply {
	rp.createInstance(int(message.InstanceNumber))

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
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
		return &common.AcceptReply{
			InstanceNumber: message.InstanceNumber,
			Accept:         true,
			AcceptBallot:   int64(message.ProposeBallot),
			Sender:         int64(rp.name),
		}
	} else {
		return &common.AcceptReply{
			InstanceNumber: message.InstanceNumber,
			Accept:         false,
			Sender:         int64(rp.name),
		}
	}
}

/*
	handler for propose message
*/

func (rp *Replica) handlePropose(message *common.ProposeRequest) {

	// handle the prepare slot for the next instance
	promiseResponse := rp.processPrepare(message.PrepareRequestForFutureIndex)

	// handle the propose slot
	acceptReply := rp.processPropose(message)

	acceptReply.PromiseReplyForFutureIndex = promiseResponse // can be nil

	rp.sendMessage(int32(message.Sender), common.RPCPair{
		Code: rp.messageCodes.AcceptReply,
		Obj:  acceptReply,
	})

	// handle the decided slot
	if message.DecideInfo != nil {
		rp.createInstance(int(message.DecideInfo.InstanceNumber))
		rp.baxosConsensus.replicatedLog[message.DecideInfo.InstanceNumber].decided = true
		rp.baxosConsensus.replicatedLog[message.DecideInfo.InstanceNumber].decidedValue = *message.DecideInfo.DecidedValue
		rp.updateSMR()
	}
}
