package src

import (
	"baxos/common"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type BaxosProposerInstance struct {
	preparedBallot        int32 // the ballot number for which the prepare message was sent
	numSuccessfulPrepares int32 // the number of successful prepare messages received

	highestSeenAcceptedBallot int32               // the highest accepted ballot number among them set of Promise messages
	highestSeenAcceptedValue  common.ReplicaBatch // the highest accepted value among the set of Promise messages

	proposedValue        common.ReplicaBatch // the value that is proposed
	numSuccessfulAccepts int32               // the number of successful accept messages received
}

type BaxosAcceptorInstance struct {
	promisedBallot int32
	acceptedBallot int32
	acceptedValue  common.ReplicaBatch
}

/*
	instance defines the content of a single Baxos consensus instance
*/

type BaxosInstance struct {
	proposer_bookkeeping BaxosProposerInstance
	acceptor_bookkeeping BaxosAcceptorInstance
	decidedValue         common.ReplicaBatch
	decided              bool
}

/*
	Baxos struct defines the replica wide consensus variables
*/

type Baxos struct {
	name int32

	lastCommittedLogIndex int32                   // the last log position that is committed
	replicatedLog         []BaxosInstance         // the replicated log of commands
	timer                 *common.TimerWithCancel // the timer for collecting promise / accept responses
	roundTripTime         int64                   // network round trip time in microseconds

	isBackingOff bool                    // if the replica is backing off
	wakeupTimer  *common.TimerWithCancel // to wake up after backing off
	retries      int

	startTime time.Time // time when the consensus was started

	replica *Replica

	isAsync      bool
	asyncTimeout int
}

// external API for Baxos messages

func (rp *Replica) handleBaxosConsensus(message common.Serializable, code uint8) {

	if code == rp.messageCodes.PrepareRequest {
		prepareRequest := message.(*common.PrepareRequest)
		if rp.debugOn {
			rp.debug("Received a prepare message from "+strconv.Itoa(int(prepareRequest.Sender))+" for instance "+strconv.Itoa(int(prepareRequest.InstanceNumber))+" for prepare ballot "+strconv.Itoa(int(prepareRequest.PrepareBallot)), 0)
		}
		rp.handlePrepare(prepareRequest)
	}

	if code == rp.messageCodes.PromiseReply {
		promiseReply := message.(*common.PromiseReply)
		if rp.debugOn {
			rp.debug("Received a promise message from "+strconv.Itoa(int(promiseReply.Sender))+" for instance "+strconv.Itoa(int(promiseReply.InstanceNumber))+" for promise ballot "+strconv.Itoa(int(promiseReply.LastPromisedBallot)), 0)
		}
		rp.handlePromise(promiseReply)
	}

	if code == rp.messageCodes.ProposeRequest {
		proposeRequest := message.(*common.ProposeRequest)
		if rp.debugOn {
			rp.debug("Received a propose message from "+strconv.Itoa(int(proposeRequest.Sender))+" for instance "+strconv.Itoa(int(proposeRequest.InstanceNumber))+" for propose ballot "+strconv.Itoa(int(proposeRequest.ProposeBallot)), 0)
		}
		rp.handlePropose(proposeRequest)
	}

	if code == rp.messageCodes.AcceptReply {
		acceptReply := message.(*common.AcceptReply)
		if rp.debugOn {
			rp.debug("Received a accept message from "+strconv.Itoa(int(acceptReply.Sender))+" for instance "+strconv.Itoa(int(acceptReply.InstanceNumber))+" for accept ballot "+strconv.Itoa(int(acceptReply.AcceptBallot)), 0)
		}
		rp.handleAccept(acceptReply)
	}
}

/*
	init Paxos Consensus data structs
*/

func InitBaxosConsensus(name int32, replica *Replica, isAsync bool, asyncTimeout int, round_trip_time int64) *Baxos {

	replicatedLog := make([]BaxosInstance, 0)
	// create the genesis slot
	replicatedLog = append(replicatedLog, BaxosInstance{
		decidedValue: common.ReplicaBatch{},
		decided:      true,
	})

	return &Baxos{
		name:                  name,
		lastCommittedLogIndex: 0,
		replicatedLog:         replicatedLog,
		timer:                 nil,
		roundTripTime:         round_trip_time,
		isBackingOff:          false,
		wakeupTimer:           nil,
		retries:               0,
		startTime:             time.Now(),
		replica:               replica,
		isAsync:               isAsync,
		asyncTimeout:          asyncTimeout,
	}
}

/*
	create instance number n
*/

func (rp *Replica) createInstance(n int) {

	if len(rp.baxosConsensus.replicatedLog) > n {
		return
	}

	number_of_new_instances := n - len(rp.baxosConsensus.replicatedLog) + 1

	for i := 0; i < number_of_new_instances; i++ {

		rp.baxosConsensus.replicatedLog = append(rp.baxosConsensus.replicatedLog, BaxosInstance{
			proposer_bookkeeping: BaxosProposerInstance{
				preparedBallot:            -1,
				numSuccessfulPrepares:     0,
				highestSeenAcceptedBallot: -1,
				highestSeenAcceptedValue:  common.ReplicaBatch{},
				proposedValue:             common.ReplicaBatch{},
				numSuccessfulAccepts:      0,
			},
			acceptor_bookkeeping: BaxosAcceptorInstance{
				promisedBallot: -1,
				acceptedBallot: -1,
				acceptedValue:  common.ReplicaBatch{},
			},
			decidedValue: common.ReplicaBatch{},
			decided:      false,
		})
	}
}

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
	Handler for prepare message, check if it is possible to promise for the specified instance
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
	handler for propose message, If the propose ballot number is greater than or equal to the promised ballot number,
	set the accepted ballot and accepted values, and send
	an accept message, also record the decided message for the previous instance
*/

func (rp *Replica) handlePropose(message *common.ProposeRequest) {
	// handle the propose slot
	acceptReply := rp.processPropose(message)

	// handle the prepare slot
	promiseResponse := rp.processPrepare(message.PrepareRequestForFutureIndex)
	acceptReply.PromiseReplyForFutureIndex = promiseResponse

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

/*
	Sets a timer, which once timeout will send an internal notification for a prepare message after another random wait to break the ties
*/

func (rp *Replica) setPaxosViewTimer(view int32) {

	rp.paxosConsensus.viewTimer = common.NewTimerWithCancel(time.Duration(rp.viewTimeout+rand.Intn(rp.viewTimeout+int(rp.name))) * time.Microsecond)

	rp.paxosConsensus.viewTimer.SetTimeoutFuntion(func() {

		// this function runs in a separate thread, hence we do not send prepare message in this function, instead send a timeout-internal signal
		internalTimeoutNotification := PaxosConsensus{
			Sender:   rp.name,
			Receiver: rp.name,
			Type:     5,
			View:     view,
		}

		rpcPair := common.RPCPair{
			Code: rp.messageCodes.PaxosConsensus,
			Obj:  &internalTimeoutNotification,
		}
		rp.sendMessage(rp.name, rpcPair)
		//rp.debug("Sent an internal timeout notification for view "+strconv.Itoa(int(view)), 0)

	})
	rp.paxosConsensus.viewTimer.Start()
}

/*
	print the replicated log to check for log consistency
*/

func (rp *Replica) printBaxosLogConsensus() {
	f, err := os.Create(rp.logFilePath + strconv.Itoa(int(rp.name)) + "-consensus.txt")
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	for i := int32(0); i < rp.paxosConsensus.lastCommittedLogIndex; i++ {
		if rp.paxosConsensus.replicatedLog[i].decided == false {
			panic("should not happen")
		}
		for j := 0; j < len(rp.paxosConsensus.replicatedLog[i].decidedValue.Requests); j++ {
			for k := 0; k < len(rp.paxosConsensus.replicatedLog[i].decidedValue.Requests[j].Requests); k++ {
				_, _ = f.WriteString(strconv.Itoa(int(i)) + "-" + strconv.Itoa(j) + "-" + strconv.Itoa(k) + ":" + rp.paxosConsensus.replicatedLog[i].decidedValue.Requests[j].Requests[k].Command + "\n")

			}
		}
	}
}

/*
	upon a view change / upon bootstrap send a prepare message for all instances from last committed index +1 to len(log)
*/

func (rp *Replica) sendPrepare() {

	//rp.debug("sending prepare for view "+strconv.Itoa(int(rp.paxosConsensus.view)), 7)

	rp.createPaxosInstanceIfMissing(int(rp.paxosConsensus.lastCommittedLogIndex + 1))

	// reset the promise response map, because all we care is new view change messages
	rp.paxosConsensus.promiseResponses = make(map[int32][]*PaxosConsensus)

	if rp.paxosConsensus.lastPromisedBallot > rp.paxosConsensus.lastPreparedBallot {
		rp.paxosConsensus.lastPreparedBallot = rp.paxosConsensus.lastPromisedBallot
	}
	rp.paxosConsensus.lastPreparedBallot = rp.paxosConsensus.lastPreparedBallot + 100*rp.name + 2

	rp.paxosConsensus.state = "C" // become a contestant
	// increase the view number
	rp.paxosConsensus.view++
	// broadcast a prepare message
	for name, _ := range rp.replicaAddrList {
		prepareMsg := PaxosConsensus{
			Sender:         rp.name,
			Receiver:       name,
			Type:           1,
			InstanceNumber: rp.paxosConsensus.lastCommittedLogIndex + 1,
			Ballot:         rp.paxosConsensus.lastPreparedBallot,
			View:           rp.paxosConsensus.view,
		}

		rpcPair := common.RPCPair{
			Code: rp.messageCodes.PaxosConsensus,
			Obj:  &prepareMsg,
		}

		rp.sendMessage(name, rpcPair)
		//rp.debug("Sent prepare to "+strconv.Itoa(int(name)), 0)
	}

	// cancel the view timer
	if rp.paxosConsensus.viewTimer != nil {
		rp.paxosConsensus.viewTimer.Cancel()
		rp.paxosConsensus.viewTimer = nil
	}
	// set the view timer
	rp.setPaxosViewTimer(rp.paxosConsensus.view)

	// cancel the current leader
	rp.paxosConsensus.currentLeader = -1
}

/*
	Handler for promise messages
*/

func (rp *Replica) handlePromise(message *PaxosConsensus) {
	if message.Ballot == rp.paxosConsensus.lastPreparedBallot && message.View == rp.paxosConsensus.view && rp.paxosConsensus.state == "C" {
		// save the promise message
		_, ok := rp.paxosConsensus.promiseResponses[message.View]
		if ok {
			rp.paxosConsensus.promiseResponses[message.View] = append(rp.paxosConsensus.promiseResponses[message.View], message)
		} else {
			rp.paxosConsensus.promiseResponses[message.View] = make([]*PaxosConsensus, 0)
			rp.paxosConsensus.promiseResponses[message.View] = append(rp.paxosConsensus.promiseResponses[message.View], message)
		}

		if len(rp.paxosConsensus.promiseResponses[message.View]) == rp.numReplicas/2+1 {
			// cancel the view timer
			if rp.paxosConsensus.viewTimer != nil {
				rp.paxosConsensus.viewTimer.Cancel()
				rp.paxosConsensus.viewTimer = nil
			}
			// we have majority promise messages for the same view
			// update the highest accepted ballot and the values
			for i := 0; i < len(rp.paxosConsensus.promiseResponses[message.View]); i++ {
				lastAcceptedEntries := rp.paxosConsensus.promiseResponses[message.View][i].PromiseReply
				for j := 0; j < len(lastAcceptedEntries); j++ {
					instanceNumber := lastAcceptedEntries[j].Number
					rp.createPaxosInstanceIfMissing(int(instanceNumber))
					if lastAcceptedEntries[j].Ballot > rp.paxosConsensus.replicatedLog[instanceNumber].highestSeenAcceptedBallot {
						rp.paxosConsensus.replicatedLog[instanceNumber].highestSeenAcceptedBallot = lastAcceptedEntries[j].Ballot
						rp.paxosConsensus.replicatedLog[instanceNumber].highestSeenAcceptedValue = *lastAcceptedEntries[j].Value
					}
				}
			}
			rp.paxosConsensus.state = "L"
			//rp.debug("Became the leader in view "+strconv.Itoa(int(rp.paxosConsensus.view)), 7)
			rp.paxosConsensus.currentLeader = rp.name
		}
	}

}

/*
	leader invokes this function to replicate a new instance for lastProposedLogIndex +1
*/

func (rp *Replica) sendPropose(requests []*common.ClientBatch) { // requests can be empty
	if rp.paxosConsensus.isBackingOff {
		return
	}
	if rp.paxosConsensus.lastProposedLogIndex < rp.paxosConsensus.lastCommittedLogIndex {
		rp.paxosConsensus.lastProposedLogIndex = rp.paxosConsensus.lastCommittedLogIndex
	}

	if rp.paxosConsensus.state == "L" &&
		rp.paxosConsensus.lastPreparedBallot >= rp.paxosConsensus.lastPromisedBallot {

		rp.paxosConsensus.lastProposedLogIndex++
		rp.createPaxosInstanceIfMissing(int(rp.paxosConsensus.lastProposedLogIndex))

		for rp.paxosConsensus.replicatedLog[rp.paxosConsensus.lastProposedLogIndex].decided {
			rp.paxosConsensus.lastProposedLogIndex++
			rp.createPaxosInstanceIfMissing(int(rp.paxosConsensus.lastProposedLogIndex))
		}

		proposeValue := &common.ReplicaBatch{
			UniqueId: "",
			Requests: requests,
			Sender:   int64(rp.name),
		}
		if rp.paxosConsensus.replicatedLog[rp.paxosConsensus.lastProposedLogIndex].highestSeenAcceptedBallot != -1 {
			proposeValue = &rp.paxosConsensus.replicatedLog[rp.paxosConsensus.lastProposedLogIndex].highestSeenAcceptedValue
			rp.incomingRequests = append(rp.incomingRequests, requests...)
		}

		// set the proposed ballot for this instance
		rp.paxosConsensus.replicatedLog[rp.paxosConsensus.lastProposedLogIndex].proposedBallot = rp.paxosConsensus.lastPreparedBallot
		rp.paxosConsensus.replicatedLog[rp.paxosConsensus.lastProposedLogIndex].proposeResponses = 0
		rp.paxosConsensus.replicatedLog[rp.paxosConsensus.lastProposedLogIndex].proposedValue = *proposeValue

		decided_values := make([]*PaxosConsensusInstance, 0)

		for i := 0; i < len(rp.paxosConsensus.decidedIndexes); i++ {
			decided_values = append(decided_values, &PaxosConsensusInstance{
				Number: int32(rp.paxosConsensus.decidedIndexes[i]),
				Value:  &rp.paxosConsensus.replicatedLog[rp.paxosConsensus.decidedIndexes[i]].decidedValue,
			})
		}
		// reset decided indexes
		rp.paxosConsensus.decidedIndexes = make([]int, 0)

		if rp.isAsynchronous {

			epoch := time.Now().Sub(rp.paxosConsensus.startTime).Milliseconds() / int64(rp.timeEpochSize)

			if rp.amIAttacked(int(epoch)) {
				time.Sleep(time.Duration(rp.asyncSimulationTimeout) * time.Millisecond)
			}
		}

		// send a propose message
		for name, _ := range rp.replicaAddrList {
			proposeMsg := PaxosConsensus{
				Sender:         rp.name,
				Receiver:       name,
				Type:           3,
				InstanceNumber: rp.paxosConsensus.lastProposedLogIndex,
				Ballot:         rp.paxosConsensus.lastPreparedBallot,
				View:           rp.paxosConsensus.view,
				ProposeValue:   proposeValue,
				DecidedValues:  decided_values,
			}

			rpcPair := common.RPCPair{
				Code: rp.messageCodes.PaxosConsensus,
				Obj:  &proposeMsg,
			}

			rp.sendMessage(name, rpcPair)
			//rp.debug("Sent propose to "+strconv.Itoa(int(name)), 1)
		}
		//rp.debug("Sent proposal for index "+strconv.Itoa(int(rp.paxosConsensus.lastProposedLogIndex))+" in view "+strconv.Itoa(int(rp.paxosConsensus.view)), 7)
		if rp.paxosConsensus.viewTimer != nil {
			rp.paxosConsensus.viewTimer.Cancel()
			rp.paxosConsensus.viewTimer = nil
		}
		rp.setPaxosViewTimer(rp.paxosConsensus.view)
	} else if rp.paxosConsensus.state == "L" && rp.paxosConsensus.lastPreparedBallot >= rp.paxosConsensus.lastPromisedBallot {
		rp.incomingRequests = append(rp.incomingRequests, requests...)
		//rp.debug("saving for later proposal due to full pipeline while I am the leader "+" in view "+strconv.Itoa(int(rp.paxosConsensus.view)), 0)
	} else {
		//rp.debug("dropping requests because i am not the leader "+" in view "+strconv.Itoa(int(rp.paxosConsensus.view)), 0)
	}
}

/*
	handler for accept messages. Upon collecting n-f accept messages, mark the instance as decided, call SMR and
*/

func (rp *Replica) handleAccept(message *PaxosConsensus) {
	if int32(len(rp.paxosConsensus.replicatedLog)) < message.InstanceNumber+1 {
		panic("Received accept without having an instance")
	}

	if message.View <= rp.paxosConsensus.view && message.Ballot == rp.paxosConsensus.replicatedLog[message.InstanceNumber].proposedBallot && rp.paxosConsensus.state == "L" {

		// add the accept to proposeResponses
		rp.paxosConsensus.replicatedLog[message.InstanceNumber].proposeResponses++

		// if there are n-f accept messages
		if rp.paxosConsensus.replicatedLog[message.InstanceNumber].proposeResponses == rp.numReplicas/2+1 && !rp.paxosConsensus.replicatedLog[message.InstanceNumber].decided {

			if rp.paxosConsensus.viewTimer != nil {
				rp.paxosConsensus.viewTimer.Cancel()
				rp.paxosConsensus.viewTimer = nil
			}

			rp.paxosConsensus.replicatedLog[message.InstanceNumber].decided = true
			rp.paxosConsensus.replicatedLog[message.InstanceNumber].decidedValue = rp.paxosConsensus.replicatedLog[message.InstanceNumber].proposedValue
			rp.paxosConsensus.replicatedLog[message.InstanceNumber].proposedValue = common.ReplicaBatch{}
			//rp.debug("Decided upon receiving n-f accept message for instance "+strconv.Itoa(int(message.InstanceNumber)), 7)
			rp.paxosConsensus.decidedIndexes = append(rp.paxosConsensus.decidedIndexes, int(message.InstanceNumber))
			rp.updatePaxosSMR()
			rp.setPaxosViewTimer(rp.paxosConsensus.view)
		}
	}
}

/*
	handler for internal timeout messages, send a prepare message
*/

func (rp *Replica) handlePaxosInternalTimeout(message *PaxosConsensus) {
	// Retry from the prepare stage

}

/*
	update SMR logic
*/

func (rp *Replica) updateSMR() {

	for i := rp.paxosConsensus.lastCommittedLogIndex + 1; i < int32(len(rp.paxosConsensus.replicatedLog)); i++ {

		if rp.paxosConsensus.replicatedLog[i].decided == true {
			var cllientResponses []*common.ClientBatch
			cllientResponses = rp.updateApplicationLogic(rp.paxosConsensus.replicatedLog[i].decidedValue.Requests)
			if rp.paxosConsensus.replicatedLog[i].decidedValue.Sender == int64(rp.name) {
				rp.sendClientResponses(cllientResponses)
			}
			//rp.debug("Committed paxos consensus instance "+"."+strconv.Itoa(int(i)), 7)
			rp.paxosConsensus.lastCommittedLogIndex = i
		} else {
			break
		}

	}
}
