package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/node"
	"BlockChainSimulator/utils"
	"fmt"

	"github.com/spf13/pflag"
)

func main() {
	args := config.Args{}
	// <-- Blockchain Config Related -->
	blockchainFlags := pflag.NewFlagSet("Blockchain Config Related", pflag.ExitOnError)
	blockchainFlags.IntVarP(&args.NodeID, "nodeID", "n", 0, "id of this node, for example, 0")
	blockchainFlags.IntVarP(&args.NodeNum, "nodeNum", "N", 4, "indicate how many nodes of each shard are deployed")
	blockchainFlags.IntVarP(&args.ShardID, "shardID", "s", 0, "id of the shard to which this node belongs, for example, 0")
	blockchainFlags.IntVarP(&args.ShardNum, "shardNum", "S", 1, "indicate that how many shards are deployed")
	blockchainFlags.IntVarP(&args.BlockSize, "blockSize", "b", 500, "how many Txs per block")
	// <-- Running Config Related -->
	runningFlags := pflag.NewFlagSet("Running Config Related", pflag.ExitOnError)
	runningFlags.BoolVarP(&args.IsClient, "isClient", "c", false, "whether this node is a client")
	runningFlags.BoolVarP(&args.IsDistribute, "isDistribute", "d", false, "whether the environment is distribute or local")

	runningFlags.Float64VarP(&args.MaliciousRatio, "maliciousRatio", "r", 0, "the ratio of malicious nodes in the network")
	runningFlags.Float64VarP(&args.ResilientRatio, "resilientRatio", "R", 0.5, "the ratio of resilient nodes in the network")
	runningFlags.BoolVarP(&args.IsMalicious, "isMalicious", "M", false, "whether this node is malicious")
	runningFlags.StringVarP(&args.ConsensusMethod, "consensusMethod", "m", "Monoxide", "choice fo consensus Method, for example, Monoxide")
	runningFlags.StringVarP(&args.TxType, "txType", "t", "UTXO", "choice of TxType, for example, UTXO")
	runningFlags.StringVarP(&args.LogLevel, "logLevel", "l", "INFO", "Set the log level of [DEBUG, INFO, WARN, ERROR]")
	// <-- Client Config Related -->
	clientFlags := pflag.NewFlagSet("Client Config Related", pflag.ExitOnError)
	clientFlags.IntVarP(&args.TxInjectCount, "txInjectCount", "i", 80000, "how many txs to inject")
	clientFlags.IntVarP(&args.TxInjectSpeed, "txInjectSpeed", "p", 100, "how many txs to inject per second")

	pflag.CommandLine.AddFlagSet(blockchainFlags)
	pflag.CommandLine.AddFlagSet(runningFlags)
	pflag.CommandLine.AddFlagSet(clientFlags)

	pflag.Usage = func() {
		fmt.Println("Usage of application:")
		fmt.Println("\nBlockchain Config Related:")
		blockchainFlags.PrintDefaults()

		fmt.Println("\nRunning Config Related:")
		runningFlags.PrintDefaults()

		fmt.Println("\nClient Config Related:")
		clientFlags.PrintDefaults()
	}
	pflag.Parse()

	config.InitConfig(&args)
	utils.LoggerInstance, _ = utils.NewLogger(&args, args.LogLevel, true, false)

	pcc := config.ChainConfig{
		NodeID:    args.NodeID,
		NodeNum:   args.NodeNum,
		ShardID:   args.ShardID,
		ShardNum:  args.ShardNum,
		BlockSize: args.BlockSize,
	}

	var runningNode *node.Node
	var err error

	// choose the running mods
	if _, ok := PredefinedProtocolMods[args.ConsensusMethod]; !ok {
		utils.LoggerInstance.Error("Method %v is not supported", args.ConsensusMethod)
		return
	}

	if args.IsClient {
		utils.LoggerInstance.Debug("This node is a client")
		runningNode, err = node.NewNode(config.ClientShard, 0, &pcc, PredefinedProtocolMods[args.ConsensusMethod].clientMods)
	} else if args.NodeID == config.ViewNodeId {
		utils.LoggerInstance.Debug("This node is a view node")
		runningNode, err = node.NewNode(args.ShardID, args.NodeID, &pcc, PredefinedProtocolMods[args.ConsensusMethod].viewNodeMods)
	} else {
		utils.LoggerInstance.Debug("This node is a normal node")
		runningNode, err = node.NewNode(args.ShardID, args.NodeID, &pcc, PredefinedProtocolMods[args.ConsensusMethod].nodeMods)
	}

	if err != nil {
		utils.LoggerInstance.Error("Error creating node: %v", err)
		return
	}

	if runningNode == nil {
		utils.LoggerInstance.Error("runningNode is nil")
		return
	}

	runningNode.Run()
}
