package config

type Address = string

// The command line arguments
type Args struct {
	// <-- Blockchain Config Related -->
	NodeID    int // id of this node, for example, 0
	NodeNum   int // indicate how many nodes of each shard are deployed
	ShardID   int // id of the shard to which this node belongs, for example, 0
	ShardNum  int // indicate that how many shards are deployed
	BlockSize int // how many Txs per block

	// <-- Running Config Related -->
	IsClient     bool // whether this node is a client
	IsDistribute bool // whether the environment is distribute or local

	MaliciousRatio float64 // the ratio of malicious nodes in the network
	ResilientRatio float64 // the ratio of resilient nodes in the network
	IsMalicious    bool    // whether this node is malicious

	ConsensusMethod string // choice fo consensus Method, for example, CShard
	TxType          string // choice of TxType, for example, UTXO
	LogLevel        string // Set the log level of [DEBUG, INFO, WARN, ERROR]

	// <-- Client Config Related -->
	TxInjectCount int // how many txs to inject
	TxInjectSpeed int // how many txs to inject per second
}

// Configuration for a Blockchain
type ChainConfig struct {
	NodeID   int // id of this node, for example, 0
	NodeNum  int // indicate how many nodes of each shard are deployed
	ShardID  int // id of the shard to which this node belongs, for example, 0
	ShardNum int // indicate that how many shards are deployed

	BlockSize int // how many Txs per block
}
