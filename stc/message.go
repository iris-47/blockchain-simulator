package main

import "BlockChainSimulator/message"

const (
	MsgSTCConsensus message.MessageType = iota + 2000
	MsgSTCControl
	MsgSTCQuery
	MsgSTCQueryReply
	MsgSTCMetricsReport
	MsgSTCForwardTx
)

type STCEnvelope struct {
	ShardID   int
	NodeID    int
	Timestamp int64
	Location  string
}

type STCTransaction struct {
	ID             string
	Payload        string
	SubmitUnixNano int64
	FromShard      int
}

type STCInjectRequest struct {
	Txs []STCTransaction
}

type STCBlock struct {
	Height     int64
	PrevHash   string
	Hash       string
	TxIDs      []string
	Timestamp  int64
	ShardID    int
	ProposerID int
}

type STCConsensusPacket struct {
	Envelope STCEnvelope
	Type     string
	Block    STCBlock
}

type STCControlRequest struct {
	Action       string
	TargetShard  int
	TargetNode   int
	Enabled      bool
	Behavior     string
	FakeLocation string
}

type STCQueryRequest struct {
	RequestID   string
	Action      string
	StartHeight int64
	EndHeight   int64
	ReplyTo     string
}

type STCBlockHashRecord struct {
	Height int64
	Hash   string
}

type STCLogEvent struct {
	Level     string
	Type      string
	Timestamp int64
	NodeID    int
	ShardID   int
	Detail    string
}

type STCMetricsSnapshot struct {
	ShardID             int
	NodeID              int
	TPS                 float64
	AvgConfirmLatencyMs float64
	TotalCommittedTx    int64
	PendingTx           int
	LatestHeight        int64
	LastUpdatedUnixNano int64
}

type STCNodeStatus struct {
	ShardID    int
	NodeID     int
	Online     bool
	Role       string
	Malicious  bool
	Behavior   string
	Location   string
	IP         string
	LatestHash string
	Metrics    STCMetricsSnapshot
	Logs       []STCLogEvent
}

type STCQueryReply struct {
	RequestID string
	Action    string
	OK        bool
	Error     string
	Status    *STCNodeStatus
	Blocks    []STCBlockHashRecord
	Metrics   *STCMetricsSnapshot
}

type STCMetricsReport struct {
	Envelope STCEnvelope
	Metrics  STCMetricsSnapshot
}
