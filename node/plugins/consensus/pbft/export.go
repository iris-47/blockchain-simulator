package main

import "BlockChainSimulator/node/plugins/plugininterface"

var PluginPackage = plugininterface.PluginPackage{

	PackageName: "pbft",
	Version:     "1.0.0",
	Plugins: map[string]plugininterface.PluginConstructor{
		"PbftConsensus": NewPbftCosensusPlugin,
	},
	Metadata: map[string]*plugininterface.PluginMetadata{
		"PbftConsensus": {
			Name:         "PBFT Consensus",
			Description:  "用于实现PBFT共识算法的插件",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryConsensus,
			Dependencies: []string{},
		},
	},
}
