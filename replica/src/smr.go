package src

import (
	"time"
)

// add the client batch to buffer and propose

func (rp *Replica) handleClientBatch(batch *ClientBatch) {
	rp.incomingRequests = append(rp.incomingRequests, batch)

	if (time.Now().Sub(rp.lastProposedTime).Microseconds() > int64(rp.replicaBatchTime) && len(rp.incomingRequests) > 0) || len(rp.incomingRequests) >= rp.replicaBatchSize {
		var proposals []*ClientBatch
		if len(rp.incomingRequests) > rp.replicaBatchSize {
			proposals = rp.incomingRequests[:rp.replicaBatchSize]
			rp.incomingRequests = rp.incomingRequests[rp.replicaBatchSize:]
		} else {
			proposals = rp.incomingRequests
			rp.incomingRequests = make([]*ClientBatch, 0)
		}
		rp.sendPropose(proposals)
		rp.lastProposedTime = time.Now()
	} else {
		//rp.debug("Still did not invoke propose from smr, num client batches = "+strconv.Itoa(int(len(rp.incomingRequests)))+" ,time since last proposal "+strconv.Itoa(int(time.Now().Sub(rp.lastProposedTime).Microseconds())), 0)
	}

}

// call the state machine

func (rp *Replica) updateApplicationLogic(requests []*ClientBatch) []*ClientBatch {
	return rp.state.Execute(requests)
}

// send back the client responses

func (rp *Replica) sendClientResponses(responses []*ClientBatch) {
	for i := 0; i < len(responses); i++ {
		if responses[i].Sender == -1 {
			continue
		}
		rp.sendMessage(int32(responses[i].Sender), RPCPair{
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
				clientBatch := ClientBatch{
					UniqueId: "nil",
					Requests: make([]*SingleOperation, 0),
					Sender:   -1,
				}
				rp.sendMessage(rp.name, RPCPair{
					Code: rp.messageCodes.ClientBatchRpc,
					Obj:  &clientBatch,
				})
				break
			}

		}
	}()
}
