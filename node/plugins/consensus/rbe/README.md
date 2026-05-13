无人机身份认证模块：**RBEIdentityAuthMod**。
该模块基于论文 *Efficient Registration-Based Encryption (Glaeser et al., CCS 2023)* 的 Registration-Based Encryption（RBE）算法

---

### 一、RBE算法背景与逻辑概述

RBE 是一种介于 **身份基加密 (IBE)** 与 **公钥基础设施 (PKI)** 之间的新型加密体系，消除了传统 IBE 的“密钥托管（Key Escrow）”问题。
在 RBE 中：

* 用户自己生成密钥；
* KC（Key Curator）负责注册与更新公共参数；
* KC 不掌握任何私钥，仅维护系统状态；
* 用户通过公钥注册后，即可参与加密通信或身份认证。

系统中的核心算法包括：

| 算法                              | 执行方   | 功能                                   |
| ------------------------------- | ----- | ------------------------------------ |
| **Setup(1λ, N)**                | 系统初始化 | 生成公共参考字符串 (CRS) 与公共参数结构。             |
| **Gen(crs)**                    | 无人机节点 | 本地生成密钥对 `(SK_i, PK_i)` 以及辅助数据 `ξ_i`。 |
| **Reg(crs, pp, id, PK_i, ξ_i)** | KC    | 验证注册请求并更新公共参数 `pp`、生成开口信息 `Λ_i`。     |
| **Enc(crs, pp, id, M)**         | 节点    | 基于 RBE 算法对消息 M 生成签名（或加密）σ。           |
| **Upd(pp, id)**                 | KC    | 返回指定身份的最新开口信息。                       |
| **Dec(SK_i, Λ_i, σ)**           | 节点    | 使用私钥与开口信息验证或解密签名σ。                   |

---

### 二、系统运行流程

整个无人机系统由 `N` 个节点组成，分为 `S` 个分片，每个分片有一个 KC，为简单起见，默认id为0的节点为KC。各分片并行执行身份认证。系统结构是共有多个分片，每个分片多个节点，每个节点都是从0开始编号ID，kc的id为0

#### （1）系统初始化

* 生成公共参考字符串 `crs`；
* 为每个分片初始化 KC，KC 维护：

  * 公共参数 `pp`
  * 辅助数据 `aux`
  * 注册表 `{DID → PK}`；
* 启动 DID 注册与 RBE 注册进程。

#### （2）无人机节点注册

* 节点执行 `Gen(crs)` → 得到 `(SK_i, PK_i, ξ_i)`；
* 注册 DID；
* 向所在分片 KC 提交 `(DID, PK_i, ξ_i)`；
* KC 执行 `Reg(crs, pp, DID, PK_i, ξ_i)` 验证并更新；
* KC 生成开口信息 `Λ_i` ，上链并返回；
* 节点持久化存储 `(SK_i, PK_i, DID, Λ_i)`。

#### （3）节点签名

* 发送方节点B构造消息 `M`；
* 执行 `Enc(crs, pp, DID_B, M)` → 生成签名σ；
* 向目标节点A发送 `{M, σ}`。

#### （4）身份验证

* 节点A接收 `{M, σ}`；
* 步骤1：向KC验证 `DID_B` 是否已注册；
* 步骤2：使用 `pp` 与 `Λ_B` 验证σ与M对应；
* 验证通过则确认B身份。

> 各分片独立执行，KC 之间无交叉管理。
> 节点的开口信息 `Λ_i` 存储在区块链上，使用时从区块链请求数据（未实现）
> 现版本节点的开口信息 `Λ_i` 本地持久化，不依赖动态请求。

---

### 三、通信模式说明

* **节点间通信：**
  若仅涉及节点 A 与节点 B，则通过点对点消息（A→B）传输；
* **分片内通信：**
  若KC需同步更新，可使用 `p2pMod.ConnManager.Broadcast()` 进行广播；
* **跨分片通信：**
  不在本模块范围内，可忽略。

---

### 四、RBEIdentityAuthMod 模块开发任务

模块名称：`RBEIdentityAuthMod`
模块职责：在现有框架下，实现基于RBE算法的身份注册、签名与认证逻辑。
该模块需实现接口：

```go
type RunningMod interface {
    RegisterHandlers()
    Run(ctx context.Context, wg *sync.WaitGroup)
}
```

---

#### 模块功能点

1. **节点注册**

   * 监听 `MsgRBERegister` 消息；
   * 执行注册逻辑并与KC同步；
   * KC在分片内广播注册结果；
   * 注册完成后在节点本地保存 `SK, PK, DID, Λ`。

2. **节点签名**

   * 监听 `MsgRBESign` 消息；
   * 使用伪函数 `Enc()` 模拟签名过程；
   * 将 `{M, σ}` 发送给目标节点。

3. **身份验证**

   * 监听 `MsgRBEVerify` 消息；
   * 向KC检查目标DID是否存在；
   * 使用伪函数 `Dec()` 验证签名；
   * 验证成功后发送 `MsgRBEReply`。

4. **周期性任务**

   * 在 `Run()` 中循环执行注册、验证或签名任务；
   * 响应 `ctx.Done()` 优雅退出。

---

### 五、伪函数约定（不实现底层密码学，原因：RBE 基于 双线性群（Bilinear Groups）与向量承诺（Vector Commitments），目前go语言没有很好的现成的库实现，考虑以后再实现，现在先用伪函数保证运行）

目前可能使用的库有：

// github.com/cloudflare/bn256 (提供pairing，但非完整RBE)；

// github.com/fentec-project/bn256 (更通用，但性能较低)；

// 部分科研实现（通常为C或Python版本）。


```go
func Enc(crs, pp, DID, M string) (sigma string)
func Dec(SK, Λ, sigma, M string) bool
func Reg(crs, pp, DID, PK, ξ string) (Λ string)
func Upd(pp, DID string) (Λ string)
```

---

### 六、性能
12个节点，占用64% CPU， 10M内存。性能开销主要在CPU，并能吃的比较满

### 七、待做
1. （1天）BUG: KC节点发送消息到其它节点会出现连接错误
2. （3~4天）分布式节点部署。将系统部署到云服务器上并进行测试
3. （3~4天，优先级较低）实际接入区块链。连接到区块链并调试可能需要1天左右。但区块链网络同样涉及部署到云服务器的过程，比较耗时。目前的实现是本地缓存。
4. （暂无法估计耗时天，优先级很低）底层密码学实现。
