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
	IsClient      bool   // whether this node is a client
	IsDistribute  bool   // whether the environment is distribute or local
	CommitteeType string // choice Committee Method, for example, CShard
	TxType        string // choice of TxType, for example, UTXO
	// UseUTXO      bool   // use utxo model or not
	LogLevel string // Set the log level of [NONE, INFO, DBG]

	// <-- Generate batch processing file-->
	GenBat       bool // generation bat for the whole system
	GenServerBat bool // generation bat for a server
}

// Configuration for a Blockchain
type ChainConfig struct {
	NodeID   int // id of this node, for example, 0
	NodeNum  int // indicate how many nodes of each shard are deployed
	ShardID  int // id of the shard to which this node belongs, for example, 0
	ShardNum int // indicate that how many shards are deployed

	BlockSize int // how many Txs per block
}