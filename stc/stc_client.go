package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/plugins/plugininterface"
	"BlockChainSimulator/utils"
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var _ plugininterface.Plugin = &STCClientPlugin{}

type STCClientPlugin struct {
	nodeAttr *nodeattr.NodeAttr
	p2pMod   *p2p.P2PMod

	httpSrv *http.Server

	mu          sync.RWMutex
	nodes       map[string]*STCNodeStatus
	metrics     map[int]STCMetricsSnapshot
	pendingResp map[string]chan STCQueryReply
}

func NewSTCClientPlugin(attr *nodeattr.NodeAttr, p2pMod *p2p.P2PMod) plugininterface.Plugin {
	return &STCClientPlugin{
		nodeAttr:    attr,
		p2pMod:      p2pMod,
		nodes:       make(map[string]*STCNodeStatus),
		metrics:     make(map[int]STCMetricsSnapshot),
		pendingResp: make(map[string]chan STCQueryReply),
	}
}

func (c *STCClientPlugin) Initialize() {
	c.p2pMod.RegisterHandler(MsgSTCQueryReply, c.handleQueryReply)
	c.p2pMod.RegisterHandler(MsgSTCMetricsReport, c.handleMetricsReport)
}

func (c *STCClientPlugin) Cleanup() {}

func (c *STCClientPlugin) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	addr := c.httpListenAddr()
	mux := http.NewServeMux()
	c.registerHTTPRoutes(mux)
	c.httpSrv = &http.Server{Addr: addr, Handler: mux}
	go func() {
		if err := c.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.LoggerInstance.Error("STCClient http server error: %v", err)
		}
	}()
	utils.LoggerInstance.Info("STCClient dashboard listening at %s", addr)

	pollTicker := time.NewTicker(time.Second)
	defer pollTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_ = c.httpSrv.Shutdown(shutdownCtx)
			cancel()
			return
		case <-pollTicker.C:
			c.pollNodeStatus()
		}
	}
}

func (c *STCClientPlugin) httpListenAddr() string {
	host, portStr, err := net.SplitHostPort(config.ClientAddr)
	if err != nil {
		return "127.0.0.1:18080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "127.0.0.1:18080"
	}
	return fmt.Sprintf("%s:%d", host, port+1000)
}

func (c *STCClientPlugin) pollNodeStatus() {
	for sid := 0; sid < config.ShardNum; sid++ {
		for nid := 0; nid < config.NodeNum; nid++ {
			go func(shardID, nodeID int) {
				req := STCQueryRequest{RequestID: fmt.Sprintf("poll-%d-%d-%d", shardID, nodeID, time.Now().UnixNano()), Action: "status", ReplyTo: c.nodeAttr.Ipaddr}
				msg := message.Message{MsgType: MsgSTCQuery, Content: utils.Encode(req)}
				_ = c.p2pMod.ConnManager.Send(config.IPMap[shardID][nodeID], msg.JsonEncode())
			}(sid, nid)
		}
	}
}

func (c *STCClientPlugin) handleQueryReply(msg *message.Message) {
	var reply STCQueryReply
	if err := utils.Decode(msg.Content, &reply); err != nil {
		utils.LoggerInstance.Error("STCClient decode query reply failed: %v", err)
		return
	}
	if reply.Status != nil {
		key := fmt.Sprintf("%d-%d", reply.Status.ShardID, reply.Status.NodeID)
		c.mu.Lock()
		c.nodes[key] = reply.Status
		c.metrics[reply.Status.ShardID] = reply.Status.Metrics
		c.mu.Unlock()
	}
	if reply.Metrics != nil {
		c.mu.Lock()
		c.metrics[reply.Metrics.ShardID] = *reply.Metrics
		c.mu.Unlock()
	}
	c.mu.Lock()
	if ch, ok := c.pendingResp[reply.RequestID]; ok {
		select {
		case ch <- reply:
		default:
		}
	}
	c.mu.Unlock()
}

func (c *STCClientPlugin) handleMetricsReport(msg *message.Message) {
	var report STCMetricsReport
	if err := utils.Decode(msg.Content, &report); err != nil {
		utils.LoggerInstance.Error("STCClient decode metrics report failed: %v", err)
		return
	}
	c.mu.Lock()
	c.metrics[report.Metrics.ShardID] = report.Metrics
	c.mu.Unlock()
}
