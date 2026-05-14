package main

import "BlockChainSimulator/message"

const (
	MsgSTCConsensus message.MessageType = iota + 1000
	MsgSTCControl
	MsgSTCByzantineBlock
)

type SpaceTimeEnvelope struct {
	ShardID   int
	NodeID    int
	Timestamp int64
	Location  string
}

type STCConsensusPayload struct {
	Phase     string
	Height    int64
	Block     []byte
	BlockHash []byte
	Envelope  SpaceTimeEnvelope
}

type STCControlPayload struct {
	Action        string
	Enabled       bool
	TimeOffsetSec int64
	FakeLocation  string
	TargetTPS     int64
	DurationSec   int64
}

type STCQuery struct {
	RequestID   string
	Action      string
	StartHeight int64
	EndHeight   int64
}

type STCBlockHashItem struct {
	Height int64
	Hash   string
}

type STCMetrics struct {
	CurrentTPS    float64
	AverageTPS    float64
	AverageDelayS float64
	CommittedTxs  int64
	CurrentHeight int64
}

type STCQueryReply struct {
	RequestID   string
	Action      string
	ShardID     int
	NodeID      int
	Online      bool
	Metrics     STCMetrics
	BlockHashes []STCBlockHashItem
	Events      []string
}
