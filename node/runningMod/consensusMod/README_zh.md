# 共识模块

本文件夹定义了一些共识通用的模块。每个子文件夹定义了一种共识算法

## /proposexx.go 
proposexxx.go定义主节点接受request，并向整个分片发送 propose 消息的行为。如：
- proposeBlock.go 接受client发送的交易并放入交易池，从交易池打包区块并发起 <propose 区块>
- proposeTxs.go 接受client发送的交易并放入交易池，并发起 <propose 交易>
- proposeString.go 接受client发送的字符串，并发起 <propsoe 字符串>

💥 **注意**：  
- proposexxx.go中的"Propose" 主要指的是将某个请求（或者在不同共识协议中有不同名称，但本质上是指需要达成共识的对象）广播到区块链（或区块链分片）上的过程。这个过程名称借鉴了 PBFT 协议中的 "Propose"。

- 关于共识轮次的控制（即在一轮共识没有结束的时候，不发起下一轮共识）: 
早期版本中由识模块控制，propose模块只管不停propose就行；但我认为由propose模块控制阻塞，共识模块调用`p2pMod.MsgHandlerMap[message.MsgConsensusDone]`触发下一轮的propose更合理，由于时间问题，仅在
proposeString.go 中实现该机制。

## /pbft/
定义PBFT共识协议，并预留PBFT分片的自定义处理接口

## /ds/
定义Dolev-Strong协议
