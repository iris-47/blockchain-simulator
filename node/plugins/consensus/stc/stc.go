package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/plugins/plugininterface"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

var _ plugininterface.Plugin = &STCPlugin{}

type STCPlugin struct {
	nodeAttr *nodeattr.NodeAttr
	p2pMod   *p2p.P2PMod

	txPool *structs.TxPool

	validator *SpaceTimeValidator
	monitor   *TPSMonitor

	stateMu      sync.Mutex
	recentEvents []string

	forgeNext      bool
	forgeOffsetSec int64
	forgeLocation  string
	maliciousDrop  bool
	broadcastEvil  bool
}

func NewSTCPlugin(attr *nodeattr.NodeAttr, p2pMod *p2p.P2PMod) plugininterface.Plugin {
	p := &STCPlugin{
		nodeAttr:      attr,
		p2pMod:        p2pMod,
		txPool:        structs.NewTxPool(config.BlockSize),
		monitor:       NewTPSMonitor(),
		recentEvents:  make([]string, 0, 256),
		forgeLocation: fmt.Sprintf("shard-%d-node-%d", attr.Sid, attr.Nid),
	}
	p.validator = NewSpaceTimeValidator(p.recordEvent)
	return p
}

func (p *STCPlugin) Initialize() {
	p.p2pMod.RegisterHandler(message.MsgInject, p.handleInject)
	p.p2pMod.RegisterHandler(MsgSTCConsensus, p.handleConsensus)
	p.p2pMod.RegisterHandler(MsgSTCControl, p.handleControl)
	p.p2pMod.RegisterHandler(MsgSTCByzantineBlock, p.handleByzantineBlock)
	p.p2pMod.RegisterHandler(message.MsgQuery, p.handleQuery)
}

func (p *STCPlugin) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	proposeTicker := time.NewTicker(30 * time.Millisecond)
	defer proposeTicker.Stop()

	emptyTicker := time.NewTicker(250 * time.Millisecond)
	defer emptyTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-proposeTicker.C:
			if p.nodeAttr.Nid == config.ViewNodeId {
				p.proposeNext(false)
			}
		case <-emptyTicker.C:
			if p.nodeAttr.Nid == config.ViewNodeId {
				p.proposeNext(true)
			}
			if p.broadcastEvil {
				p.sendByzantineBlock()
			}
		}
	}
}

func (p *STCPlugin) Cleanup() {}

func (p *STCPlugin) proposeNext(allowEmpty bool) {
	var txs []structs.Transaction
	if allowEmpty {
		txs = p.txPool.GetTxs(config.BlockSize)
	} else {
		txs = p.txPool.GetEnoughTxs(1, config.BlockSize)
	}
	if txs == nil {
		return
	}

	b := p.nodeAttr.CurChain.NewBlock(txs)
	payload := STCConsensusPayload{
		Phase:     "preprepare",
		Height:    b.Header.Height,
		Block:     utils.Encode(b),
		BlockHash: b.Hash,
		Envelope:  p.nextEnvelope(),
	}
	msg := message.Message{Sender: p.nodeAttr.Ipaddr, MsgType: MsgSTCConsensus, Content: utils.Encode(payload)}
	p.p2pMod.ConnManager.Broadcast(p.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[p.nodeAttr.Sid], p.nodeAttr.Ipaddr), msg.JsonEncode())
	p.acceptPrePrepare(p.nodeAttr.Ipaddr, &payload)
}

func (p *STCPlugin) handleInject(msg *message.Message) {
	txs := make([]structs.Transaction, 0)
	if err := utils.Decode(msg.Content, &txs); err != nil {
		utils.LoggerInstance.Error("STC decode inject failed: %v", err)
		return
	}
	p.txPool.AddTxs(txs)
}

