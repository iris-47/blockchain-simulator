package auxiliaryMod

import (
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/utils"
	"context"
	"sync"
	"time"
)

var _ runningModInterface.RunningMod = &TestAuxiliaryMod{}

// this mod will receive the txs from client and propose them to the shard
type TestAuxiliaryMod struct {
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node
}

// this mod will receive the txs from client and propose them to the shard
func NewTestAuxiliaryMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	tam := new(TestAuxiliaryMod)
	tam.nodeAttr = attr
	tam.p2pMod = p2p

	return tam
}

func (tam *TestAuxiliaryMod) RegisterHandlers() {
}

// get the txs from the txPool and propose them to the shard
func (tam *TestAuxiliaryMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			utils.LoggerInstance.Info("Stop the TestAuxiliaryMod")
			return
		default:
			time.Sleep(1 * time.Second)
			utils.LoggerInstance.Info("TestAuxiliaryMod is running")
		}
	}
}
