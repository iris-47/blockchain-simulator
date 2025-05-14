// Internal configuration initialized via command-line arguments.
// Governs runtime behavior (e.g., node counts, consensus parameters) and experimental settings.
// Separated to allow dynamic tuning during execution and avoid hardcoding operational parameters (e.g., test scalability, malicious node ratios).
package config

import (
	"math/big"
	"strconv"
)

var (
	TxVerifyTime  = true // Estimate means just estimate the time, not really update the state
	ExecTimeDelay = 2    // mimic the execution time of a transaction
)

// config of the client
var (
	TxInjectCount = 80000 // How much Txs to inject?
	TxInjectSpeed = 4000  // How many Txs to inject per second?
	BatchSize     = 4000  // client read a batch of txs and then send them once
	WaitTime      = 10    // Client wait the Txs to be processed by the blockchain system(seconds)
)

// config of the blockchain
var (
	TxType            = string("")
	BlockSize         = 500
	ConsensusInterval = 100                                                                         // (ms) the interval of the each round of consensus
	Init_Balance, _   = new(big.Int).SetString("100000000000000000000000000000000000000000000", 10) // A new coinbase Tx
	IPMap             = make(map[int]map[int]string)                                                // IPmap_nodeTable[shardID][nodeID] = "IP:Port"
	MeasureMethod     = []string{"TPS", "TCL", "WaitLen"}                                           // the client measure method, must muanlly set at here
	ConsensusMethod   = string("")                                                                  // the method of the consensus, set through the command line
)

// config of the running environment
var (
	NodeNum  = 4 // total number of nodes in a shard
	ShardNum = 1 // total number of shards, default no sharding

	LogLevel = "INFO" // default log level

	MaliciousRatio float64 = 1. / 3 // the ratio of malicious nodes in the network
	ResilientRatio float64 = 1. / 2 // the ratio of resilient nodes in the network
	IsMalicious            = false  // whether this node is malicious
	IsDistributed          = false  // Running in local environment or not

	ConnectRemoteDemo = false
)

func InitConfig(args *Args) {
	NodeNum = args.NodeNum
	ShardNum = args.ShardNum
	BlockSize = args.BlockSize
	IsDistributed = args.IsDistribute
	ConsensusMethod = args.ConsensusMethod
	MaliciousRatio = args.MaliciousRatio
	ResilientRatio = args.ResilientRatio
	IsMalicious = args.IsMalicious
	ConnectRemoteDemo = args.ConnetRemoteDemo
	TxType = args.TxType
	LogLevel = args.LogLevel
	TxInjectCount = args.TxInjectCount
	TxInjectSpeed = args.TxInjectSpeed

	if args.IsClient {
		args.ShardID = ClientShard
		args.NodeID = 0
	}
	// init the IPMap
	if !IsDistributed {
		// local envieronment
		for i := 0; i < ShardNum; i++ {
			if _, ok := IPMap[i]; !ok {
				IPMap[i] = make(map[int]string)
			}
			for j := 0; j < NodeNum; j++ {
				// local environment, assume no more than 100 nodes in a shard
				IPMap[i][j] = "127.0.0.1:" + strconv.Itoa(StartPort+100*i+j)
			}
		}
	} else { // distributed environment
		IPList := make([]string, 0)
		for i := 0; i < len(ServerAddrs); i++ {
			for j := 0; j < NodePerServer; j++ {
				IPList = append(IPList, ServerAddrs[i]+":"+strconv.Itoa(StartPort+j))
			}
		}

		// distribute the IPs from IPList to IPMap
		for i := 0; i < ShardNum; i++ {
			if _, ok := IPMap[i]; !ok {
				IPMap[i] = make(map[int]string)
			}
			for j := 0; j < NodeNum; j++ {
				IPMap[i][j] = IPList[i*NodeNum+j]
			}
		}
	}
	IPMap[ClientShard] = make(map[int]string)
	IPMap[ClientShard][0] = ClientAddr
}

func SetDemoServerURL(url string) {
	DemoServerURL = url
}
