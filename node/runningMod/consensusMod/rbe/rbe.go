// rbe_identity_auth_mod.go
package rbe

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/utils"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ==================== RBE模块主结构 ====================

type RBEIdentityAuthMod struct {
	nodeAttr *nodeattr.NodeAttr
	p2pMod   *p2p.P2PMod

	// RBE系统参数
	crs string // 公共参考字符串
	pp  string // 公共参数
	aux string // 辅助数据（KC使用）

	// 节点密钥和身份
	SK  string // 私钥
	PK  string // 公钥
	DID string // 去中心化身份
	Λ   string // 开口信息(Lambda)
	ξ   string // 辅助数据(Ksi)

	// KC相关（仅KC节点使用）
	isKC         bool
	registryLock sync.RWMutex
	registry     map[string]string // DID -> PK映射
	openingInfo  map[string]string // DID -> Λ映射

	// 注册状态
	registered   bool
	registerLock sync.Mutex

	// 性能统计
	verifyCount      int64     // 原子操作，当前周期验证次数（用于计算瞬时TPS）
	totalVerifyCount int64     // 原子操作，历史累计验证总次数
	startTime        time.Time // 系统开始时间
	metricsLock      sync.RWMutex
	lastReportTime   time.Time

	// 待验证请求缓存（用于异步验证流程）
	pendingVerify     map[string]*VerifyRequest
	pendingVerifyLock sync.Mutex
}

// ==================== 构造函数 ====================

func NewRBEIdentityAuthMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	mod := &RBEIdentityAuthMod{
		nodeAttr:         attr,
		p2pMod:           p2p,
		registry:         make(map[string]string),
		openingInfo:      make(map[string]string),
		pendingVerify:    make(map[string]*VerifyRequest),
		isKC:             attr.Nid == 0, // ID为0的节点是KC
		registered:       false,
		verifyCount:      0,
		totalVerifyCount: 0,
	}

	// 初始化CRS（所有节点共享）
	mod.crs = Setup()

	// 如果是KC节点，初始化公共参数
	if mod.isKC {
		mod.pp = fmt.Sprintf("pp_shard_%d", attr.Sid)
		mod.aux = fmt.Sprintf("aux_shard_%d", attr.Sid)
		utils.LoggerInstance.Info("节点 [分片%d, 节点%d] 初始化为KC（Key Curator）", attr.Sid, attr.Nid)
	}

	return mod
}

// ==================== RBE伪函数实现 ====================
// RBE 基于 双线性群（Bilinear Groups）与向量承诺（Vector Commitments），标准 Go 语言库中并没有对应实现。
// 目前可能使用的库有：
// github.com/cloudflare/bn256 (提供pairing，但非完整RBE）；
// github.com/fentec-project/bn256 (更通用，但性能较低）；
// 部分科研实现（通常为C或Python版本）。

