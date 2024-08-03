package src

import (
	"baxos/common"
	"math/rand"
	"strconv"
	"time"
)

/*
	Sets a timer, which once timeout will send an internal notification for setting the backoff timer
*/

func (rp *Replica) setTimer(instance int64) {

	rp.baxosConsensus.timer = common.NewTimerWithCancel(time.Duration(rp.baxosConsensus.roundTripTime+int64(rand.Intn(int(rp.baxosConsensus.roundTripTime)+int(rp.name)))) * time.Microsecond)

	rp.baxosConsensus.timer.SetTimeoutFuntion(func() {
		rp.baxosConsensus.timeOutChan <- instance
		if rp.debugOn {
			rp.debug("Timed on when proposing for instance "+strconv.Itoa(int(instance)), 0)
		}

	})
	rp.baxosConsensus.timer.Start()
}

// this is triggered when the proposer timer timeout after waiting for promise / accept messages

func (rp *Replica) randomBackOff(instance int64) {
	rp.baxosConsensus.isBackingOff = true
	rp.baxosConsensus.isProposing = true
	rp.baxosConsensus.timer = nil
	rp.baxosConsensus.retries++

	// reset the proposer bookkeeping
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.preparedBallot = -1
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.numSuccessfulPromises = 0
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.highestSeenAcceptedBallot = -1
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.highestSeenAcceptedValue = common.ReplicaBatch{}
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.proposedValue = common.ReplicaBatch{}
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.numSuccessfulAccepts = 0

	// set the backing off timer
	rp.baxosConsensus.wakeupTimer = common.NewTimerWithCancel(rp.calculateBackOffTime() * time.Microsecond)

	rp.baxosConsensus.wakeupTimer.SetTimeoutFuntion(func() {
		rp.baxosConsensus.wakeupChan <- true
		if rp.debugOn {
			rp.debug("Finished backing off ", 0)
		}
	})
}

// this is triggered after the backoff timer timeouts and the proposer is ready to propose again

func (rp *Replica) proposeAfterBackingOff() {
	rp.baxosConsensus.isBackingOff = false
	rp.sendPrepare()
}

/*
	send a prepare message to lsatCommittedIndex + 1
*/

func (rp *Replica) sendPrepare() {
	nextFreeInstance := rp.baxosConsensus.lastCommittedLogIndex + 1
	for rp.baxosConsensus.replicatedLog[nextFreeInstance].decided {
		nextFreeInstance++
	}

	rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.preparedBallot = -1
	rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.numSuccessfulPromises = 0
	rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.highestSeenAcceptedBallot = -1
	rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.highestSeenAcceptedValue = common.ReplicaBatch{}
	rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.proposedValue = common.ReplicaBatch{}
	rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.numSuccessfulAccepts = 0

	rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.preparedBallot =
		rp.baxosConsensus.replicatedLog[nextFreeInstance].acceptor_bookkeeping.promisedBallot + rp.name + 1

	for k, _ := range rp.replicaAddrList {
		prepareMessage := common.PrepareRequest{
			InstanceNumber: nextFreeInstance,
			PrepareBallot:  rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.preparedBallot,
			Sender:         int64(rp.name),
		}
		rp.sendMessage(k, common.RPCPair{Code: rp.messageCodes.PrepareRequest, Obj: &prepareMessage})
	}

	if rp.baxosConsensus.timer != nil {
		rp.baxosConsensus.timer.Cancel()
	}
	rp.setTimer(int64(nextFreeInstance))

}

/*
	Handler for promise message. Promise
*/

func (rp *Replica) handlePromise(message *common.PromiseReply) {

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
		return
	}

	if message.Decided {
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided = true
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].decidedValue = *message.DecidedValue
		return
	}

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulPromises >= rp.baxosConsensus.quorumSize {
		return
	}

	if message.Promise && int32(message.LastPromisedBallot) == rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.preparedBallot {
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulPromises++
		if int32(message.LastAcceptedBallot) > rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.highestSeenAcceptedBallot {
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.highestSeenAcceptedBallot = int32(message.LastAcceptedBallot)
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.highestSeenAcceptedValue = *message.LastAcceptedValue
		}
		if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulPromises == rp.baxosConsensus.quorumSize {
			rp.sendPropose(message.InstanceNumber)
		}

	}
}

// invoked upon receiving a client batch

