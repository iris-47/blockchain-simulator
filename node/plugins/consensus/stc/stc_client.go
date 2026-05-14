package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/plugins/plugininterface"
	"BlockChainSimulator/utils"
	"context"
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

	httpServer *http.Server
	apiAddr    string

	pendingMu sync.Mutex
	pending   map[string]chan STCQueryReply

	cacheMu     sync.Mutex
	latestStats map[string]STCQueryReply
}

func NewSTCClientPlugin(attr *nodeattr.NodeAttr, p2pMod *p2p.P2PMod) plugininterface.Plugin {
	apiAddr := "127.0.0.1:18080"
	if host, port, err := net.SplitHostPort(config.ClientAddr); err == nil {
		if p, e := strconv.Atoi(port); e == nil {
			apiAddr = net.JoinHostPort(host, strconv.Itoa(p+1))
		}
	}

	return &STCClientPlugin{
		nodeAttr:    attr,
		p2pMod:      p2pMod,
		apiAddr:     apiAddr,
		pending:     make(map[string]chan STCQueryReply),
		latestStats: make(map[string]STCQueryReply),
	}
}

func (c *STCClientPlugin) Initialize() {
	c.p2pMod.RegisterHandler(message.MsgReplyQuery, c.handleReplyQuery)
	c.setupRoutes()
}

func (c *STCClientPlugin) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	c.httpServer = &http.Server{Addr: c.apiAddr, Handler: c.routes()}
	go func() {
		if err := c.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.LoggerInstance.Error("STCClient HTTP server failed: %v", err)
		}
	}()
	utils.LoggerInstance.Info("STCClient dashboard at http://%s", c.apiAddr)

	pollTicker := time.NewTicker(1 * time.Second)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_ = c.httpServer.Shutdown(shutdownCtx)
			cancel()
			return
		case <-pollTicker.C:
			c.pollMetricsAllNodes()
		}
	}
}

func (c *STCClientPlugin) Cleanup() {}
