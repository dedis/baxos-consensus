package tests

import (
	"bytes"
	"paxos_raft/replica/src"
	"testing"
)

func TestClientBatchMarshalUnmarshal(t *testing.T) {
	// Step 1: Generate a ClientBatch

	requests := []*src.SingleOperation{{Command: "a"}, {Command: "b"}}
	originalClientBatch := &src.ClientBatch{
		UniqueId: "a",
		Requests: requests,
		Sender:   1,
	}

	// Step 2: Marshal the ClientBatch
	var buf bytes.Buffer
	err := originalClientBatch.Marshal(&buf)
	if err != nil {
		t.Fatalf("Failed to marshal ClientBatch: %v", err)
	}

	// Step 3: Unmarshal into a new ClientBatch
	newClientBatch := &src.ClientBatch{}
	err = newClientBatch.Unmarshal(&buf)
	if err != nil {
		t.Fatalf("Failed to unmarshal ClientBatch: %v", err)
	}

	// Step 4: Compare the original and the unmarshalled ClientBatch
	if !isEqual(originalClientBatch, newClientBatch) {
		t.Fatalf("Original and unmarshalled ClientBatch are not equal:\nOriginal: %+v\nNew: %+v", originalClientBatch, newClientBatch)
	}
}

// isEqual compares two ClientBatch objects for equality.
func isEqual(a, b *src.ClientBatch) bool {
	if a.Sender != b.Sender {
		return false
	}
	if a.UniqueId != b.UniqueId {
		return false
	}
	if len(a.Requests) != len(b.Requests) {
		return false
	}
	for i := 0; i < len(a.Requests); i++ {
		if a.Requests[i].Command != b.Requests[i].Command {
			return false
		}
	}
	return true

}
