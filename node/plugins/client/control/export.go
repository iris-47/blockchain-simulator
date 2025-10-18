package main

import "BlockChainSimulator/node/plugins/plugininterface"

var PluginPackage = plugininterface.PluginPackage{
	PackageName: "control",
	Version:     "1.0.0",
	Plugins: map[string]plugininterface.PluginConstructor{
		"StartSystem": NewStartSystemPlugin,
		"StopSystem":  NewStopSystemPlugin,
	},
	Metadata: map[string]*plugininterface.PluginMetadata{
		"StartSystem": {
			Name:         "StartSystem",
			Description:  "用于启动整个区块链系统的插件，通过ssh启动各个节点",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryClient,
			Dependencies: []string{},
		},
		"StopSystem": {
			Name:         "StopSystem",
			Description:  "用于停止整个区块链系统的插件，通过发送Stop消息停止各个节点，无法强行终止故障节点",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryClient,
			Dependencies: []string{},
		},
	},
}