func (p *STCPlugin) handleConsensus(msg *message.Message) {
	if p.maliciousDrop {
		return
	}

	var payload STCConsensusPayload
	if err := utils.Decode(msg.Content, &payload); err != nil {
		utils.LoggerInstance.Error("STC decode consensus failed: %v", err)
		return
	}

	if !p.validator.Validate(payload.Envelope) {
		return
	}

	switch payload.Phase {
	case "preprepare":
		p.acceptPrePrepare(msg.Sender, &payload)
	case "prepare":
		p.acceptPrepare(&payload)
	case "commit":
		p.acceptCommit(&payload)
	}
}

func (p *STCPlugin) acceptPrePrepare(sender string, payload *STCConsensusPayload) {
	leaderIP := config.IPMap[p.nodeAttr.Sid][config.ViewNodeId]
	if sender != leaderIP {
		p.recordEvent("STC_IGNORE non-leader preprepare sender=%s height=%d", sender, payload.Height)
		return
	}

	if payload.Block == nil {
		return
	}

	var b structs.Block
	if err := utils.Decode(payload.Block, &b); err != nil {
		utils.LoggerInstance.Error("STC decode block failed: %v", err)
		return
	}

	if b.Header == nil {
		return
	}
	if b.Header.Height <= p.nodeAttr.CurChain.CurrentBlock.Header.Height {
		return
	}

	p.nodeAttr.CurChain.CommitBlock(&b)

	var totalDelay float64
	for _, tx := range b.Transactions {
		d := time.Since(tx.GetTime()).Seconds()
		if d < 0 {
			d = 0
		}
		totalDelay += d
	}
	p.monitor.RecordCommit(len(b.Transactions), totalDelay)

	prepare := STCConsensusPayload{
		Phase:     "prepare",
		Height:    payload.Height,
		BlockHash: b.Hash,
		Envelope:  p.nextEnvelope(),
	}
	m := message.Message{Sender: p.nodeAttr.Ipaddr, MsgType: MsgSTCConsensus, Content: utils.Encode(prepare)}
	p.p2pMod.ConnManager.Broadcast(p.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[p.nodeAttr.Sid], p.nodeAttr.Ipaddr), m.JsonEncode())
}

func (p *STCPlugin) acceptPrepare(payload *STCConsensusPayload) {
	commit := STCConsensusPayload{
		Phase:     "commit",
		Height:    payload.Height,
		BlockHash: payload.BlockHash,
		Envelope:  p.nextEnvelope(),
	}
	m := message.Message{Sender: p.nodeAttr.Ipaddr, MsgType: MsgSTCConsensus, Content: utils.Encode(commit)}
	p.p2pMod.ConnManager.Broadcast(p.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[p.nodeAttr.Sid], p.nodeAttr.Ipaddr), m.JsonEncode())
}

func (p *STCPlugin) acceptCommit(_ *STCConsensusPayload) {}

func (p *STCPlugin) nextEnvelope() SpaceTimeEnvelope {
	offset := int64(0)
	location := fmt.Sprintf("shard-%d-node-%d", p.nodeAttr.Sid, p.nodeAttr.Nid)

	p.stateMu.Lock()
	if p.forgeNext {
		offset = p.forgeOffsetSec
		if p.forgeLocation != "" {
			location = p.forgeLocation
		}
		p.forgeNext = false
	}
	p.stateMu.Unlock()

	return SpaceTimeEnvelope{
		ShardID:   p.nodeAttr.Sid,
		NodeID:    p.nodeAttr.Nid,
		Timestamp: time.Now().Add(time.Duration(offset) * time.Second).Unix(),
		Location:  location,
	}
}

