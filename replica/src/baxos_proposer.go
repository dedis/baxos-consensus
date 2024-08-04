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

	})
	rp.baxosConsensus.timer.Start()
}

// this is triggered when the proposer timer timeout after waiting for promise / accept messages

func (rp *Replica) randomBackOff(instance int64) {
	if rp.debugOn {
		rp.debug("PROPOSER: Timed out when proposing for instance "+strconv.Itoa(int(instance)), 1)
	}

	if rp.baxosConsensus.replicatedLog[instance].decided && rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.numSuccessfulAccepts >= rp.baxosConsensus.quorumSize {
		if rp.debugOn {
			rp.debug("PROPOSER: Instance "+strconv.Itoa(int(instance))+" already decided, hence ignoring the timeout indication", 1)
		}
		return
	}

	rp.baxosConsensus.isBackingOff = true
	rp.baxosConsensus.isProposing = true
	rp.baxosConsensus.timer = nil
	rp.baxosConsensus.retries++
	if rp.baxosConsensus.retries > 10 {
		rp.baxosConsensus.retries = 10
	}

	// reset the proposer bookkeeping
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.preparedBallot = -1
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.numSuccessfulPromises = 0
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.highestSeenAcceptedBallot = -1
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.highestSeenAcceptedValue = common.ReplicaBatch{}
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.proposedValue = common.ReplicaBatch{}
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.numSuccessfulAccepts = 0

	if rp.debugOn {
		rp.debug("PROPOSER: reset the instance "+strconv.Itoa(int(instance))+" proposer bookkeeping after timeout", 1)
	}

	// set the backing off timer
	backoffTime := rp.calculateBackOffTime()

	if backoffTime > 5000000 {
		backoffTime = 5000000 // maximum 5s cap on backoff time
	}

	if rp.debugOn {
		rp.debug("PROPOSER: Backing off for "+strconv.Itoa(int(backoffTime))+" microseconds", 2)
	}
	rp.baxosConsensus.wakeupTimer = common.NewTimerWithCancel(time.Duration(backoffTime) * time.Microsecond)

	rp.baxosConsensus.wakeupTimer.SetTimeoutFuntion(func() {
		rp.baxosConsensus.wakeupChan <- true
		if rp.debugOn {
			rp.debug("PROPOSER: Finished backing off ", 1)
		}
	})
	rp.baxosConsensus.wakeupTimer.Start()
}

// this is triggered after the backoff timer timeouts and the proposer is ready to propose again

func (rp *Replica) proposeAfterBackingOff() {
	if rp.debugOn {
		rp.debug("PROPOSER: Proposing after backing off", 1)
	}
	rp.baxosConsensus.isBackingOff = false
	rp.sendPrepare()
}

/*
	send a prepare message to lsatCommittedIndex + 1
*/

func (rp *Replica) sendPrepare() {
	nextFreeInstance := rp.baxosConsensus.lastCommittedLogIndex + 1
	rp.createInstance(int(nextFreeInstance))
	for rp.baxosConsensus.replicatedLog[nextFreeInstance].decided {
		nextFreeInstance++
		rp.createInstance(int(nextFreeInstance))
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
		rp.debug("PROPOSER: Sent prepare for instance "+strconv.Itoa(int(nextFreeInstance))+" with prepared ballot "+strconv.Itoa(int(rp.baxosConsensus.replicatedLog[nextFreeInstance].proposer_bookkeeping.preparedBallot)), 1)
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
			rp.debug("PROPOSER: Instance "+strconv.Itoa(int(message.InstanceNumber))+" already decided, hence ignoring the promise", 1)
		}
		return
	}

	if message.Decided {
		if !rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided = true
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].decidedValue = *message.DecidedValue
			if rp.debugOn {
				rp.debug("PROPOSER: Instance "+strconv.Itoa(int(message.InstanceNumber))+" decided using promise response, hence setting the decided value", 1)
			}
			rp.updateSMR()
			return
		}
	}

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulPromises >= rp.baxosConsensus.quorumSize {
		if rp.debugOn {
			rp.debug("PROPOSER: Instance "+strconv.Itoa(int(message.InstanceNumber))+" already has enough promises, hence ignoring the promise", 1)
		}
		return
	}

	if message.Promise && int32(message.LastPromisedBallot) == rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.preparedBallot {
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulPromises++
		if rp.debugOn {
			rp.debug("PROPOSER: Instance "+strconv.Itoa(int(message.InstanceNumber))+" received a promise, hence incrementing the promise count", 1)
		}
		if int32(message.LastAcceptedBallot) > rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.highestSeenAcceptedBallot {
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.highestSeenAcceptedBallot = int32(message.LastAcceptedBallot)
			rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.highestSeenAcceptedValue = *message.LastAcceptedValue
		}
		if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulPromises == rp.baxosConsensus.quorumSize {
			if rp.debugOn {
				rp.debug("PROPOSER: Instance "+strconv.Itoa(int(message.InstanceNumber))+" received quorum of promises, hence proposing", 1)
			}
			rp.sendPropose(message.InstanceNumber)
		}

	}
}

