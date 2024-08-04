package src

import (
	"baxos/common"
	"math/rand"
	"strconv"
	"time"
)

/*
	sets a timer, which once timeout will send an internal notification for setting the backoff timer
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
	if rp.debugOn {
		rp.debug("Proposing after backing off", 0)
	}
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

	rp.createInstance(int(nextFreeInstance))

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
	if rp.debugOn {
		rp.debug("Sent prepare for instance "+strconv.Itoa(int(nextFreeInstance))+" with prepared ballot "+strconv.Itoa(int(rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.preparedBallot)), 0)
	}

	if rp.baxosConsensus.timer != nil {
		rp.baxosConsensus.timer.Cancel()
	}
	rp.setTimer(int64(nextFreeInstance))

}

/*
	Handler for promise message
*/

func (rp *Replica) handlePromise(message *common.PromiseReply) {

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {

		if rp.debugOn {
			rp.debug("Instance "+strconv.Itoa(int(message.InstanceNumber))+" already decided, hence ignoring the promise", 0)
		}
		return
	}

	if message.Decided {
		if !rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided = true
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].decidedValue = *message.DecidedValue
			if rp.debugOn {
				rp.debug("Instance "+strconv.Itoa(int(message.InstanceNumber))+" decided using promise response, hence setting the decided value", 0)
			}
			rp.updateSMR()
			return
		}
	}

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulPromises >= rp.baxosConsensus.quorumSize {
		if rp.debugOn {
			rp.debug("Instance "+strconv.Itoa(int(message.InstanceNumber))+" already has enough promises, hence ignoring the promise", 0)
		}
		return
	}

	if message.Promise && int32(message.LastPromisedBallot) == rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.preparedBallot {
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulPromises++
		if rp.debugOn {
			rp.debug("Instance "+strconv.Itoa(int(message.InstanceNumber))+" received a promise, hence incrementing the promise count", 0)
		}
		if int32(message.LastAcceptedBallot) > rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.highestSeenAcceptedBallot {
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.highestSeenAcceptedBallot = int32(message.LastAcceptedBallot)
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.highestSeenAcceptedValue = *message.LastAcceptedValue
		}
		if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulPromises == rp.baxosConsensus.quorumSize {
			if rp.debugOn {
				rp.debug("Instance "+strconv.Itoa(int(message.InstanceNumber))+" received quorum of promises, hence proposing", 0)
			}
			rp.sendPropose(message.InstanceNumber)
		}

	}
}

// invoked upon receiving a client batch

func (rp *Replica) tryPropose() {

	if rp.baxosConsensus.isProposing || rp.baxosConsensus.isBackingOff {
		if rp.debugOn {
			rp.debug("Already proposing or backing off, hence ignoring the propose request", 0)
		}
		return
	}

	nextFreeInstance := rp.baxosConsensus.lastCommittedLogIndex + 1
	rp.createInstance(int(nextFreeInstance))
	for rp.baxosConsensus.replicatedLog[nextFreeInstance].decided {
		nextFreeInstance++
		rp.createInstance(int(nextFreeInstance))
	}

	if rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.numSuccessfulPromises >= rp.baxosConsensus.quorumSize {
		if rp.debugOn {
			rp.debug("Already have enough promises, hence proposing", 0)
		}
		rp.sendPropose(nextFreeInstance)
	} else {
		if rp.debugOn {
			rp.debug("Not enough promises, hence sending prepare", 0)
		}
		rp.sendPrepare()
		rp.baxosConsensus.isProposing = true
	}

}

/*
	propose a command for instance
*/

func (rp *Replica) sendPropose(instance int32) {

	rp.createInstance(int(instance + 1))

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
	if rp.debugOn {
		rp.debug("Broadcast propose for instance "+strconv.Itoa(int(instance))+" for ballot "+strconv.Itoa(int(rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.preparedBallot)), 0)

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
		if rp.debugOn {
			rp.debug("Instance "+strconv.Itoa(int(message.InstanceNumber))+" already decided, hence ignoring the accept", 0)
		}
		return
	}

	if message.Decided && !rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
		if rp.debugOn {
			rp.debug("Instance "+strconv.Itoa(int(message.InstanceNumber))+" decided using accept response, hence setting the decided value", 0)
		}
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided = true
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].decidedValue = *message.DecidedValue
		rp.updateSMR()
		return
	}

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulAccepts >= rp.baxosConsensus.quorumSize {
		if rp.debugOn {
			rp.debug("Instance "+strconv.Itoa(int(message.InstanceNumber))+" already has enough accepts, hence ignoring the accept", 0)
		}
		return
	}

	if message.Accept && int32(message.AcceptBallot) == rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.preparedBallot {
		if rp.debugOn {
			rp.debug("Instance "+strconv.Itoa(int(message.InstanceNumber))+" received an accept, hence incrementing the accept count", 0)
		}
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulAccepts++
		if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulAccepts == rp.baxosConsensus.quorumSize {
			if !rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
				rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided = true
				rp.baxosConsensus.replicatedLog[message.InstanceNumber].decidedValue = rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.proposedValue
				rp.updateSMR()
			}
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
