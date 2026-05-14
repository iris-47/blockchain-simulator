package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/plugins/plugininterface"
	"BlockChainSimulator/utils"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

var _ plugininterface.Plugin = &STCPlugin{}

const forgedTimestampOffset = 10 * time.Minute

type STCPlugin struct {
	nodeAttr *nodeattr.NodeAttr
	p2pMod   *p2p.P2PMod

	validator *SpaceTimeValidator
	monitor   *TPSMonitor

	mu          sync.RWMutex
	pendingTx   map[string]STCTransaction
	pendingList []string
	blockHashes map[int64]string
	latestHash  string
	latestH     int64
	logs        []STCLogEvent

	forgeNext     bool
	forgeLocation string
	location      string

	malicious bool
	behavior  string
}

func NewSTCPlugin(attr *nodeattr.NodeAttr, p2pMod *p2p.P2PMod) plugininterface.Plugin {
	p := &STCPlugin{
		nodeAttr:    attr,
		p2pMod:      p2pMod,
		validator:   NewSpaceTimeValidator(),
		monitor:     NewTPSMonitor(),
		pendingTx:   make(map[string]STCTransaction),
		blockHashes: make(map[int64]string),
		logs:        make([]STCLogEvent, 0, 128),
		location:    fmt.Sprintf("zone-shard-%d", attr.Sid),
	}
	p.latestH = 0
	p.latestHash = fmt.Sprintf("GENESIS-%d", attr.Sid)
	p.blockHashes[0] = p.latestHash
	return p
}

func (p *STCPlugin) Initialize() {
	p.p2pMod.RegisterHandler(message.MsgInject, p.handleInject)
	p.p2pMod.RegisterHandler(MsgSTCForwardTx, p.handleForwardTx)
	p.p2pMod.RegisterHandler(MsgSTCConsensus, p.handleConsensus)
	p.p2pMod.RegisterHandler(MsgSTCControl, p.handleControl)
	p.p2pMod.RegisterHandler(MsgSTCQuery, p.handleQuery)
}

func (p *STCPlugin) Cleanup() {}

func (p *STCPlugin) isView() bool {
	return p.nodeAttr.Nid == config.ViewNodeId
}

func (p *STCPlugin) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	consensusTicker := time.NewTicker(50 * time.Millisecond)
	metricsTicker := time.NewTicker(time.Second)
	maliciousTicker := time.NewTicker(1500 * time.Millisecond)
	defer consensusTicker.Stop()
	defer metricsTicker.Stop()
	defer maliciousTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-consensusTicker.C:
			if p.isView() && !(p.malicious && p.behavior == "silent") {
				p.proposeAndCommitBatch()
			}
		case <-metricsTicker.C:
			p.sendMetricsReport()
		case <-maliciousTicker.C:
			if p.malicious && p.behavior == "broadcast_bad_block" {
				p.broadcastBadBlock()
			}
		}
	}
}

func (p *STCPlugin) proposeAndCommitBatch() {
	txs := p.popBatch(5000)
	if len(txs) == 0 {
		return
	}
	b := p.buildBlock(txs)
	p.applyCommit(b, txs)
	packet := STCConsensusPacket{
		Envelope: p.makeEnvelope(),
		Type:     "commit",
		Block:    b,
	}
	msg := message.Message{MsgType: MsgSTCConsensus, Content: utils.Encode(packet)}
	p.p2pMod.ConnManager.Broadcast(p.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[p.nodeAttr.Sid], p.nodeAttr.Ipaddr), msg.JsonEncode())
}

func (p *STCPlugin) popBatch(limit int) []STCTransaction {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.pendingList) == 0 {
		return nil
	}
	if limit > len(p.pendingList) {
		limit = len(p.pendingList)
	}
	res := make([]STCTransaction, 0, limit)
	for i := 0; i < limit; i++ {
		id := p.pendingList[i]
		tx, ok := p.pendingTx[id]
		if !ok {
			continue
		}
		res = append(res, tx)
		delete(p.pendingTx, id)
	}
	p.pendingList = p.pendingList[limit:]
	return res
}

func (p *STCPlugin) buildBlock(txs []STCTransaction) STCBlock {
	p.mu.RLock()
	prevHash := p.latestHash
	h := p.latestH + 1
	p.mu.RUnlock()

	hasher := sha256.New()
	hasher.Write([]byte(prevHash))
	hasher.Write([]byte(fmt.Sprintf("%d|%d|%d", h, p.nodeAttr.Sid, p.nodeAttr.Nid)))
	txIDs := make([]string, 0, len(txs))
	for _, tx := range txs {
		txIDs = append(txIDs, tx.ID)
		hasher.Write([]byte(tx.ID))
	}
	return STCBlock{
		Height:     h,
		PrevHash:   prevHash,
		Hash:       hex.EncodeToString(hasher.Sum(nil)),
		TxIDs:      txIDs,
		Timestamp:  time.Now().UnixNano(),
		ShardID:    p.nodeAttr.Sid,
		ProposerID: p.nodeAttr.Nid,
	}
}

