// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.12.4
// source: common/definitions.proto

package common

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SingleOperation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Command string `protobuf:"bytes,1,opt,name=command,proto3" json:"command,omitempty"`
}

func (x *SingleOperation) Reset() {
	*x = SingleOperation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_definitions_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SingleOperation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SingleOperation) ProtoMessage() {}

func (x *SingleOperation) ProtoReflect() protoreflect.Message {
	mi := &file_common_definitions_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SingleOperation.ProtoReflect.Descriptor instead.
func (*SingleOperation) Descriptor() ([]byte, []int) {
	return file_common_definitions_proto_rawDescGZIP(), []int{0}
}

func (x *SingleOperation) GetCommand() string {
	if x != nil {
		return x.Command
	}
	return ""
}

type ClientBatch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UniqueId string             `protobuf:"bytes,1,opt,name=unique_id,json=uniqueId,proto3" json:"unique_id,omitempty"`
	Requests []*SingleOperation `protobuf:"bytes,2,rep,name=requests,proto3" json:"requests,omitempty"`
	Sender   int64              `protobuf:"varint,3,opt,name=sender,proto3" json:"sender,omitempty"`
}

func (x *ClientBatch) Reset() {
	*x = ClientBatch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_definitions_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientBatch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientBatch) ProtoMessage() {}

func (x *ClientBatch) ProtoReflect() protoreflect.Message {
	mi := &file_common_definitions_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientBatch.ProtoReflect.Descriptor instead.
func (*ClientBatch) Descriptor() ([]byte, []int) {
	return file_common_definitions_proto_rawDescGZIP(), []int{1}
}

func (x *ClientBatch) GetUniqueId() string {
	if x != nil {
		return x.UniqueId
	}
	return ""
}

func (x *ClientBatch) GetRequests() []*SingleOperation {
	if x != nil {
		return x.Requests
	}
	return nil
}

func (x *ClientBatch) GetSender() int64 {
	if x != nil {
		return x.Sender
	}
	return 0
}

type ReplicaBatch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UniqueId string         `protobuf:"bytes,1,opt,name=unique_id,json=uniqueId,proto3" json:"unique_id,omitempty"`
	Requests []*ClientBatch `protobuf:"bytes,2,rep,name=requests,proto3" json:"requests,omitempty"`
}

func (x *ReplicaBatch) Reset() {
	*x = ReplicaBatch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_definitions_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReplicaBatch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReplicaBatch) ProtoMessage() {}

func (x *ReplicaBatch) ProtoReflect() protoreflect.Message {
	mi := &file_common_definitions_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReplicaBatch.ProtoReflect.Descriptor instead.
func (*ReplicaBatch) Descriptor() ([]byte, []int) {
	return file_common_definitions_proto_rawDescGZIP(), []int{2}
}

func (x *ReplicaBatch) GetUniqueId() string {
	if x != nil {
		return x.UniqueId
	}
	return ""
}

func (x *ReplicaBatch) GetRequests() []*ClientBatch {
	if x != nil {
		return x.Requests
	}
	return nil
}

type Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type   int32  `protobuf:"varint,1,opt,name=type,proto3" json:"type,omitempty"` // 1 for bootstrap, 2 for log print, 3 consensus start
	Note   string `protobuf:"bytes,2,opt,name=note,proto3" json:"note,omitempty"`
	Sender int64  `protobuf:"varint,3,opt,name=sender,proto3" json:"sender,omitempty"`
}

func (x *Status) Reset() {
	*x = Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_definitions_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Status) ProtoMessage() {}

func (x *Status) ProtoReflect() protoreflect.Message {
	mi := &file_common_definitions_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Status.ProtoReflect.Descriptor instead.
func (*Status) Descriptor() ([]byte, []int) {
	return file_common_definitions_proto_rawDescGZIP(), []int{3}
}

func (x *Status) GetType() int32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *Status) GetNote() string {
	if x != nil {
		return x.Note
	}
	return ""
}

func (x *Status) GetSender() int64 {
	if x != nil {
		return x.Sender
	}
	return 0
}

type PrepareRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceNumber int32 `protobuf:"varint,1,opt,name=instance_number,json=instanceNumber,proto3" json:"instance_number,omitempty"`
	PrepareBallot  int32 `protobuf:"varint,2,opt,name=prepare_ballot,json=prepareBallot,proto3" json:"prepare_ballot,omitempty"`
	Sender         int64 `protobuf:"varint,3,opt,name=sender,proto3" json:"sender,omitempty"`
}

func (x *PrepareRequest) Reset() {
	*x = PrepareRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_definitions_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PrepareRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrepareRequest) ProtoMessage() {}

func (x *PrepareRequest) ProtoReflect() protoreflect.Message {
	mi := &file_common_definitions_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrepareRequest.ProtoReflect.Descriptor instead.
func (*PrepareRequest) Descriptor() ([]byte, []int) {
	return file_common_definitions_proto_rawDescGZIP(), []int{4}
}

func (x *PrepareRequest) GetInstanceNumber() int32 {
	if x != nil {
		return x.InstanceNumber
	}
	return 0
}

func (x *PrepareRequest) GetPrepareBallot() int32 {
	if x != nil {
		return x.PrepareBallot
	}
	return 0
}

func (x *PrepareRequest) GetSender() int64 {
	if x != nil {
		return x.Sender
	}
	return 0
}

type PromiseReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceNumber     int32         `protobuf:"varint,1,opt,name=instance_number,json=instanceNumber,proto3" json:"instance_number,omitempty"`
	Promise            bool          `protobuf:"varint,2,opt,name=promise,proto3" json:"promise,omitempty"`
	LastPromisedBallot int64         `protobuf:"varint,3,opt,name=last_promised_ballot,json=lastPromisedBallot,proto3" json:"last_promised_ballot,omitempty"`
	LastAcceptedBallot int64         `protobuf:"varint,4,opt,name=last_accepted_ballot,json=lastAcceptedBallot,proto3" json:"last_accepted_ballot,omitempty"`
	LastAcceptedValue  *ReplicaBatch `protobuf:"bytes,5,opt,name=last_accepted_value,json=lastAcceptedValue,proto3" json:"last_accepted_value,omitempty"`
	Decided            bool          `protobuf:"varint,6,opt,name=decided,proto3" json:"decided,omitempty"` // if the current instance is already decided
	DecidedValue       *ReplicaBatch `protobuf:"bytes,7,opt,name=decided_value,json=decidedValue,proto3" json:"decided_value,omitempty"`
	Sender             int64         `protobuf:"varint,8,opt,name=sender,proto3" json:"sender,omitempty"`
}

func (x *PromiseReply) Reset() {
	*x = PromiseReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_definitions_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PromiseReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PromiseReply) ProtoMessage() {}

func (x *PromiseReply) ProtoReflect() protoreflect.Message {
	mi := &file_common_definitions_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PromiseReply.ProtoReflect.Descriptor instead.
func (*PromiseReply) Descriptor() ([]byte, []int) {
	return file_common_definitions_proto_rawDescGZIP(), []int{5}
}

func (x *PromiseReply) GetInstanceNumber() int32 {
	if x != nil {
		return x.InstanceNumber
	}
	return 0
}

func (x *PromiseReply) GetPromise() bool {
	if x != nil {
		return x.Promise
	}
	return false
}

func (x *PromiseReply) GetLastPromisedBallot() int64 {
	if x != nil {
		return x.LastPromisedBallot
	}
	return 0
}

func (x *PromiseReply) GetLastAcceptedBallot() int64 {
	if x != nil {
		return x.LastAcceptedBallot
	}
	return 0
}

func (x *PromiseReply) GetLastAcceptedValue() *ReplicaBatch {
	if x != nil {
		return x.LastAcceptedValue
	}
	return nil
}

func (x *PromiseReply) GetDecided() bool {
	if x != nil {
		return x.Decided
	}
	return false
}

func (x *PromiseReply) GetDecidedValue() *ReplicaBatch {
	if x != nil {
		return x.DecidedValue
	}
	return nil
}

func (x *PromiseReply) GetSender() int64 {
	if x != nil {
		return x.Sender
	}
	return 0
}

type ProposeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceNumber               int32           `protobuf:"varint,1,opt,name=instance_number,json=instanceNumber,proto3" json:"instance_number,omitempty"`
	ProposeBallot                int32           `protobuf:"varint,2,opt,name=propose_ballot,json=proposeBallot,proto3" json:"propose_ballot,omitempty"`
	ProposeValue                 *ReplicaBatch   `protobuf:"bytes,3,opt,name=propose_value,json=proposeValue,proto3" json:"propose_value,omitempty"`
	PrepareRequestForFutureIndex *PrepareRequest `protobuf:"bytes,4,opt,name=prepare_request_for_future_index,json=prepareRequestForFutureIndex,proto3" json:"prepare_request_for_future_index,omitempty"`
	Sender                       int64           `protobuf:"varint,5,opt,name=sender,proto3" json:"sender,omitempty"`
}

