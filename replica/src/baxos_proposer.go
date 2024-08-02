package src

import (
	"baxos/common"
	"math/rand"
	"time"
)

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
