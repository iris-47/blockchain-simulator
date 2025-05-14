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

	ConnetRemoteDemo bool // whether the node is connected to the remote demo

	// <-- Client Config Related -->
	TxInjectCount int // how many txs to inject
	TxInjectSpeed int // how many txs to inject per second
}

type ExtConfig struct {
	ViewNodeId    *int      `json:"viewNodeId"`
	ClientShard   *int      `json:"clientShard"`
	StoragePath   *string   `json:"storagePath"`
	ResultPath    *string   `json:"resultPath"`
	LogPath       *string   `json:"logPath"`
	StartPort     *int      `json:"startPort"`
	ClientAddr    *string   `json:"clientAddr"`
	FileInput     *string   `json:"fileInput"`
	DemoServerURL *string   `json:"demoServerURL"`
	NodePerServer *int      `json:"nodePerServer"`
	ServerAddrs   *[]string `json:"serverAddrs"`
	StartTimeWait *int64    `json:"startTimeWait"`
	TickInterval  *int64    `json:"tickInterval"`
}

// Configuration for a Blockchain
type ChainConfig struct {
	NodeID   int // id of this node, for example, 0
	NodeNum  int // indicate how many nodes of each shard are deployed
	ShardID  int // id of the shard to which this node belongs, for example, 0
	ShardNum int // indicate that how many shards are deployed

	BlockSize int // how many Txs per block
}
