package src

import (
	"baxos/common"
	"math/rand"
	"strconv"
	"time"
)

// this is triggered when the proposer timer timesout after waiting for promise / accept messages

func (rp *Replica) randomBackOff(instance int64) {
	rp.baxosConsensus.isBackingOff = true
	rp.baxosConsensus.timer = nil
	rp.baxosConsensus.retries++

	// reset the proposer bookkeeping
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.preparedBallot = -1
	rp.baxosConsensus.replicatedLog[instance].proposer_bookkeeping.numSuccessfulPrepares = 0
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

// this is triggered after the backoff timer timeouts and the proposer is ready to propose again

func (rp *Replica) proposeAfterBackingOff() {
	rp.sendPrepare()
}

/*
	send a prepare message to lsatCommittedIndex + 1
*/

func (rp *Replica) sendPrepare() {

}

/*
	Handler for promise messages
*/

func (rp *Replica) handlePromise(message *common.PromiseReply) {

}

/*
	leader invokes this function to replicate a new instance for lastProposedLogIndex +1
	can be invoked when receiving a new client batch / when the promise set is successful
	should propose only if there is no outstanding proposal
*/

func (rp *Replica) sendPropose() {

}

/*
	handler for accept messages
*/

func (rp *Replica) handleAccept(message *common.AcceptReply) {

}
