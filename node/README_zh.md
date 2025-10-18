# 🧩 BlockChainSimulator 运行插件开发指南

本项目已经实现一个p2p网络节点框架，在此基础上通过 **Go 动态插件机制（`-buildmode=plugin`）** 实现可扩展的模块化区块链模拟系统。
每个插件独立编译为 `.so` 文件，在运行时由 `pluginloader` 动态加载。

本文档将指导你如何创建并注册一个新的插件。

---

## 📁 一、目录结构约定（非常重要）

插件源码必须严格放置在以下四个类别目录之一：

```
node/plugins/
├── consensus/   # 共识算法插件（如 PBFT、PoW、Raft 或其它运行逻辑等）
├── client/      # 客户端行为插件（如 节点控制、网页服务、钱包、用户行为、请求模拟）
├── auxiliary/   # 辅助类插件（如 监控、统计、可视化等）
└── global/      # 全局系统插件（暂时保留。最初设计是网络通信类插件，但既然大家都需要这个功能，所以直接加进主程序了）
```

每个插件需要实现以下接口：
```go
type Plugin interface {
	Initialize()                                 // 插件初始化（如：注册消息处理函数到 p2pMod）
	Run(ctx context.Context, wg *sync.WaitGroup) // 节点运行 （如：循环做某件事）
	Cleanup()                                    // 插件资源清理 （通常没用）
}
```

一个package应当放在一个文件夹下。一个package可以包含多个插件入口，称为一个`插件包`。每个插件包会被编译为一个`.so` 文件。每个插件包需要有如下导出结构体（推荐单独起一个`pluginpack_export.go`来定义插件包的导出结构体）：
```go
// PluginConstructor 插件构造函数类型
type PluginConstructor func(attr *nodeattr.NodeAttr, p2pMod *p2p.P2PMod) Plugin

// PluginPackage 插件包导出结构（每个 .so 文件导出此结构）
type PluginPackage struct {
	PackageName string                       // 包名（与 .so 文件名相同）
	Version     string                       // 包版本
	Plugins     map[string]PluginConstructor // 插件名 -> 构造函数
	Metadata    map[string]*PluginMetadata   // 插件名 -> 元数据
}
```

插件包必须单独形成文件夹，如consensus/pbft/pbft.go, 且代码包名、插件包名与文件夹名需要保持一致。


编译后，所有插件 `.so` 文件会生成到：

```
plugins_bin/
```

例如：

```
plugins_bin/pbft.so
plugins_bin/statistics.so
```

> ⚠️ **目录结构是 Makefile 的硬性要求**
> Makefile 会通过 `$(wildcard $(CONSENSUS_DIR)/*)` 自动查找子目录。
> 若插件未放在上述四类目录下，将无法被识别和编译！

---

## 二、快速创建一个新插件

以创建一个新的共识插件 `MyConsensus` 为例：

### 1️⃣ 新建源码目录

在 `node/plugins/consensus` 下创建子目录：

```bash
mkdir -p node/plugins/consensus/myconsensus
```

### 2️⃣ 编写插件代码

创建文件 `node/plugins/consensus/myconsensus/myconsensus.go`：

```go
package myconsensus

import (
    "BlockChainSimulator/node/nodeattr"
    "BlockChainSimulator/node/p2p"
    "BlockChainSimulator/node/plugins/plugininterface"
    "context"
    "sync"
    "fmt"
)

// 插件实例结构体（实现 plugininterface.Plugin 接口）
type MyConsensus struct {
    attr  *nodeattr.NodeAttr
    p2p   *p2p.P2PMod
}

// ---- 实现 Plugin 接口 ----
func (mc *MyConsensus) Initialize() {
    fmt.Println("[MyConsensus] Initialized")
}

func (mc *MyConsensus) Run(ctx context.Context, wg *sync.WaitGroup) {
    defer wg.Done()
    fmt.Println("[MyConsensus] Running ...")
    <-ctx.Done() // 等待关闭信号
    fmt.Println("[MyConsensus] Stopped")
}

func (mc *MyConsensus) Cleanup() {
    fmt.Println("[MyConsensus] Cleaned up")
}

// ---- 导出插件包结构 ----
var Package = plugininterface.PluginPackage{
    PackageName: "myconsensus",
    Version:     "v1.0.0",
    Plugins: map[string]plugininterface.PluginConstructor{
        "MyConsensus": func(attr *nodeattr.NodeAttr, p2pMod *p2p.P2PMod) plugininterface.Plugin {
            return &MyConsensus{attr: attr, p2p: p2pMod}
        },
    },
    Metadata: map[string]*plugininterface.PluginMetadata{
        "MyConsensus": {
            Name:        "MyConsensus",
            Description: "An example consensus plugin",
            Version:     "v1.0.0",
            Category:    "consensus",
            Dependencies: []string{},
        },
    },
}
```

