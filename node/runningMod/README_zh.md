# 运行模块

本文件夹定义了节点的行为模块，主要分为三个子文件夹：`clientMod`、`consensusMod` 和 `auxiliaryMod`。这些文件夹分别定义了客户端节点、共识节点以及其他公共行为模块。

## 模块定义

所有节点行为模块需要符合以下接口定义：

```go
type RunningMod interface {
    RegisterHandlers()                           // 注册消息处理函数到 p2pMod
    Run(ctx context.Context, wg *sync.WaitGroup) // 节点初始化后会调用Run()，ctx 用于在节点关闭时执行必要的操作
}
```

## 添加新模块
新添加的节点模块需要在 `runningModRegister.go` 文件中进行注册。为模块取一个唯一的字符串名称，以便在初始化节点时能够方便地加入该模块。
`predefinedSolution.go` 文件中定义了一些常用的节点模块组合，用于快速初始化节点。这些预定义的组合可以帮助开发者快速搭建和测试节点。

## 待做
（2~3天）现有的机制每次更新源码需要重新编译，考虑使用plugins包，将这些模块变成动态链接库的形式，用起来更方便。但现在主库（网络等）仍然不稳定，时常有更新，故暂时不做。