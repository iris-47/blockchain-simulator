package clientMod

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/utils"
	"context"
	"sync"
)

// used by client node to run the whole blockchain system
type StopSystemAuxiliaryMod struct {
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node
}

func NewStopSystemAuxiliaryMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	ssam := new(StopSystemAuxiliaryMod)
	ssam.nodeAttr = attr
	ssam.p2pMod = p2p

	return ssam
}

func (ssam *StopSystemAuxiliaryMod) RegisterHandlers() {

}

func (ssam *StopSystemAuxiliaryMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	<-ctx.Done()
	utils.LoggerInstance.Info("Stop the StopSystemAuxiliaryMod, send stop message to all nodes")
	for i := 0; i < config.ShardNum; i++ {
		for j := 0; j < config.NodeNum; j++ {
			msg := message.Message{
				MsgType: message.MsgStop,
			}
			ssam.p2pMod.ConnMananger.Send(config.IPMap[i][j], msg.JsonEncode())
		}
	}
}