func (x *ProposeRequest) Reset() {
	*x = ProposeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_definitions_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProposeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProposeRequest) ProtoMessage() {}

func (x *ProposeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_common_definitions_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProposeRequest.ProtoReflect.Descriptor instead.
func (*ProposeRequest) Descriptor() ([]byte, []int) {
	return file_common_definitions_proto_rawDescGZIP(), []int{6}
}

func (x *ProposeRequest) GetInstanceNumber() int32 {
	if x != nil {
		return x.InstanceNumber
	}
	return 0
}

func (x *ProposeRequest) GetProposeBallot() int32 {
	if x != nil {
		return x.ProposeBallot
	}
	return 0
}

func (x *ProposeRequest) GetProposeValue() *ReplicaBatch {
	if x != nil {
		return x.ProposeValue
	}
	return nil
}

func (x *ProposeRequest) GetPrepareRequestForFutureIndex() *PrepareRequest {
	if x != nil {
		return x.PrepareRequestForFutureIndex
	}
	return nil
}

func (x *ProposeRequest) GetSender() int64 {
	if x != nil {
		return x.Sender
	}
	return 0
}

type AcceptReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceNumber             int32         `protobuf:"varint,1,opt,name=instance_number,json=instanceNumber,proto3" json:"instance_number,omitempty"`
	Accept                     bool          `protobuf:"varint,2,opt,name=accept,proto3" json:"accept,omitempty"`
	AcceptBallot               int64         `protobuf:"varint,3,opt,name=accept_ballot,json=acceptBallot,proto3" json:"accept_ballot,omitempty"`
	Decided                    bool          `protobuf:"varint,4,opt,name=decided,proto3" json:"decided,omitempty"` // if the current instance is already decided
	DecidedValue               *ReplicaBatch `protobuf:"bytes,5,opt,name=decided_value,json=decidedValue,proto3" json:"decided_value,omitempty"`
	PromiseReplyForFutureIndex *PromiseReply `protobuf:"bytes,6,opt,name=promise_reply_for_future_index,json=promiseReplyForFutureIndex,proto3" json:"promise_reply_for_future_index,omitempty"`
	Sender                     int64         `protobuf:"varint,7,opt,name=sender,proto3" json:"sender,omitempty"`
}

func (x *AcceptReply) Reset() {
	*x = AcceptReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_definitions_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AcceptReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AcceptReply) ProtoMessage() {}