func (p *STCPlugin) handleControl(msg *message.Message) {
	var ctl STCControlPayload
	if err := utils.Decode(msg.Content, &ctl); err != nil {
		utils.LoggerInstance.Error("STC decode control failed: %v", err)
		return
	}

	switch ctl.Action {
	case "forge_next_spacetime":
		p.stateMu.Lock()
		p.forgeNext = true
		p.forgeOffsetSec = ctl.TimeOffsetSec
		p.forgeLocation = ctl.FakeLocation
		p.stateMu.Unlock()
		p.recordEvent("STC_CONTROL forge_next_spacetime offset=%d location=%s", ctl.TimeOffsetSec, ctl.FakeLocation)
	case "set_malicious_drop":
		p.stateMu.Lock()
		p.maliciousDrop = ctl.Enabled
		p.stateMu.Unlock()
		p.recordEvent("STC_CONTROL set_malicious_drop enabled=%v", ctl.Enabled)
	case "broadcast_malicious_block":
		p.stateMu.Lock()
		p.broadcastEvil = ctl.Enabled
		p.stateMu.Unlock()
		p.recordEvent("STC_CONTROL broadcast_malicious_block enabled=%v", ctl.Enabled)
	case "set_target_tps":
		p.monitor.SetVirtualTPSFloor(ctl.TargetTPS, ctl.DurationSec)
		p.recordEvent("STC_CONTROL set_target_tps target=%d duration=%d", ctl.TargetTPS, ctl.DurationSec)
	}
}

func (p *STCPlugin) sendByzantineBlock() {
	payload := STCConsensusPayload{
		Phase:     "byzantine",
		Height:    p.nodeAttr.CurChain.CurrentBlock.Header.Height + 1,
		BlockHash: []byte("evil"),
		Envelope:  p.nextEnvelope(),
	}
	msg := message.Message{Sender: p.nodeAttr.Ipaddr, MsgType: MsgSTCByzantineBlock, Content: utils.Encode(payload)}
	p.p2pMod.ConnManager.Broadcast(p.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[p.nodeAttr.Sid], p.nodeAttr.Ipaddr), msg.JsonEncode())
}

func (p *STCPlugin) handleByzantineBlock(msg *message.Message) {
	var payload STCConsensusPayload
	if err := utils.Decode(msg.Content, &payload); err != nil {
		return
	}
	p.recordEvent("STC_REJECT byzantine_block sender=%s height=%d", msg.Sender, payload.Height)
}

func (p *STCPlugin) handleQuery(msg *message.Message) {
	var q STCQuery
	if err := utils.Decode(msg.Content, &q); err != nil {
		return
	}

	rep := STCQueryReply{
		RequestID: q.RequestID,
		Action:    q.Action,
		ShardID:   p.nodeAttr.Sid,
		NodeID:    p.nodeAttr.Nid,
		Online:    true,
	}

	switch q.Action {
	case "metrics", "node_overview":
		rep.Metrics = p.monitor.Snapshot(p.nodeAttr.CurChain.CurrentBlock.Header.Height)
		rep.Events = p.getRecentEvents()
	case "block_hashes":
		rep.BlockHashes = p.collectBlockHashes(q.StartHeight, q.EndHeight)
	}

	msgOut := message.Message{Sender: p.nodeAttr.Ipaddr, MsgType: message.MsgReplyQuery, Content: utils.Encode(rep)}
	if msg.Sender != "" {
		_ = p.p2pMod.ConnManager.Send(msg.Sender, msgOut.JsonEncode())
	}
}

func (p *STCPlugin) collectBlockHashes(start, end int64) []STCBlockHashItem {
	if end < start {
		start, end = end, start
	}
	if start < 0 {
		start = 0
	}

	items := make([]STCBlockHashItem, 0)
	cur := p.nodeAttr.CurChain.CurrentBlock
	for cur != nil {
		h := cur.Header.Height
		if h >= start && h <= end {
			items = append(items, STCBlockHashItem{Height: h, Hash: hex.EncodeToString(cur.Hash)})
		}
		if h <= start || h == 0 {
			break
		}
		next, err := p.nodeAttr.CurChain.Storage.GetBlock(cur.Header.PrevBlockHash)
		if err != nil {
			break
		}
		cur = next
	}
	return items
}

func (p *STCPlugin) recordEvent(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	utils.LoggerInstance.Warn(msg)
	p.stateMu.Lock()
	defer p.stateMu.Unlock()
	if len(p.recentEvents) >= 256 {
		p.recentEvents = p.recentEvents[1:]
	}
	p.recentEvents = append(p.recentEvents, msg)
}

func (p *STCPlugin) getRecentEvents() []string {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()
	out := make([]string, len(p.recentEvents))
	copy(out, p.recentEvents)
	return out
}
