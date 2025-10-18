package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/node"
	"BlockChainSimulator/node/pluginloader"
	"BlockChainSimulator/utils"
	"fmt"

	"github.com/spf13/pflag"
)

func main() {
	args := config.Args{}
	// <-- Blockchain Config Related -->
	blockchainFlags := pflag.NewFlagSet("Blockchain Config Related", pflag.ExitOnError)
	blockchainFlags.IntVarP(&args.NodeID, "nodeID", "n", 0, "id of this node")
	blockchainFlags.IntVarP(&args.NodeNum, "nodeNum", "N", 4, "nodes per shard")
	blockchainFlags.IntVarP(&args.ShardID, "shardID", "s", 0, "shard id of this node")
	blockchainFlags.IntVarP(&args.ShardNum, "shardNum", "S", 1, "total number of shards")
	blockchainFlags.IntVarP(&args.BlockSize, "blockSize", "b", 500, "Txs per block")

	// <-- Running Config Related -->
	runningFlags := pflag.NewFlagSet("Running Config Related", pflag.ExitOnError)
	runningFlags.BoolVarP(&args.IsClient, "isClient", "c", false, "whether this node is a client")
	runningFlags.BoolVarP(&args.IsDistribute, "isDistribute", "d", false, "whether the environment is distribute or local")
	runningFlags.Float64VarP(&args.MaliciousRatio, "maliciousRatio", "r", 0, "the ratio of malicious nodes in the network")
	runningFlags.Float64VarP(&args.ResilientRatio, "resilientRatio", "R", 0.5, "the ratio of resilient nodes in the network")
	runningFlags.BoolVarP(&args.IsMalicious, "isMalicious", "M", false, "whether this node is malicious")
	runningFlags.StringVarP(&args.ConsensusMethod, "consensusMethod", "m", "Simple", "choice of consensus Method")
	runningFlags.StringVarP(&args.TxType, "txType", "t", "UTXO", "choice of TxType")
	runningFlags.StringVarP(&args.LogLevel, "logLevel", "l", "INFO", "Set the log level")
	runningFlags.BoolVarP(&args.ConnetRemoteDemo, "connetRemoteDemo", "C", false, "whether the node is connected to the remote demo")

	// <-- Client Config Related -->
	clientFlags := pflag.NewFlagSet("Client Config Related", pflag.ExitOnError)
	clientFlags.IntVarP(&args.TxInjectCount, "txInjectCount", "i", 80000, "how many txs to inject")
	clientFlags.IntVarP(&args.TxInjectSpeed, "txInjectSpeed", "p", 4000, "how many txs to inject per second")

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
	config.LoadConfig()
	config.InitConfig(&args)
	utils.LoggerInstance, _ = utils.NewLogger(&args, args.LogLevel, true, true)
	pcc := config.ChainConfig{
		NodeID:    args.NodeID,
		NodeNum:   args.NodeNum,
		ShardID:   args.ShardID,
		ShardNum:  args.ShardNum,
		BlockSize: args.BlockSize,
	}
	// 加载协议配置
	protocolsConfig, err := pluginloader.LoadProtocolsConfig("protocols.json")
	if err != nil {
		utils.LoggerInstance.Error("Failed to load protocols config: %v", err)
		return
	}
	// 检查协议是否存在
	if _, exists := protocolsConfig.Protocols[args.ConsensusMethod]; !exists {
		utils.LoggerInstance.Error("Protocol %s not found in config", args.ConsensusMethod)
		return
	}
	// 预加载该协议需要的所有包
	packages, err := protocolsConfig.GetRequiredPackages(args.ConsensusMethod)
	if err != nil {
		utils.LoggerInstance.Error("Failed to get required packages: %v", err)
		return
	}
	loader := pluginloader.GetLoader()
	if err := loader.LoadPackages(packages, config.PluginPath); err != nil {
		utils.LoggerInstance.Error("Failed to load plugin packages: %v", err)
		return
	}
	var runningNode *node.Node
	var pluginConfigs []pluginloader.PluginConfig
	// 根据节点类型选择插件配置
	if args.IsClient {
		utils.LoggerInstance.Debug("This node is a client")
		pluginConfigs, err = protocolsConfig.GetPluginsForNode(args.ConsensusMethod, "client")
	} else if args.NodeID == config.ViewNodeId {
		utils.LoggerInstance.Debug("This node is a view node")
		pluginConfigs, err = protocolsConfig.GetPluginsForNode(args.ConsensusMethod, "view")
	} else {
		utils.LoggerInstance.Debug("This node is a normal node")
		pluginConfigs, err = protocolsConfig.GetPluginsForNode(args.ConsensusMethod, "normal")
	}
	if err != nil {
		utils.LoggerInstance.Error("Failed to get plugin configs: %v", err)
		return
	}
	// 创建节点
	runningNode, err = node.NewNode(args.ShardID, args.NodeID, &pcc, pluginConfigs)
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
