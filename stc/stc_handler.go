package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func (p *STCPlugin) handleInject(msg *message.Message) {
	if p.malicious && p.behavior == "silent" {
		return
	}
	if p.nodeAttr.Nid != config.ViewNodeId {
		viewIP := config.IPMap[p.nodeAttr.Sid][config.ViewNodeId]
		forward := message.Message{MsgType: MsgSTCForwardTx, Content: msg.Content}
		_ = p.p2pMod.ConnManager.Send(viewIP, forward.JsonEncode())
		return
	}
	p.handleInjectContent(msg.Content)
}

func (p *STCPlugin) handleForwardTx(msg *message.Message) {
	if p.malicious && p.behavior == "silent" {
		return
	}
	if p.nodeAttr.Nid != config.ViewNodeId {
		return
	}
	p.handleInjectContent(msg.Content)
}

func (p *STCPlugin) handleInjectContent(content []byte) {
	var req STCInjectRequest
	if err := utils.Decode(content, &req); err == nil && len(req.Txs) > 0 {
		p.routeAndEnqueue(req.Txs)
		return
	}

	rawTxs := []structs.Transaction{}
	if err := utils.Decode(content, &rawTxs); err == nil && len(rawTxs) > 0 {
		txs := make([]STCTransaction, 0, len(rawTxs))
		now := time.Now().UnixNano()
		for i, tx := range rawTxs {
			h := sha256.Sum256(tx.ID())
			txs = append(txs, STCTransaction{
				ID:             fmt.Sprintf("legacy-%d-%s", i, hex.EncodeToString(h[:8])),
				Payload:        tx.Type(),
				SubmitUnixNano: now,
				FromShard:      p.nodeAttr.Sid,
			})
		}
		p.routeAndEnqueue(txs)
		return
	}

	fallback := STCTransaction{
		ID:             fmt.Sprintf("auto-%d-%d-%d", p.nodeAttr.Sid, p.nodeAttr.Nid, time.Now().UnixNano()),
		Payload:        "fallback",
		SubmitUnixNano: time.Now().UnixNano(),
		FromShard:      p.nodeAttr.Sid,
	}
	p.routeAndEnqueue([]STCTransaction{fallback})
}

func (p *STCPlugin) routeAndEnqueue(txs []STCTransaction) {
	local := make([]STCTransaction, 0, len(txs))
	forwardMap := make(map[int][]STCTransaction)
	for _, tx := range txs {
		target := p.pickTargetShard(tx)
		if target == p.nodeAttr.Sid {
			local = append(local, tx)
		} else {
			forwardMap[target] = append(forwardMap[target], tx)
		}
	}
	if len(local) > 0 {
		p.enqueueTxs(local)
	}
	for sid, batch := range forwardMap {
		viewIP := config.IPMap[sid][config.ViewNodeId]
		m := message.Message{MsgType: MsgSTCForwardTx, Content: utils.Encode(STCInjectRequest{Txs: batch})}
		_ = p.p2pMod.ConnManager.Send(viewIP, m.JsonEncode())
	}
}

func (p *STCPlugin) pickTargetShard(tx STCTransaction) int {
	if config.ShardNum <= 1 {
		return p.nodeAttr.Sid
	}
	h := sha256.Sum256([]byte(tx.ID))
	v := int(h[0])
	if v < 0 {
		v = -v
	}
	return v % config.ShardNum
}

func (p *STCPlugin) handleConsensus(msg *message.Message) {
	if p.malicious && p.behavior == "silent" {
		return
	}
	var packet STCConsensusPacket
	if err := utils.Decode(msg.Content, &packet); err != nil {
		utils.LoggerInstance.Error("STC decode consensus failed: %v", err)
		return
	}
	if err := p.validator.Validate(packet.Envelope); err != nil {
		p.addLog("ERROR", "space_time_anomaly", fmt.Sprintf("node=%d shard=%d err=%v", packet.Envelope.NodeID, packet.Envelope.ShardID, err))
		return
	}
	if packet.Type != "commit" {
		return
	}
	if packet.Envelope.NodeID != config.ViewNodeId || packet.Envelope.ShardID != p.nodeAttr.Sid {
		p.addLog("WARN", "byzantine_block_rejected", fmt.Sprintf("rejected commit from node %d shard %d", packet.Envelope.NodeID, packet.Envelope.ShardID))
		return
	}
	if packet.Block.Height <= 0 {
		return
	}
	txs := make([]STCTransaction, 0, len(packet.Block.TxIDs))
	now := time.Now().UnixNano()
	for _, id := range packet.Block.TxIDs {
		txs = append(txs, STCTransaction{ID: id, SubmitUnixNano: now})
	}
	p.applyCommit(packet.Block, txs)
}

