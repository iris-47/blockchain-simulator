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
	NodeNum    = 4      // total number of nodes in a shard
	ShardNum   = 1      // total number of shards, default no sharding
	ViewNodeId = 0      // the nodeID of the initial view nodes
	LogLevel   = "INFO" // default log level

	MaliciousRatio float64 = 1. / 3    // the ratio of malicious nodes in the network
	ResilientRatio float64 = 1. / 2    // the ratio of resilient nodes in the network
	IsMalicious            = false     // whether this node is malicious
	IsDistributed          = false     // Running in local environment or not
	ClientShard            = 0xfffffff // the shardID of the client

	// used to synchronize the start time of the protocol in some protocol using synchronous network model like TBB and DS
	StartTimeWait int64 = 1000 // the start time of the protocol(ms)
	TickInterval  int64 = 1000 // the interval between each clock(ms)

	StoragePath = "./blockchain_data/"                                            // the path to store the blockchain data
	ResultPath  = "./result/"                                                     // measurement data result output path
	LogPath     = "./log/"                                                        // log output path
	StartPort   = 28800                                                           // the start port of the IPnodeTable, in local environment
	ClientAddr  = "127.0.0.1:23333"                                               // client ip address
	FileInput   = `/home/pjj/Desktop/BlockChain/dataset/0to99999_Transaction.csv` // the BlockTransaction data path

	ConnectRemoteDemo = false
	DemoServerURL     = "192.168.80.1:23333" // to send the log to the demo server, empty means not to send
)

// config of the distributed environment
var (
	NodePerServer = 40        // the number of nodes in a server
	ServerAddrs   = []string{ // for distribute experiment
		"192.168.0.1", "192.168.0.245", "192.168.0.251", "192.168.0.243",
		"192.168.0.246", "192.168.0.8", "192.168.0.250", "192.168.0.252",
		"192.168.0.249", "192.168.0.5", "192.168.0.244", "192.168.0.11",
		"192.168.0.247", "192.168.0.4", "192.168.0.6", "192.168.0.14",
		"192.168.0.2", "192.168.0.10", "192.168.0.12", "192.168.0.248",
		"192.168.0.15", "192.168.0.7", "192.168.0.3", "192.168.0.9",
		"192.168.0.13",
		// "192.168.0.1", "192.168.0.2", "192.168.0.3", "192.168.0.4",
		// "192.168.0.5", "192.168.0.6", "192.168.0.7", "192.168.0.8",
		// "192.168.0.9", "192.168.0.10", "192.168.0.11", "192.168.0.12",
		// "192.168.0.13", "192.168.0.14", "192.168.0.15", "192.168.0.16",
		// "192.168.0.17", "192.168.0.18", "192.168.0.19", "192.168.0.20",
		// "192.168.0.21", "192.168.0.22", "192.168.0.23", "192.168.0.24",
		// "192.168.0.25", "192.168.0.26", "192.168.0.27", "192.168.0.28",
		// "192.168.0.29", "192.168.0.30", "192.168.0.31", "192.168.0.32",
		// "192.168.0.33", "192.168.0.34", "192.168.0.35", "192.168.0.36",
		// "192.168.0.37", "192.168.0.38", "192.168.0.39", "192.168.0.40",
		// "192.168.0.41", "192.168.0.42", "192.168.0.43", "192.168.0.44",
		// "192.168.0.45", "192.168.0.46", "192.168.0.47", "192.168.0.48",
		// "192.168.0.49", "192.168.0.50", "192.168.0.51", "192.168.0.52",
		// "192.168.0.53", "192.168.0.54", "192.168.0.55", "192.168.0.56",
		// "192.168.0.57", "192.168.0.58", "192.168.0.59", "192.168.0.60",
		// "192.168.0.61", "192.168.0.62", "192.168.0.63", "192.168.0.64",
		// "192.168.0.65", "192.168.0.66", "192.168.0.67", "192.168.0.68",
		// "192.168.0.69", "192.168.0.70", "192.168.0.71", "192.168.0.72",
		// "192.168.0.73", "192.168.0.74", "192.168.0.75", "192.168.0.76",
		// "192.168.0.77", "192.168.0.78", "192.168.0.79", "192.168.0.80",
		// "127.0.0.1",
	}
)

// // The command line arguments
// type Args struct {
// 	// <-- Blockchain Config Related -->
// 	NodeID    int // id of this node, for example, 0
// 	NodeNum   int // indicate how many nodes of each shard are deployed
// 	ShardID   int // id of the shard to which this node belongs, for example, 0
// 	ShardNum  int // indicate that how many shards are deployed
// 	BlockSize int // how many Txs per block

//		// <-- Running Config Related -->
//		IsClient        bool   // whether this node is a client
//		IsDistribute    bool   // whether the environment is distribute or local
//		ConsensusMethod string // choice fo consensus Method, for example, CShard
//		TxType          string // choice of TxType, for example, UTXO
//		LogLevel string // Set the log level of [NONE, INFO, DBG]
//	}
//

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
