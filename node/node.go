package node

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/pluginloader"
	"BlockChainSimulator/node/plugins/plugininterface"
	"BlockChainSimulator/utils"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Node struct {
	Attr    *nodeattr.NodeAttr       // the base attribute of the node
	P2PMod  *p2p.P2PMod              // the p2p network module
	Plugins []plugininterface.Plugin // 加载的插件列表
}

func NewNode(sid int, nid int, pcc *config.ChainConfig, pluginConfigs []pluginloader.PluginConfig) (*Node, error) {
	node := new(Node)
	node.Attr = nodeattr.NewNodeAttr(sid, nid, pcc)
	node.P2PMod = p2p.NewP2PMod(node.Attr.Ipaddr)

	loader := pluginloader.GetLoader()

	// 加载并创建插件实例
	for _, pluginCfg := range pluginConfigs {
		// 确保包已加载
		if err := loader.LoadPackage(pluginCfg.Package, config.PluginPath); err != nil {
			utils.LoggerInstance.Error("Failed to load package %s: %v", pluginCfg.Package, err)
			return nil, err
		}

		// 获取插件构造函数
		constructor, err := loader.GetPlugin(pluginCfg.Plugin)
		if err != nil {
			utils.LoggerInstance.Error("Failed to get plugin %s: %v", pluginCfg.Plugin, err)
			return nil, err
		}

		// 创建插件实例
		pluginInstance := constructor(node.Attr, node.P2PMod)
		node.Plugins = append(node.Plugins, pluginInstance)

		utils.LoggerInstance.Info("Created plugin instance: %s from package %s",
			pluginCfg.Plugin, pluginCfg.Package)
	}

	// 初始化所有插件
	for _, pluginInstance := range node.Plugins {
		pluginInstance.Initialize()
	}

	return node, nil
}

func (n *Node) Run() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	n.P2PMod.MsgHandlerMap[message.MsgStop] = func(msg *message.Message) {
		utils.LoggerInstance.Info("Received stop message...now closing plugins")
		cancel()
	}

	// 启动消息监听
	n.P2PMod.StartListen()

	// 启动所有插件
	for _, pluginInstance := range n.Plugins {
		wg.Add(1)
		go pluginInstance.Run(ctx, &wg)
	}

	// Ctrl+C 停止所有协程
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-sigChan:
		utils.LoggerInstance.Info("Node stopped by system interrupt...now closing plugins")
		cancel()
	}

	// 等待所有插件停止
	wg.Wait()

	// 清理插件资源
	for _, pluginInstance := range n.Plugins {
		pluginInstance.Cleanup()
	}

	utils.LoggerInstance.Info("Node stopped...")
}