// invoked upon receiving a client batch

func (rp *Replica) tryPropose() {

	if rp.baxosConsensus.isProposing || rp.baxosConsensus.isBackingOff {
		if rp.debugOn {
			rp.debug("PROPOSER: Already proposing or backing off, hence ignoring the propose request", 1)
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
			rp.debug("PROPOSER: Already have enough promises, hence proposing for instance "+strconv.Itoa(int(nextFreeInstance)), 1)
		}
		rp.sendPropose(nextFreeInstance)
	} else {
		if rp.debugOn {
			rp.debug("PROPOSER: Not enough promises for instance "+strconv.Itoa(int(nextFreeInstance))+" hence sending prepare", 1)
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
		if rp.debugOn {
			rp.debug("PROPOSER: No need to prepare for the next instance "+strconv.Itoa(int(instance+1)), 1)
		}
		prepareRequest = nil
	} else {
		rp.baxosConsensus.replicatedLog[instance+1].proposer_bookkeeping.preparedBallot = rp.baxosConsensus.replicatedLog[instance+1].acceptor_bookkeeping.promisedBallot + rp.name + 1
		prepareRequest = &common.PrepareRequest{
			InstanceNumber: instance + 1,
			PrepareBallot:  rp.baxosConsensus.replicatedLog[instance+1].proposer_bookkeeping.preparedBallot,
			Sender:         int64(rp.name),
		}
		if rp.debugOn {
			rp.debug("PROPOSER: Appending prepare message for the next instance "+strconv.Itoa(int(instance+1))+" with ballot "+strconv.Itoa(int(rp.baxosConsensus.replicatedLog[instance+1].proposer_bookkeeping.preparedBallot))+" inside the propose message", 1)
		}
	}

	// set the decided info

	if !rp.baxosConsensus.replicatedLog[instance-1].decided {
		panic("error, previous index not decided")
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
		rp.debug("PROPOSER: Broadcast propose for instance "+strconv.Itoa(int(instance))+" for ballot "+strconv.Itoa(int(rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.preparedBallot)), 1)

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
			rp.debug("PROPOSER: Instance "+strconv.Itoa(int(message.InstanceNumber))+" already decided, hence ignoring the accept", 1)
		}
		return
	}

	if message.Decided && !rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
		if rp.debugOn {
			rp.debug("PROPOSER: Instance "+strconv.Itoa(int(message.InstanceNumber))+" decided using accept response, hence setting the decided value", 1)
		}
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided = true
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].decidedValue = *message.DecidedValue
		rp.updateSMR()
		return
	}

	if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulAccepts >= rp.baxosConsensus.quorumSize {
		if rp.debugOn {
			rp.debug("PROPOSER: Instance "+strconv.Itoa(int(message.InstanceNumber))+" already has enough accepts, hence ignoring the accept", 1)
		}
		return
	}

	if message.Accept && int32(message.AcceptBallot) == rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.preparedBallot {
		if rp.debugOn {
			rp.debug("PROPOSER: Instance "+strconv.Itoa(int(message.InstanceNumber))+" received an accept for the ballot "+strconv.Itoa(int(message.AcceptBallot))+" hence incrementing the accept count", 1)
		}
		rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulAccepts++
		if rp.baxosConsensus.replicatedLog[message.InstanceNumber].proposer_bookkeeping.numSuccessfulAccepts == rp.baxosConsensus.quorumSize {

			if !rp.baxosConsensus.replicatedLog[message.InstanceNumber].decided {
				if rp.debugOn {
					rp.debug("PROPOSER: Instance "+strconv.Itoa(int(message.InstanceNumber))+" received quorum of accepts, hence deciding", 1)
				}
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

			if rp.baxosConsensus.wakeupTimer != nil {
				rp.baxosConsensus.wakeupTimer.Cancel()
			}
			rp.baxosConsensus.wakeupTimer = nil
			rp.baxosConsensus.isBackingOff = false
		}
	}
}
