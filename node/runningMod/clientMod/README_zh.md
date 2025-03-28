# Client模块族

本文件夹为client节点行为的模块，包括发送交易到区块链；启动/关闭区块链节点；收集区块链数据并计算；

## measure 模块

### 简介

`measure` 模块用于客户端节点（client），它接受区块链运行时产生的数据并进行相关的测量和计算。

### 使用方法

1. 在 `config.MeasureMethod` 中定义所需的测量方法。
2. 对应的测量方法会在 `Run()` 方法中自动加载，并在随后的过程中执行这些测量。

### 新增测量方法

如果你需要添加一个新的测量方法（例如 `XXX`）：

1. 创建新的 `measureAddon_XXX.go` 文件，编写对应的测量方法实现。
2. 在 `measureAddon_interface.go` 文件中进行注册：
   - 添加新的模块名称（字符串）。
   - 注册其初始化函数。
3. 在 `config.MeasureMethod` 中通过模块名称引用该测量方法。

---

## startSystem 模块

### 简介

`startSystem` 模块是客户端节点（client）使用的，用于启动所有其他节点。此模块仅限于本地环境运行。

### 使用方法

在 `Run()` 方法中，构建并执行所有节点启动的命令行。命令行的参数由 `config` 包定义。注意，在使用startSystem模块的时候需要先执行 `go build`

当你按下 `CTRL + C` 停止客户端节点时，`startSystem` 模块会向所有已启动的节点发送停止消息，接收到消息的节点会随即停止运行。

---

## sendXXX 模块...

`sendXXX` 模块用于向区块链发送交易。其中：
`sendMimicContractTxs` 用于每隔一段时间向指定的 Shard 发送模拟的合约交易。
`sendTxTest` 用于每隔一段时间向指定的 Shard 发送固定的交易，用于测试。
`sendStringManual` 用于手动发送 string 到区块链，用于模拟真实的“客户端”交互。
