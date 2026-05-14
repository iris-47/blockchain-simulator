package main

import "BlockChainSimulator/node/plugins/plugininterface"

var PluginPackage = plugininterface.PluginPackage{
	PackageName: "stc",
	Version:     "1.0.0",
	Plugins: map[string]plugininterface.PluginConstructor{
		"STCClient": NewSTCClientPlugin,
		"STC":       NewSTCPlugin,
	},
	Metadata: map[string]*plugininterface.PluginMetadata{
		"STCClient": {
			Name:         "STCClient",
			Description:  "",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryClient,
			Dependencies: []string{},
		},
		"STC": {
			Name:         "STC",
			Description:  "",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryConsensus,
			Dependencies: []string{},
		},
	},
}