func (p *STCPlugin) applyCommit(b STCBlock, txs []STCTransaction) {
	p.mu.Lock()
	if b.Height != p.latestH+1 || b.PrevHash != p.latestHash {
		p.mu.Unlock()
		return
	}
	p.latestH = b.Height
	p.latestHash = b.Hash
	p.blockHashes[b.Height] = b.Hash
	p.mu.Unlock()

	p.monitor.RecordCommitted(len(txs))
	now := time.Now()
	for _, tx := range txs {
		if tx.SubmitUnixNano > 0 {
			p.monitor.RecordConfirmLatency(now.Sub(time.Unix(0, tx.SubmitUnixNano)))
		}
	}
}

func (p *STCPlugin) makeEnvelope() STCEnvelope {
	env := STCEnvelope{
		ShardID:   p.nodeAttr.Sid,
		NodeID:    p.nodeAttr.Nid,
		Timestamp: time.Now().UnixNano(),
		Location:  p.location,
	}
	p.mu.Lock()
	if p.forgeNext {
		env.Timestamp = time.Now().Add(forgedTimestampOffset).UnixNano()
		if p.forgeLocation != "" {
			env.Location = p.forgeLocation
		} else {
			env.Location = p.location + "-forged"
		}
		p.forgeNext = false
	}
	p.mu.Unlock()
	return env
}

func (p *STCPlugin) broadcastBadBlock() {
	p.mu.RLock()
	b := STCBlock{
		Height:     p.latestH + 1,
		PrevHash:   "bad-prev",
		Hash:       fmt.Sprintf("malicious-%d-%d", p.nodeAttr.Nid, time.Now().UnixNano()),
		Timestamp:  time.Now().UnixNano(),
		ShardID:    p.nodeAttr.Sid,
		ProposerID: p.nodeAttr.Nid,
	}
	p.mu.RUnlock()
	packet := STCConsensusPacket{Envelope: p.makeEnvelope(), Type: "commit", Block: b}
	msg := message.Message{MsgType: MsgSTCConsensus, Content: utils.Encode(packet)}
	p.p2pMod.ConnManager.Broadcast(p.nodeAttr.Ipaddr, utils.GetNeighbours(config.IPMap[p.nodeAttr.Sid], p.nodeAttr.Ipaddr), msg.JsonEncode())
}

func (p *STCPlugin) addLog(level, t, detail string) {
	e := STCLogEvent{Level: level, Type: t, Timestamp: time.Now().UnixNano(), NodeID: p.nodeAttr.Nid, ShardID: p.nodeAttr.Sid, Detail: detail}
	p.mu.Lock()
	p.logs = append(p.logs, e)
	if len(p.logs) > 500 {
		p.logs = p.logs[len(p.logs)-500:]
	}
	p.mu.Unlock()
	if level == "ERROR" {
		utils.LoggerInstance.Error("[STC][%s] %s", t, detail)
	} else {
		utils.LoggerInstance.Warn("[STC][%s] %s", t, detail)
	}
}

func (p *STCPlugin) pendingSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.pendingList)
}

func (p *STCPlugin) metricsSnapshot() STCMetricsSnapshot {
	return p.monitor.Snapshot(p.nodeAttr.Sid, p.nodeAttr.Nid, p.pendingSize(), p.latestH)
}

func (p *STCPlugin) sendMetricsReport() {
	report := STCMetricsReport{Envelope: p.makeEnvelope(), Metrics: p.metricsSnapshot()}
	msg := message.Message{MsgType: MsgSTCMetricsReport, Content: utils.Encode(report)}
	if err := p.p2pMod.ConnManager.Send(config.ClientAddr, msg.JsonEncode()); err != nil {
		utils.LoggerInstance.Debug("failed to send STC metric report to client: %v", err)
	}
}

func (p *STCPlugin) enqueueTxs(txs []STCTransaction) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, tx := range txs {
		if tx.ID == "" {
			continue
		}
		if _, exists := p.pendingTx[tx.ID]; exists {
			continue
		}
		if tx.SubmitUnixNano == 0 {
			tx.SubmitUnixNano = time.Now().UnixNano()
		}
		p.pendingTx[tx.ID] = tx
		p.pendingList = append(p.pendingList, tx.ID)
	}
}
