package src

import (
	"baxos/common"
	"time"
)

// add the client batch to buffer and propose

func (rp *Replica) handleClientBatch(batch *common.ClientBatch) {
	rp.incomingRequests = append(rp.incomingRequests, batch)

	if (time.Now().Sub(rp.lastProposedTime).Microseconds() > int64(rp.replicaBatchTime) && len(rp.incomingRequests) > 0) || len(rp.incomingRequests) >= rp.replicaBatchSize {
		var proposals []*common.ClientBatch
		if len(rp.incomingRequests) > rp.replicaBatchSize {
			proposals = rp.incomingRequests[:rp.replicaBatchSize]
			rp.incomingRequests = rp.incomingRequests[rp.replicaBatchSize:]
		} else {
			proposals = rp.incomingRequests
			rp.incomingRequests = make([]*common.ClientBatch, 0)
		}
		rp.sendPropose(proposals)
		rp.lastProposedTime = time.Now()
	} else {
		//rp.debug("Still did not invoke propose from smr, num client batches = "+strconv.Itoa(int(len(rp.incomingRequests)))+" ,time since last proposal "+strconv.Itoa(int(time.Now().Sub(rp.lastProposedTime).Microseconds())), 0)
	}

}

// call the state machine

func (rp *Replica) updateApplicationLogic(requests []*common.ClientBatch) []*common.ClientBatch {
	return rp.state.Execute(requests)
}

// send back the client responses

func (rp *Replica) sendClientResponses(responses []*common.ClientBatch) {
	for i := 0; i < len(responses); i++ {
		if responses[i].Sender == -1 {
			continue
		}
		rp.sendMessage(int32(responses[i].Sender), common.RPCPair{
			Code: rp.messageCodes.ClientBatchRpc,
			Obj:  responses[i],
		})
		//rp.debug("send client response to "+strconv.Itoa(int(responses[i].Sender)), 0)
	}
}

// send dummy requests to avoid leader revokes

func (rp *Replica) sendDummyRequests(cancel chan bool) {
	go func() {
		for true {
			select {
			case _ = <-cancel:
				return
			default:
				time.Sleep(time.Duration(rp.viewTimeout/2) * time.Microsecond)
				clientBatch := common.ClientBatch{
					UniqueId: "nil",
					Requests: make([]*common.SingleOperation, 0),
					Sender:   -1,
				}
				rp.sendMessage(rp.name, common.RPCPair{
					Code: rp.messageCodes.ClientBatchRpc,
					Obj:  &clientBatch,
				})
				break
			}

		}
	}()
}
