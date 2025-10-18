package main

import "BlockChainSimulator/node/plugins/plugininterface"

var PluginPackage = plugininterface.PluginPackage{
	PackageName: "sendTxs",
	Version:     "1.0.0",
	Plugins: map[string]plugininterface.PluginConstructor{
		"SendTxTest":           NewSendTxTestPlugin,
		"SendMimicAccountTxs":  NewSendMimicAccountTxsPlugin,
		"SendMimicContractTxs": NewSendMimicContractTxsPlugin,
	},
	Metadata: map[string]*plugininterface.PluginMetadata{
		"SendTxTest": {
			Name:         "SendTxTest",
			Description:  "用于测试发送交易的插件，每3秒发送一次交易",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryClient,
			Dependencies: []string{},
		},
		"SendMimicAccountTxs": {
			Name:         "SendMimicAccountTxs",
			Description:  "用于发送模拟账户交易的插件，通过config.TxInjectSpeed配置控制发送速度",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryClient,
			Dependencies: []string{},
		},
		"SendMimicContractTxs": {
			Name:         "SendMimicContractTxs",
			Description:  "用于发送模拟合约交易的插件，通过config.TxInjectSpeed配置控制发送速度",
			Version:      "1.0.0",
			Category:     plugininterface.PluginCategoryClient,
			Dependencies: []string{},
		},
	},
}
