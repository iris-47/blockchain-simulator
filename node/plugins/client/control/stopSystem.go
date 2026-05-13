package main

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/plugins/plugininterface"
	"BlockChainSimulator/utils"
	"context"
	"sync"
)

var _ plugininterface.Plugin = &StopSystemPlugin{}

// used by client node to run the whole blockchain system
type StopSystemPlugin struct {
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node
}

func NewStopSystemPlugin(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) plugininterface.Plugin {
	ssam := new(StopSystemPlugin)
	ssam.nodeAttr = attr
	ssam.p2pMod = p2p

	return ssam
}

func (ssam *StopSystemPlugin) Initialize() {
}

func (ssam *StopSystemPlugin) Cleanup() {
}

func (ssam *StopSystemPlugin) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	<-ctx.Done()
	utils.LoggerInstance.Info("Stop the StopSystemPlugin, send stop message to all nodes")
	for i := 0; i < config.ShardNum; i++ {
		for j := 0; j < config.NodeNum; j++ {
			msg := message.Message{
				MsgType: message.MsgStop,
			}
			ssam.p2pMod.ConnManager.Send(config.IPMap[i][j], msg.JsonEncode())
		}
	}
}
