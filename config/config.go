package config

import "math/big"

// config of the client
var (
	TxInjectCount = 80000 // How much Txs to inject?
	BatchSize     = 4000  // client read a batch of txs and then send them once
	WaitTime      = 10    // Client wait the Txs to be processed by the blockchain system(seconds)
)

// config of the blockchain
var (
	Transaction_Method = string("")
	BlockSize          = 500
	ConsensusInterval  = 100                                                                         // (ms) the interval of the each round of consensus
	Init_Balance, _    = new(big.Int).SetString("100000000000000000000000000000000000000000000", 10) // A new coinbase Tx
	IPMap              = make(map[int]map[int]string)                                                // IPmap_nodeTable[shardID][nodeID] = "IP:Port"
	MeasureMethod      = []string{"avgTPS", "TCL", "TxNum"}                                          // the client measure method
	ConsensusMethod    = string("")                                                                  // the method of the consensus
	// Method          = string("")                                                                  // set by the command line
)

// config of the running environment
var (
	NodeNum       = 4         // total number of nodes in a shard
	ShardNum      = 4         // total number of shards
	ViewNodeId    = 0         // the nodeID of the initial view nodes
	LogLevel      = "INFO"    // default log level
	IsDistributed = false     // Running in local environment or not
	ClientShard   = 0xfffffff // the shardID of the client

	StoragePath = "./record"                                                      // the path to store the blockchain data
	ResultPath  = "./result/"                                                     // measurement data result output path
	LogPath     = "./log"                                                         // log output path
	CleintAddr  = "127.0.0.1:23333"                                               // client ip address
	FileInput   = `/home/pjj/Desktop/BlockChain/dataset/0to99999_Transaction.csv` // the BlockTransaction data path

	ServerAddrs = []string{ // for distribute experiment
		"192.168.0.1", "192.168.0.2", "192.168.0.3", "192.168.0.4",
		"192.168.0.5", "192.168.0.6", "192.168.0.7", "192.168.0.8",
		"192.168.0.9", "192.168.0.10", "192.168.0.11", "192.168.0.12",
		"192.168.0.13", "192.168.0.14", "192.168.0.15", "192.168.0.16",
		"192.168.0.17", "192.168.0.18", "192.168.0.19", "192.168.0.20",
		"192.168.0.21", "192.168.0.22", "192.168.0.23", "192.168.0.24",
		"192.168.0.25", "192.168.0.26", "192.168.0.27", "192.168.0.28",
		"192.168.0.29", "192.168.0.30", "192.168.0.31", "192.168.0.32",
		"192.168.0.33", "192.168.0.34", "192.168.0.35", "192.168.0.36",
		"192.168.0.37", "192.168.0.38", "192.168.0.39", "192.168.0.40",
		"192.168.0.41", "192.168.0.42", "192.168.0.43", "192.168.0.44",
		"192.168.0.45", "192.168.0.46", "192.168.0.47", "192.168.0.48",
		"192.168.0.49", "192.168.0.50", "192.168.0.51", "192.168.0.52",
		"192.168.0.53", "192.168.0.54", "192.168.0.55", "192.168.0.56",
		"192.168.0.57", "192.168.0.58", "192.168.0.59", "192.168.0.60",
		"192.168.0.61", "192.168.0.62", "192.168.0.63", "192.168.0.64",
		"192.168.0.65", "192.168.0.66", "192.168.0.67", "192.168.0.68",
		"192.168.0.69", "192.168.0.70", "192.168.0.71", "192.168.0.72",
		"192.168.0.73", "192.168.0.74", "192.168.0.75", "192.168.0.76",
		"192.168.0.77", "192.168.0.78", "192.168.0.79", "192.168.0.80",
	}
)