func (rp *Replica) tryPropose() {

	if rp.baxosConsensus.isProposing || rp.baxosConsensus.isBackingOff {
		return
	}

	nextFreeInstance := rp.baxosConsensus.lastCommittedLogIndex + 1
	for rp.baxosConsensus.replicatedLog[nextFreeInstance].decided {
		nextFreeInstance++
	}

	if rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.numSuccessfulPromises >= rp.baxosConsensus.quorumSize {
		rp.sendPropose(nextFreeInstance)
	} else {
		rp.sendPrepare()
		rp.baxosConsensus.isProposing = true
	}

}

/*
	leader invokes this function to replicate a new instance
	can be invoked when receiving a new client batch / when the promise set is successful
	should propose only if there is no outstanding proposal
*/

func (rp *Replica) sendPropose(instance int32) {
	// set the prepare request for the next instance
	var prepareRequest *common.PrepareRequest

	if rp.baxosConsensus.replicatedLog[instance+1].decided || rp.baxosConsensus.replicatedLog[instance+1].proposer_bookkeeping.numSuccessfulPromises >= rp.baxosConsensus.quorumSize {
		// nothing to prepare
		prepareRequest = nil
	} else {
		rp.baxosConsensus.replicatedLog[instance+1].proposer_bookkeeping.preparedBallot = rp.baxosConsensus.replicatedLog[instance+1].acceptor_bookkeeping.promisedBallot + rp.name + 1
		prepareRequest = &common.PrepareRequest{
			InstanceNumber: instance + 1,
			PrepareBallot:  rp.baxosConsensus.replicatedLog[instance+1].proposer_bookkeeping.preparedBallot,
			Sender:         int64(rp.name),
		}
	}

	// set the decided info

	if !rp.baxosConsensus.replicatedLog[instance-1].decided {
		panic("error, previous index not deided")
	}

	decideInfo := common.DecideInfo{
		InstanceNumber: instance - 1,
		DecidedValue:   &rp.baxosConsensus.replicatedLog[instance-1].decidedValue,
	}

	// propose message
	var proposeValue *common.ReplicaBatch

	if rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.highestSeenAcceptedBallot != -1 {
		proposeValue = &rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.highestSeenAcceptedValue
	} else {
		var requests []*common.ClientBatch

		if len(rp.incomingRequests) < rp.replicaBatchSize {
			for i := 0; i < len(rp.incomingRequests); i++ {
				requests = append(requests, rp.incomingRequests[i])
			}
			rp.incomingRequests = []*common.ClientBatch{}
		} else {
			for i := 0; i < rp.replicaBatchSize; i++ {
				requests = append(requests, rp.incomingRequests[i])
			}
			rp.incomingRequests = rp.incomingRequests[rp.replicaBatchSize:]
		}

		proposeValue = &common.ReplicaBatch{
			Requests: requests,
		}
	}

	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.proposedValue = *proposeValue

	for k, _ := range rp.replicaAddrList {
		proposeRequest := common.ProposeRequest{
			InstanceNumber:               instance,
			ProposeBallot:                rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.preparedBallot,
			ProposeValue:                 proposeValue,
			PrepareRequestForFutureIndex: prepareRequest,
			Sender:                       int64(rp.name),
			DecideInfo:                   &decideInfo,
		}
		rp.sendMessage(k, common.RPCPair{Code: rp.messageCodes.ProposeRequest, Obj: &proposeRequest})
	}

	if rp.baxosConsensus.timer != nil {
		rp.baxosConsensus.timer.Cancel()
	}
	rp.setTimer(int64(instance))
}

/*
	handler for accept messages
*/

func (rp *Replica) handleAccept(message *common.AcceptReply) {
	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
		return
	}

	if message.Decided {
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided = true
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].decidedValue = *message.DecidedValue
		return
	}

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulAccepts >= rp.baxosConsensus.quorumSize {
		return
	}

	if message.Accept && int32(message.AcceptBallot) == rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.preparedBallot {
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulAccepts++
		if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulAccepts == rp.baxosConsensus.quorumSize {
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided = true
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].decidedValue = rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.proposedValue
			rp.updateSMR()
			rp.baxosConsensus.isProposing = false
			rp.baxosConsensus.retries--
			if rp.baxosConsensus.retries < 0 {
				rp.baxosConsensus.retries = 0
			}
			if rp.baxosConsensus.timer != nil {
				rp.baxosConsensus.timer.Cancel()
			}
			rp.baxosConsensus.timer = nil
		}

	}
}
