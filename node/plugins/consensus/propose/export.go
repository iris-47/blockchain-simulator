package main

import "BlockChainSimulator/node/plugins/plugininterface"

var PluginPackage = plugininterface.PluginPackage{

	PackageName: "propose",
	Version:     "1.0.0",
	Plugins: map[string]plugininterface.PluginConstructor{
		"ProposeTxs":   NewProposeTxsPlugin,
		"ProposeBlock": NewProposeBlockPlugin,
	},
	Metadata: map[string]*plugininterface.PluginMetadata{
		"ProposeTxs": {
			Name:         "ProposeTxs",
			Description:  "用于基于RBE的无人机项目，向区块链(分片)提议交易的插件",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryConsensus,
			Dependencies: []string{},
		},
		"ProposeBlock": {
			Name:         "ProposeBlock",
			Description:  "用于基于RBE的无人机项目，向区块链(分片)提议区块的插件",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryConsensus,
			Dependencies: []string{},
		},
	},
}
