package main

import "BlockChainSimulator/message"

// ==================== 消息类型定义 ====================

const (
	MsgRBERegister message.MessageType = iota + 1000
	MsgRBERegisterReply
	MsgRBESign
	MsgRBEVerify
	MsgRBEVerifyReply
	MsgRBECheckDID
	MsgRBECheckDIDReply
	MsgRBEMonitorQuery  // 监控查询消息
	MsgRBEMonitorReport // 监控报告消息
)

// ==================== 消息结构体定义 ====================

// 注册请求
type RegisterRequest struct {
	DID string
	PK  string
	Xi  string
}

// 注册响应
type RegisterResponse struct {
	Success bool
	DID     string
	Lambda  string
	PP      string
}

// 签名消息
type SignedMessage struct {
	FromDID   string
	ToDID     string
	Message   string
	Sigma     string
	Timestamp int64
}

// 验证请求
type VerifyRequest struct {
	FromDID   string
	Message   string
	Sigma     string
	Timestamp int64
}

// 验证响应
type VerifyResponse struct {
	Success   bool
	FromDID   string
	Timestamp int64
}

// 检查DID请求
type CheckDIDRequest struct {
	DID       string
	RequestID string
}

// 检查DID响应
type CheckDIDResponse struct {
	Exists    bool
	DID       string
	Lambda    string
	RequestID string
}

// 监控统计报告
type MonitorReport struct {
	Sid              int
	Nid              int
	VerifyCount      int64   // 当前周期验证次数
	TotalVerifyCount int64   // 历史累计验证总次数
	Duration         float64 // 总运行时长
	InstantTPS       float64 // 当前TPS
	AvgTPS           float64 // 历史平均TPS
	Timestamp        int64
}
