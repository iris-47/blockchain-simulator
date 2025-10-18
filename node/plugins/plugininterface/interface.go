package plugininterface

import (
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"context"
	"sync"
)

const (
	PluginCategoryConsensus = "consensus"
	PluginCategoryClient    = "client"
	PluginCategoryAuxiliary = "auxiliary"
	PluginCategoryGlobal    = "global"
)

// PluginMetadata 插件元数据
type PluginMetadata struct {
	Name         string   `json:"name"`         // 插件唯一标识名
	Description  string   `json:"description"`  // 插件描述
	Version      string   `json:"version"`      // 版本号
	Category     string   `json:"category"`     // 类别：consensus/client/auxiliary/global
	Dependencies []string `json:"dependencies"` // 依赖的其他插件
}

// Plugin 插件接口，用于定义网络中一个节点的行为
type Plugin interface {
	// 节点行为定义
	Initialize()                                 // 插件初始化（如：注册消息处理函数到 p2pMod）
	Run(ctx context.Context, wg *sync.WaitGroup) // 节点运行 （如：循环做某件事）
	Cleanup()                                    // 插件资源清理 （通常没用）
}

// ↑注：Plugin接口函数都没有error返回值，这个算历史遗留问题。目前插件的错误只能通过日志打印。

// PluginConstructor 插件构造函数类型
type PluginConstructor func(attr *nodeattr.NodeAttr, p2pMod *p2p.P2PMod) Plugin

// PluginPackage 插件包导出结构（每个 .so 文件导出此结构）
type PluginPackage struct {
	PackageName string                       // 包名（与 .so 文件名相同）
	Version     string                       // 包版本
	Plugins     map[string]PluginConstructor // 插件名 -> 构造函数
	Metadata    map[string]*PluginMetadata   // 插件名 -> 元数据
}
