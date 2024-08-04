package src

import (
	"baxos/common"
	"math"
	"math/rand"
	"os"
	"strconv"
)

type BaxosProposerInstance struct {
	preparedBallot        int32 // the ballot number for which the prepare message was sent
	numSuccessfulPromises int32 // the number of successful promise messages received

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
	lastCommittedLogIndex int32                   // the last log position that is committed
	replicatedLog         []BaxosInstance         // the replicated log of commands
	timer                 *common.TimerWithCancel // the timer for collecting promise / accept responses
	roundTripTime         int64                   // network round trip time in microseconds
	timeOutChan           chan int64              // to indicate that the timer has timed out

	isBackingOff bool                    // if the replica is backing off
	isProposing  bool                    // if the replica is proposing
	wakeupTimer  *common.TimerWithCancel // to wake up after backing off
	wakeupChan   chan bool               // to indicate that backoff period is completed
	retries      int                     // number of retries

	replica *Replica

	isAsync      bool // to simulate an asynchronous network
	asyncTimeout int

	quorumSize int32
}

/*
	init Baxos Consensus data structs
*/

func InitBaxosConsensus(replica *Replica, isAsync bool, asyncTimeout int, roundTripTime int64) *Baxos {

	replicatedLog := make([]BaxosInstance, 0)
	// create the genesis slot
	replicatedLog = append(replicatedLog, BaxosInstance{
		decidedValue: common.ReplicaBatch{},
		decided:      true,
	})

	return &Baxos{
		lastCommittedLogIndex: 0,
		replicatedLog:         replicatedLog,
		timer:                 nil,
		roundTripTime:         roundTripTime,
		timeOutChan:           make(chan int64, 10000),
		isBackingOff:          false,
		isProposing:           false,
		wakeupTimer:           nil,
		wakeupChan:            make(chan bool, 10000),
		retries:               0,
		replica:               replica,
		isAsync:               isAsync,
		asyncTimeout:          asyncTimeout,
		quorumSize:            int32(replica.numReplicas/2 + 1),
	}
}

// calculate the backoff time for the proposer

func (rp *Replica) calculateBackOffTime() int64 {
	// k × 2^retries × 2 × RTT
	k := 1.0 - rand.Float64()
	backoff_time := k * math.Pow(2, float64(rp.baxosConsensus.retries+1)) * float64(rp.baxosConsensus.roundTripTime)
	return int64(backoff_time)
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
		rp.handlePromise(promiseReply, true)
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
			rp.debug("Received an accept message from "+strconv.Itoa(int(acceptReply.Sender))+" for instance "+strconv.Itoa(int(acceptReply.InstanceNumber))+" for accept ballot "+strconv.Itoa(int(acceptReply.AcceptBallot)), 0)
		}
		rp.handleAccept(acceptReply)
	}
}

/*
	create instance number n
*/

func (rp *Replica) createInstance(n int) {

	if len(rp.baxosConsensus.replicatedLog) > n {
		// already exists
		return
	}

	numberOfNewInstances := n - len(rp.baxosConsensus.replicatedLog) + 1

	for i := 0; i < numberOfNewInstances; i++ {

		rp.baxosConsensus.replicatedLog = append(rp.baxosConsensus.replicatedLog, BaxosInstance{
			proposer_bookkeeping: BaxosProposerInstance{
				preparedBallot:            -1,
				numSuccessfulPromises:     0,
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
	print the replicated log to check for log consistency
*/

func (rp *Replica) printBaxosLogConsensus() {
	f, err := os.Create(rp.logFilePath + strconv.Itoa(int(rp.name)) + "-consensus.txt")
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	for i := int32(0); i <= rp.baxosConsensus.lastCommittedLogIndex; i++ {
		if rp.baxosConsensus.replicatedLog[i].decided == false {
			panic("should not happen")
		}
		for j := 0; j < len(rp.baxosConsensus.replicatedLog[i].decidedValue.Requests); j++ {
			for k := 0; k < len(rp.baxosConsensus.replicatedLog[i].decidedValue.Requests[j].Requests); k++ {
				_, _ = f.WriteString(strconv.Itoa(int(i)) + "-" + strconv.Itoa(j) + "-" + strconv.Itoa(k) + ":" + rp.baxosConsensus.replicatedLog[i].decidedValue.Requests[j].Requests[k].Command + "\n")
			}
		}
	}
}

/*
	update SMR logic
*/

func (rp *Replica) updateSMR() {

	for i := rp.baxosConsensus.lastCommittedLogIndex + 1; i < int32(len(rp.baxosConsensus.replicatedLog)); i++ {

		if rp.baxosConsensus.replicatedLog[i].decided == true {
			var clientResponses []*common.ClientBatch
			clientResponses = rp.updateApplicationLogic(rp.baxosConsensus.replicatedLog[i].decidedValue.Requests)
			rp.sendClientResponses(clientResponses)
			if rp.debugOn {
				rp.debug("Committed baxos consensus instance "+"."+strconv.Itoa(int(i)), 0)
			}
			rp.baxosConsensus.lastCommittedLogIndex = i
		} else {
			break
		}
	}
}