> ✅ 关键点：
>
> * 导出变量名必须是 **`Package`**（大小写敏感），且类型为 `plugininterface.PluginPackage`。
> * `PackageName` 必须与目录名和最终 `.so` 文件名一致（例如 `myconsensus`）。
> * 所有插件均需实现 `plugininterface.Plugin` 接口的三个方法：
>
>   * `Initialize()`
>   * `Run(ctx context.Context, wg *sync.WaitGroup)`
>   * `Cleanup()`

---

##  三、编译插件

### 编译所有插件

```bash
make plugins
```

### 或仅编译单个插件

```bash
make plugin PKG=myconsensus
```

生成结果：

```
plugins_bin/myconsensus.so
```

> 如果路径或包名错误，Makefile 将提示：
>
> ```
> Plugin package myconsensus not found
> ```

---

##  四、插件加载机制简介

运行时，`node/node.go` 会：

1. 读取配置文件中声明的插件；
2. 调用 `pluginloader.LoadPackage(packageName, config.PluginPath)` 动态加载 `.so`；
3. 通过 `loader.GetPlugin(pluginName)` 获取构造函数；
4. 实例化插件并调用其：

   * `Initialize()` → 注册网络消息、初始化状态；
   * `Run()` → 在独立 goroutine 中执行逻辑；
   * `Cleanup()` → 释放资源（停止时调用）。

---

## ⚠️ 五、常见问题与注意事项

| 问题                           | 原因                                   | 解决方法                                                       |
| ---------------------------- | ------------------------------------ | ---------------------------------------------------------- |
| ❌ 插件未被编译                     | 目录不在四类路径下                            | 确保插件放在 `consensus/client/auxiliary/global` 下               |
| ❌ `.so` 加载失败                 | `Package.PackageName` 与 `.so` 文件名不一致 | 确保两者完全相同                                                   |
| ❌ “symbol Package not found” | 未导出 `var Package` 变量                 | 确保文件中存在：`var Package = plugininterface.PluginPackage{...}` |
| ⚠️ 插件运行但无输出                  | 未实现 `Run()` 或忘记启动 goroutine          | 检查是否正确实现接口方法                                               |
| ⚠️ 编译错误                      | 使用了主程序未导出的内部包                        | 插件仅能依赖 `node/`, `utils/`, `config/` 等公开模块                  |

---

## 🧩 六、开发建议

* **日志输出**请使用统一的 `utils.LoggerInstance`。
* 插件运行逻辑需支持 `context.Context` 停止信号。
* 若插件依赖其他插件，请在 `Metadata.Dependencies` 中声明。
* 测试可使用：

  ```bash
  make test
  ```
* 不建议在插件中直接使用 `os.Exit()`，会导致整个系统退出。

---

## ✅ 七、示例插件

你可以参考以下现有插件了解结构：

| 类别    | 示例路径                                |
| ----- | ----------------------------------- |
| 共识插件  | `node/plugins/consensus/pbft/`      |
| 客户端插件 | `node/plugins/client/runsystem/` |


---

> ✨ **提示**：所有 `.so` 文件都会被动态加载，编译完成后无需修改主程序。
> 这意味着你可以独立开发、热插拔测试插件，而无需重构核心代码。

插件无法互相调用，如果你需要，请在插件对应的包中开发你的包，并跟随插件进行导出。或者基于该插件进行开发，参考/node/plugins/consensus/pbft