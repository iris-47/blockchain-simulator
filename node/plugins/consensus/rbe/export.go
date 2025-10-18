package main

import "BlockChainSimulator/node/plugins/plugininterface"

var PluginPackage = plugininterface.PluginPackage{

	PackageName: "rbe",
	Version:     "1.0.0",
	Plugins: map[string]plugininterface.PluginConstructor{
		"RBEMonitoring":   NewRBEMonitorPlugin,
		"RBEIdentityAuth": NewRBEIdentityAuthPlugin,
	},
	Metadata: map[string]*plugininterface.PluginMetadata{
		"RBEMonitoring": {
			Name:         "RBEMonitoring",
			Description:  "用于基于RBE的无人机项目，性能监控节点模块",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryClient,
			Dependencies: []string{},
		},
		"RBEIdentityAuth": {
			Name:         "RBEIdentityAuth",
			Description:  "用于基于RBE的无人机项目，身份认证插件",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryConsensus,
			Dependencies: []string{},
		},
	},
}