func (p *STCPlugin) handleControl(msg *message.Message) {
	var req STCControlRequest
	if err := utils.Decode(msg.Content, &req); err != nil {
		utils.LoggerInstance.Error("STC decode control failed: %v", err)
		return
	}
	if req.TargetShard >= 0 && req.TargetShard != p.nodeAttr.Sid {
		return
	}
	if req.TargetNode >= 0 && req.TargetNode != p.nodeAttr.Nid {
		return
	}
	switch req.Action {
	case "forge_spacetime_once":
		p.mu.Lock()
		p.forgeNext = true
		p.forgeLocation = req.FakeLocation
		p.mu.Unlock()
		p.addLog("WARN", "control", "forge spacetime for next consensus message")
	case "set_byzantine":
		p.mu.Lock()
		p.malicious = req.Enabled
		p.behavior = req.Behavior
		p.mu.Unlock()
		p.addLog("WARN", "control", fmt.Sprintf("set malicious=%v behavior=%s", req.Enabled, req.Behavior))
	case "clear_logs":
		p.mu.Lock()
		p.logs = nil
		p.mu.Unlock()
	}
}

func (p *STCPlugin) handleQuery(msg *message.Message) {
	var req STCQueryRequest
	if err := utils.Decode(msg.Content, &req); err != nil {
		utils.LoggerInstance.Error("STC decode query failed: %v", err)
		return
	}
	reply := STCQueryReply{RequestID: req.RequestID, Action: req.Action, OK: true}
	switch req.Action {
	case "status", "node_info":
		reply.Status = p.getNodeStatus()
	case "metrics":
		m := p.metricsSnapshot()
		reply.Metrics = &m
	case "blocks":
		reply.Blocks = p.getBlockRange(req.StartHeight, req.EndHeight)
	case "logs":
		s := p.getNodeStatus()
		reply.Status = s
	default:
		reply.OK = false
		reply.Error = "unsupported action"
	}
	if strings.TrimSpace(req.ReplyTo) == "" {
		return
	}
	resp := message.Message{MsgType: MsgSTCQueryReply, Content: utils.Encode(reply)}
	if err := p.p2pMod.ConnManager.Send(req.ReplyTo, resp.JsonEncode()); err != nil {
		utils.LoggerInstance.Error("STC send query reply failed: %v", err)
	}
}

func (p *STCPlugin) getBlockRange(start, end int64) []STCBlockHashRecord {
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	p.mu.RLock()
	if end > p.latestH {
		end = p.latestH
	}
	res := make([]STCBlockHashRecord, 0, end-start+1)
	for h := start; h <= end; h++ {
		if hash, ok := p.blockHashes[h]; ok {
			res = append(res, STCBlockHashRecord{Height: h, Hash: hash})
		}
	}
	p.mu.RUnlock()
	return res
}

func (p *STCPlugin) getNodeStatus() *STCNodeStatus {
	m := p.metricsSnapshot()
	p.mu.RLock()
	logs := make([]STCLogEvent, len(p.logs))
	copy(logs, p.logs)
	status := &STCNodeStatus{
		ShardID:    p.nodeAttr.Sid,
		NodeID:     p.nodeAttr.Nid,
		Online:     true,
		Role:       map[bool]string{true: "view", false: "normal"}[p.isView()],
		Malicious:  p.malicious,
		Behavior:   p.behavior,
		Location:   p.location,
		IP:         p.nodeAttr.Ipaddr,
		LatestHash: p.latestHash,
		Metrics:    m,
		Logs:       logs,
	}
	p.mu.RUnlock()
	return status
}