// 无现成RBE支持。
// Setup 生成公共参考字符串
func Setup() string {
	data := fmt.Sprintf("crs_%d", time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// Gen 生成密钥对和辅助数据
func (mod *RBEIdentityAuthMod) Gen() {
	timestamp := time.Now().UnixNano()
	baseData := fmt.Sprintf("node_%d_%d_%d", mod.nodeAttr.Sid, mod.nodeAttr.Nid, timestamp)

	// 生成私钥
	skHash := sha256.Sum256([]byte(baseData + "_sk"))
	mod.SK = hex.EncodeToString(skHash[:16])

	// 生成公钥
	pkHash := sha256.Sum256([]byte(baseData + "_pk"))
	mod.PK = hex.EncodeToString(pkHash[:16])

	// 生成辅助数据
	xiHash := sha256.Sum256([]byte(baseData + "_xi"))
	mod.ξ = hex.EncodeToString(xiHash[:16])

	// 生成DID
	mod.DID = fmt.Sprintf("did_s%d_n%d", mod.nodeAttr.Sid, mod.nodeAttr.Nid)

	utils.LoggerInstance.Info("节点 [分片%d, 节点%d] 生成密钥对，DID: %s",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid, mod.DID)
}

// Enc 模拟加密/签名过程
func Enc(crs, pp, did, m string) string {
	data := fmt.Sprintf("%s|%s|%s|%s", crs, pp, did, m)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Dec 模拟解密/验证过程
func Dec(sk, lambda, sigma, m string) bool {
	// 简单验证：检查所有参数非空且sigma长度正确
	return len(sk) > 0 && len(lambda) > 0 && len(sigma) == 64
}

// Reg KC执行注册逻辑
func Reg(crs, pp, did, pk, xi string) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%d", crs, pp, did, pk, xi, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Upd 更新开口信息
func Upd(pp, did string) string {
	data := fmt.Sprintf("%s|%s|%d", pp, did, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ==================== 接口实现 ====================

// RegisterHandlers 注册消息处理函数
func (mod *RBEIdentityAuthMod) RegisterHandlers() {
	mod.p2pMod.RegisterHandler(MsgRBERegister, mod.handleRegisterMsg)
	mod.p2pMod.RegisterHandler(MsgRBERegisterReply, mod.handleRegisterReplyMsg)
	mod.p2pMod.RegisterHandler(MsgRBESign, mod.handleSignMsg)
	mod.p2pMod.RegisterHandler(MsgRBEVerify, mod.handleVerifyMsg)
	mod.p2pMod.RegisterHandler(MsgRBEVerifyReply, mod.handleVerifyReplyMsg)
	mod.p2pMod.RegisterHandler(MsgRBECheckDID, mod.handleCheckDIDMsg)
	mod.p2pMod.RegisterHandler(MsgRBECheckDIDReply, mod.handleCheckDIDReplyMsg)
	mod.p2pMod.RegisterHandler(MsgRBEMonitorQuery, mod.handleMonitorQueryMsg)

	utils.LoggerInstance.Info("节点 [分片%d, 节点%d] 注册RBE消息处理器完成",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid)
}

// Run 节点运行主循环
func (mod *RBEIdentityAuthMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	utils.LoggerInstance.Info("节点 [分片%d, 节点%d] RBE模块开始运行（KC=%v）",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid, mod.isKC)

	// 等待所有节点启动
	time.Sleep(500 * time.Millisecond)

	// KC节点运行逻辑
	if mod.isKC {
		mod.runAsKC(ctx)
		return
	}

	// 普通节点：首先注册到KC
	mod.registerToKC()

	// 等待注册完成
	for !mod.registered {
		time.Sleep(100 * time.Millisecond)
	}

	utils.LoggerInstance.Info("节点 [分片%d, 节点%d] 注册完成，开始认证测试",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid)

	// 记录开始时间
	mod.startTime = time.Now()
	mod.lastReportTime = mod.startTime

	// 定期执行认证任务
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	// 定期报告统计
	reportTicker := time.NewTicker(5 * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			mod.reportStats()
			utils.LoggerInstance.Info("节点 [分片%d, 节点%d] RBE模块停止",
				mod.nodeAttr.Sid, mod.nodeAttr.Nid)
			return

		case <-ticker.C:
			// 执行随机认证
			mod.performRandomAuth()

		case <-reportTicker.C:
			// 定期报告统计
			mod.reportStats()
		}
	}
}

// runAsKC KC节点运行逻辑
func (mod *RBEIdentityAuthMod) runAsKC(ctx context.Context) {
	// KC节点自己也需要注册
	mod.registerLock.Lock()
	mod.Gen()
	mod.Λ = Reg(mod.crs, mod.pp, mod.DID, mod.PK, mod.ξ)
	mod.registryLock.Lock()
	mod.registry[mod.DID] = mod.PK
	mod.openingInfo[mod.DID] = mod.Λ
	mod.registryLock.Unlock()
	mod.registered = true
	mod.registerLock.Unlock()

	utils.LoggerInstance.Info("KC节点 [分片%d, 节点%d] 自注册完成",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid)

	mod.startTime = time.Now()

	// 定期报告KC状态
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			mod.reportKCStats()
			utils.LoggerInstance.Info("KC节点 [分片%d, 节点%d] RBE模块停止",
				mod.nodeAttr.Sid, mod.nodeAttr.Nid)
			return

		case <-ticker.C:
			mod.reportKCStats()
		}
	}
}

// ==================== 注册流程 ====================

// registerToKC 向KC发送注册请求
func (mod *RBEIdentityAuthMod) registerToKC() {
	mod.registerLock.Lock()
	defer mod.registerLock.Unlock()

	if mod.registered {
		return
	}

	// 生成密钥对
	mod.Gen()

	// 构造注册请求
	req := RegisterRequest{
		DID: mod.DID,
		PK:  mod.PK,
		Xi:  mod.ξ,
	}

	// 发送到KC（节点0）
	kcIP := config.IPMap[mod.nodeAttr.Sid][0]
	msg := message.Message{
		MsgType: MsgRBERegister,
		Content: utils.Encode(req),
	}

	mod.p2pMod.ConnMananger.Send(kcIP, msg.JsonEncode())
	utils.LoggerInstance.Info("节点 [分片%d, 节点%d] 发送注册请求到KC，DID: %s",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid, mod.DID)
}

// handleRegisterMsg KC处理注册请求
func (mod *RBEIdentityAuthMod) handleRegisterMsg(msg *message.Message) {
	if !mod.isKC {
		return
	}

	var req RegisterRequest
	err := utils.Decode(msg.Content, &req)
	if err != nil {
		utils.LoggerInstance.Error("KC节点解码注册请求失败: %v", err)
		return
	}

	utils.LoggerInstance.Info("KC节点 [分片%d, 节点%d] 收到注册请求，DID: %s",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid, req.DID)

	// 执行注册算法
	lambda := Reg(mod.crs, mod.pp, req.DID, req.PK, req.Xi)

	// 更新注册表
	mod.registryLock.Lock()
	mod.registry[req.DID] = req.PK
	mod.openingInfo[req.DID] = lambda
	regCount := len(mod.registry)
	mod.registryLock.Unlock()

	// 构造响应
	resp := RegisterResponse{
		Success: true,
		DID:     req.DID,
		Lambda:  lambda,
		PP:      mod.pp,
	}

	// 广播注册结果到分片内所有节点
	respMsg := message.Message{
		MsgType: MsgRBERegisterReply,
		Content: utils.Encode(resp),
	}

	neighbors := utils.GetNeighbours(config.IPMap[mod.nodeAttr.Sid], mod.nodeAttr.Ipaddr)
	mod.p2pMod.ConnMananger.Broadcast(mod.nodeAttr.Ipaddr, neighbors, respMsg.JsonEncode())

	utils.LoggerInstance.Info("KC节点 [分片%d, 节点%d] 完成注册，广播并上链，DID: %s，已注册节点数: %d",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid, req.DID, regCount)
}

// handleRegisterReplyMsg 处理注册响应
func (mod *RBEIdentityAuthMod) handleRegisterReplyMsg(msg *message.Message) {
	var resp RegisterResponse
	err := utils.Decode(msg.Content, &resp)
	if err != nil {
		utils.LoggerInstance.Error("节点解码注册响应失败: %v", err)
		return
	}

	// 只处理自己的注册响应
	if resp.DID != mod.DID {
		return
	}

	mod.registerLock.Lock()
	defer mod.registerLock.Unlock()

	if !mod.registered && resp.Success {
		mod.Λ = resp.Lambda
		mod.pp = resp.PP
		mod.registered = true

		utils.LoggerInstance.Info("节点 [分片%d, 节点%d] 收到注册成功响应，DID: %s",
			mod.nodeAttr.Sid, mod.nodeAttr.Nid, mod.DID)
	}
}

// ==================== 认证流程 ====================

// performRandomAuth 执行随机认证测试
func (mod *RBEIdentityAuthMod) performRandomAuth() {
	// 随机选择分片内的另一个节点
	shardNodes := config.IPMap[mod.nodeAttr.Sid]
	nodeCount := len(shardNodes)
	if nodeCount <= 1 {
		return
	}

	// 随机选择目标节点（避免选择自己和KC）
	targetNid := (mod.nodeAttr.Nid + 1 + int(time.Now().UnixNano()%int64(nodeCount-1))) % nodeCount
	if targetNid == mod.nodeAttr.Nid {
		targetNid = (targetNid + 1) % nodeCount
	}

	// targetDID := fmt.Sprintf("did_s%d_n%d", mod.nodeAttr.Sid, targetNid)
	testMessage := fmt.Sprintf("auth_test_%d", time.Now().UnixNano())

	// 使用Enc生成签名
	sigma := Enc(mod.crs, mod.pp, mod.DID, testMessage)

	// 构造验证请求
	verifyReq := VerifyRequest{
		FromDID:   mod.DID,
		Message:   testMessage,
		Sigma:     sigma,
		Timestamp: time.Now().UnixNano(),
	}

	// 发送到目标节点
	targetIP := config.IPMap[mod.nodeAttr.Sid][targetNid]
	msg := message.Message{
		MsgType: MsgRBEVerify,
		Content: utils.Encode(verifyReq),
	}

	mod.p2pMod.ConnMananger.Send(targetIP, msg.JsonEncode())
}

// handleSignMsg 处理签名请求
func (mod *RBEIdentityAuthMod) handleSignMsg(msg *message.Message) {
	var signedMsg SignedMessage
	err := utils.Decode(msg.Content, &signedMsg)
	if err != nil {
		utils.LoggerInstance.Error("节点解码签名消息失败: %v", err)
		return
	}

	utils.LoggerInstance.Debug("节点 [分片%d, 节点%d] 收到签名消息，来自: %s",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid, signedMsg.FromDID)
}

// handleVerifyMsg 处理验证请求
func (mod *RBEIdentityAuthMod) handleVerifyMsg(msg *message.Message) {
	var req VerifyRequest
	err := utils.Decode(msg.Content, &req)
	if err != nil {
		utils.LoggerInstance.Error("节点解码验证请求失败: %v", err)
		return
	}

	utils.LoggerInstance.Debug("节点 [分片%d, 节点%d] 收到验证请求，来自DID: %s",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid, req.FromDID)

	// 生成请求ID
	requestID := fmt.Sprintf("verify_%d_%d", mod.nodeAttr.Nid, time.Now().UnixNano())

	// 缓存验证请求
	mod.pendingVerifyLock.Lock()
	mod.pendingVerify[requestID] = &req
	mod.pendingVerifyLock.Unlock()

	// 向KC查询DID是否已注册
	checkReq := CheckDIDRequest{
		DID:       req.FromDID,
		RequestID: requestID,
	}

	kcIP := config.IPMap[mod.nodeAttr.Sid][0]
	checkMsg := message.Message{
		MsgType: MsgRBECheckDID,
		Content: utils.Encode(checkReq),
	}

	mod.p2pMod.ConnMananger.Send(kcIP, checkMsg.JsonEncode())
}

// handleCheckDIDMsg KC处理DID检查请求
func (mod *RBEIdentityAuthMod) handleCheckDIDMsg(msg *message.Message) {
	if !mod.isKC {
		return
	}

	var req CheckDIDRequest
	err := utils.Decode(msg.Content, &req)
	if err != nil {
		utils.LoggerInstance.Error("KC节点解码检查DID请求失败: %v", err)
		return
	}

	// 查询注册表
	mod.registryLock.RLock()
	_, exists := mod.registry[req.DID]
	lambda := mod.openingInfo[req.DID]
	mod.registryLock.RUnlock()

	// 构造响应
	resp := CheckDIDResponse{
		Exists:    exists,
		DID:       req.DID,
		Lambda:    lambda,
		RequestID: req.RequestID,
	}

	// 发送响应（广播，简化处理）
	respMsg := message.Message{
		MsgType: MsgRBECheckDIDReply,
		Content: utils.Encode(resp),
	}

	neighbors := utils.GetNeighbours(config.IPMap[mod.nodeAttr.Sid], mod.nodeAttr.Ipaddr)
	mod.p2pMod.ConnMananger.Broadcast(mod.nodeAttr.Ipaddr, neighbors, respMsg.JsonEncode())
}

// handleCheckDIDReplyMsg 处理DID检查响应
func (mod *RBEIdentityAuthMod) handleCheckDIDReplyMsg(msg *message.Message) {
	var resp CheckDIDResponse
	err := utils.Decode(msg.Content, &resp)
	if err != nil {
		utils.LoggerInstance.Error("节点解码检查DID响应失败: %v", err)
		return
	}

	// 获取原始验证请求
	mod.pendingVerifyLock.Lock()
	verifyReq, exists := mod.pendingVerify[resp.RequestID]
	if exists {
		delete(mod.pendingVerify, resp.RequestID)
	}
	mod.pendingVerifyLock.Unlock()

	if !exists {
		return
	}

	// 检查DID是否存在
	if !resp.Exists {
		utils.LoggerInstance.Warn("节点 [分片%d, 节点%d] DID未注册: %s",
			mod.nodeAttr.Sid, mod.nodeAttr.Nid, resp.DID)
		return
	}

	// 使用Dec验证签名
	success := Dec(mod.SK, resp.Lambda, verifyReq.Sigma, verifyReq.Message)

	if success {
		// 增加验证计数
		atomic.AddInt64(&mod.verifyCount, 1)
		atomic.AddInt64(&mod.totalVerifyCount, 1)

		utils.LoggerInstance.Debug("节点 [分片%d, 节点%d] 验证成功，来自DID: %s",
			mod.nodeAttr.Sid, mod.nodeAttr.Nid, verifyReq.FromDID)
	} else {
		utils.LoggerInstance.Warn("节点 [分片%d, 节点%d] 验证失败，来自DID: %s",
			mod.nodeAttr.Sid, mod.nodeAttr.Nid, verifyReq.FromDID)
	}

	// 构造验证响应
	verifyResp := VerifyResponse{
		Success:   success,
		FromDID:   verifyReq.FromDID,
		Timestamp: verifyReq.Timestamp,
	}

	// 发送验证响应（如果需要）
	// 这里简化处理，不发送响应
	_ = verifyResp
}

// handleVerifyReplyMsg 处理验证响应
func (mod *RBEIdentityAuthMod) handleVerifyReplyMsg(msg *message.Message) {
	var resp VerifyResponse
	err := utils.Decode(msg.Content, &resp)
	if err != nil {
		utils.LoggerInstance.Error("节点解码验证响应失败: %v", err)
		return
	}

	if resp.Success {
		utils.LoggerInstance.Debug("节点 [分片%d, 节点%d] 收到验证成功响应",
			mod.nodeAttr.Sid, mod.nodeAttr.Nid)
	}
}

// ==================== 统计报告 ====================

// reportStats 报告统计信息
func (mod *RBEIdentityAuthMod) reportStats() {
	mod.metricsLock.Lock()
	defer mod.metricsLock.Unlock()
	currentCount := atomic.LoadInt64(&mod.verifyCount)
	totalCount := atomic.LoadInt64(&mod.totalVerifyCount)

	duration := time.Since(mod.lastReportTime).Seconds()
	totalDuration := time.Since(mod.startTime).Seconds()
	if duration < 0.1 {
		return
	}
	instantTPS := float64(currentCount) / duration // 当前TPS
	avgTPS := float64(totalCount) / totalDuration  // 平均TPS

	utils.LoggerInstance.Info("节点 [分片%d, 节点%d] 统计 - 周期验证: %d, 累计验证: %d, 瞬时TPS: %.2f, 平均TPS: %.2f",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid, currentCount, totalCount, instantTPS, avgTPS)
	// 只重置周期计数，保留总计数
	atomic.StoreInt64(&mod.verifyCount, 0)
	mod.lastReportTime = time.Now()
}

// reportKCStats KC节点报告统计
func (mod *RBEIdentityAuthMod) reportKCStats() {
	mod.registryLock.RLock()
	regCount := len(mod.registry)
	mod.registryLock.RUnlock()

	duration := time.Since(mod.startTime).Seconds()

	utils.LoggerInstance.Info("KC节点 [分片%d, 节点%d] 统计 - 注册节点数: %d, 运行时长: %.2fs",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid, regCount, duration)
}

// ==================== 监控功能 ====================

// handleMonitorQueryMsg 处理监控查询消息
func (mod *RBEIdentityAuthMod) handleMonitorQueryMsg(msg *message.Message) {
	// 获取当前统计数据
	currentCount := atomic.LoadInt64(&mod.verifyCount)
	totalCount := atomic.LoadInt64(&mod.totalVerifyCount) // 使用总计数
	totalDuration := time.Since(mod.startTime).Seconds()

	// 计算瞬时TPS（基于当前周期）
	instantDuration := time.Since(mod.lastReportTime).Seconds()
	instantTPS := 0.0
	if instantDuration > 0 {
		instantTPS = float64(currentCount) / instantDuration
	}

	// 计算平均TPS（基于总时长）
	avgTPS := 0.0
	if totalDuration > 0 {
		avgTPS = float64(totalCount) / totalDuration
	}
	// 构造报告
	report := MonitorReport{
		Sid:              mod.nodeAttr.Sid,
		Nid:              mod.nodeAttr.Nid,
		VerifyCount:      currentCount,
		TotalVerifyCount: totalCount, // 新增
		Duration:         totalDuration,
		InstantTPS:       instantTPS, // 新增
		AvgTPS:           avgTPS,
		Timestamp:        time.Now().UnixNano(),
	}

	// 解析发送方IP（从消息中获取）
	var queryIP string
	err := utils.Decode(msg.Content, &queryIP)
	if err != nil {
		utils.LoggerInstance.Error("节点解码监控查询失败: %v", err)
		return
	}

	// 发送报告
	reportMsg := message.Message{
		MsgType: MsgRBEMonitorReport,
		Content: utils.Encode(report),
	}

	mod.p2pMod.ConnMananger.Send(queryIP, reportMsg.JsonEncode())
	utils.LoggerInstance.Debug("节点 [分片%d, 节点%d] 发送监控报告",
		mod.nodeAttr.Sid, mod.nodeAttr.Nid)
}

// GetStats 获取当前统计数据（供外部调用）
func (mod *RBEIdentityAuthMod) GetStats() (totalCount int64, avgTPS float64, instantTPS float64) {
	totalCount = atomic.LoadInt64(&mod.totalVerifyCount)
	currentCount := atomic.LoadInt64(&mod.verifyCount)

	totalDuration := time.Since(mod.startTime).Seconds()
	if totalDuration > 0 {
		avgTPS = float64(totalCount) / totalDuration
	}

	instantDuration := time.Since(mod.lastReportTime).Seconds()
	if instantDuration > 0 {
		instantTPS = float64(currentCount) / instantDuration
	}

	return totalCount, avgTPS, instantTPS
}