func (x *AcceptReply) ProtoReflect() protoreflect.Message {
	mi := &file_common_definitions_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AcceptReply.ProtoReflect.Descriptor instead.
func (*AcceptReply) Descriptor() ([]byte, []int) {
	return file_common_definitions_proto_rawDescGZIP(), []int{7}
}

func (x *AcceptReply) GetInstanceNumber() int32 {
	if x != nil {
		return x.InstanceNumber
	}
	return 0
}

func (x *AcceptReply) GetAccept() bool {
	if x != nil {
		return x.Accept
	}
	return false
}

func (x *AcceptReply) GetAcceptBallot() int64 {
	if x != nil {
		return x.AcceptBallot
	}
	return 0
}

func (x *AcceptReply) GetDecided() bool {
	if x != nil {
		return x.Decided
	}
	return false
}

func (x *AcceptReply) GetDecidedValue() *ReplicaBatch {
	if x != nil {
		return x.DecidedValue
	}
	return nil
}

func (x *AcceptReply) GetPromiseReplyForFutureIndex() *PromiseReply {
	if x != nil {
		return x.PromiseReplyForFutureIndex
	}
	return nil
}

func (x *AcceptReply) GetSender() int64 {
	if x != nil {
		return x.Sender
	}
	return 0
}

var File_common_definitions_proto protoreflect.FileDescriptor

var file_common_definitions_proto_rawDesc = []byte{
	0x0a, 0x18, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x2b, 0x0a, 0x0f, 0x53, 0x69,
	0x6e, 0x67, 0x6c, 0x65, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a,
	0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x22, 0x70, 0x0a, 0x0b, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x42, 0x61, 0x74, 0x63, 0x68, 0x12, 0x1b, 0x0a, 0x09, 0x75, 0x6e, 0x69, 0x71, 0x75, 0x65,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x6e, 0x69, 0x71, 0x75,
	0x65, 0x49, 0x64, 0x12, 0x2c, 0x0a, 0x08, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x53, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x4f, 0x70,
	0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x08, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x73, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x22, 0x55, 0x0a, 0x0c, 0x52, 0x65, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x42, 0x61, 0x74, 0x63, 0x68, 0x12, 0x1b, 0x0a, 0x09, 0x75, 0x6e, 0x69,
	0x71, 0x75, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x6e,
	0x69, 0x71, 0x75, 0x65, 0x49, 0x64, 0x12, 0x28, 0x0a, 0x08, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x08, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73,
	0x22, 0x48, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x6f, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x6f,
	0x74, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x22, 0x78, 0x0a, 0x0e, 0x50, 0x72,
	0x65, 0x70, 0x61, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x0f,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x25, 0x0a, 0x0e, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65,
	0x5f, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x70,
	0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x42, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x73, 0x65,
	0x6e, 0x64, 0x65, 0x72, 0x22, 0xda, 0x02, 0x0a, 0x0c, 0x50, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x27, 0x0a, 0x0f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x18,
	0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x07, 0x70, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x12, 0x30, 0x0a, 0x14, 0x6c, 0x61, 0x73, 0x74,
	0x5f, 0x70, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x64, 0x5f, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x12, 0x6c, 0x61, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x6d,
	0x69, 0x73, 0x65, 0x64, 0x42, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x12, 0x30, 0x0a, 0x14, 0x6c, 0x61,
	0x73, 0x74, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x5f, 0x62, 0x61, 0x6c, 0x6c,
	0x6f, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x12, 0x6c, 0x61, 0x73, 0x74, 0x41, 0x63,
	0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x42, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x12, 0x3d, 0x0a, 0x13,
	0x6c, 0x61, 0x73, 0x74, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x5f, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x52, 0x65, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x11, 0x6c, 0x61, 0x73, 0x74, 0x41, 0x63,
	0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x64,
	0x65, 0x63, 0x69, 0x64, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x64, 0x65,
	0x63, 0x69, 0x64, 0x65, 0x64, 0x12, 0x32, 0x0a, 0x0d, 0x64, 0x65, 0x63, 0x69, 0x64, 0x65, 0x64,
	0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x52,
	0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x0c, 0x64, 0x65, 0x63,
	0x69, 0x64, 0x65, 0x64, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e,
	0x64, 0x65, 0x72, 0x18, 0x08, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65,
	0x72, 0x22, 0x85, 0x02, 0x0a, 0x0e, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x69,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x25, 0x0a,
	0x0e, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x5f, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x42, 0x61,
	0x6c, 0x6c, 0x6f, 0x74, 0x12, 0x32, 0x0a, 0x0d, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x5f,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x52, 0x65,
	0x70, 0x6c, 0x69, 0x63, 0x61, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x0c, 0x70, 0x72, 0x6f, 0x70,
	0x6f, 0x73, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x57, 0x0a, 0x20, 0x70, 0x72, 0x65, 0x70,
	0x61, 0x72, 0x65, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x66, 0x6f, 0x72, 0x5f,
	0x66, 0x75, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x52, 0x1c, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x46, 0x6f, 0x72, 0x46, 0x75, 0x74, 0x75, 0x72, 0x65, 0x49, 0x6e, 0x64, 0x65,
	0x78, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x22, 0xac, 0x02, 0x0a, 0x0b, 0x41, 0x63,
	0x63, 0x65, 0x70, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x27, 0x0a, 0x0f, 0x69, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4e, 0x75, 0x6d, 0x62,
	0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x06, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x61, 0x63,
	0x63, 0x65, 0x70, 0x74, 0x5f, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x0c, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x42, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x12,
	0x18, 0x0a, 0x07, 0x64, 0x65, 0x63, 0x69, 0x64, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x64, 0x65, 0x63, 0x69, 0x64, 0x65, 0x64, 0x12, 0x32, 0x0a, 0x0d, 0x64, 0x65, 0x63,
	0x69, 0x64, 0x65, 0x64, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0d, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52,
	0x0c, 0x64, 0x65, 0x63, 0x69, 0x64, 0x65, 0x64, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x51, 0x0a,
	0x1e, 0x70, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x5f, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x5f, 0x66,
	0x6f, 0x72, 0x5f, 0x66, 0x75, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x50, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x52, 0x1a, 0x70, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x46, 0x6f, 0x72, 0x46, 0x75, 0x74, 0x75, 0x72, 0x65, 0x49, 0x6e, 0x64, 0x65, 0x78,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x42, 0x09, 0x5a, 0x07, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_common_definitions_proto_rawDescOnce sync.Once
	file_common_definitions_proto_rawDescData = file_common_definitions_proto_rawDesc
)

func file_common_definitions_proto_rawDescGZIP() []byte {
	file_common_definitions_proto_rawDescOnce.Do(func() {
		file_common_definitions_proto_rawDescData = protoimpl.X.CompressGZIP(file_common_definitions_proto_rawDescData)
	})
	return file_common_definitions_proto_rawDescData
}

var file_common_definitions_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_common_definitions_proto_goTypes = []any{
	(*SingleOperation)(nil), // 0: SingleOperation
	(*ClientBatch)(nil),     // 1: ClientBatch
	(*ReplicaBatch)(nil),    // 2: ReplicaBatch
	(*Status)(nil),          // 3: Status
	(*PrepareRequest)(nil),  // 4: PrepareRequest
	(*PromiseReply)(nil),    // 5: PromiseReply
	(*ProposeRequest)(nil),  // 6: ProposeRequest
	(*AcceptReply)(nil),     // 7: AcceptReply
}
var file_common_definitions_proto_depIdxs = []int32{
	0, // 0: ClientBatch.requests:type_name -> SingleOperation
	1, // 1: ReplicaBatch.requests:type_name -> ClientBatch
	2, // 2: PromiseReply.last_accepted_value:type_name -> ReplicaBatch
	2, // 3: PromiseReply.decided_value:type_name -> ReplicaBatch
	2, // 4: ProposeRequest.propose_value:type_name -> ReplicaBatch
	4, // 5: ProposeRequest.prepare_request_for_future_index:type_name -> PrepareRequest
	2, // 6: AcceptReply.decided_value:type_name -> ReplicaBatch
	5, // 7: AcceptReply.promise_reply_for_future_index:type_name -> PromiseReply
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_common_definitions_proto_init() }
func file_common_definitions_proto_init() {
	if File_common_definitions_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_common_definitions_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SingleOperation); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_definitions_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*ClientBatch); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_definitions_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*ReplicaBatch); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_definitions_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*Status); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_definitions_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*PrepareRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_definitions_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*PromiseReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_definitions_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*ProposeRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_definitions_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*AcceptReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_common_definitions_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_definitions_proto_goTypes,
		DependencyIndexes: file_common_definitions_proto_depIdxs,
		MessageInfos:      file_common_definitions_proto_msgTypes,
	}.Build()
	File_common_definitions_proto = out.File
	file_common_definitions_proto_rawDesc = nil
	file_common_definitions_proto_goTypes = nil
	file_common_definitions_proto_depIdxs = nil
}
