package consensusMod

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"context"
	"sync"
	"time"
)

var _ runningModInterface.RunningMod = &ProposeTxsAuxiliaryMod{}

// this mod will receive the txs from client and propose them to the shard
type ProposeTxsAuxiliaryMod struct {
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node

	txPool *structs.TxPool
}

// this mod will receive the txs from client and propose them to the shard
func NewProposeTxsAuxiliaryMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	sam := new(ProposeTxsAuxiliaryMod)
	sam.nodeAttr = attr
	sam.p2pMod = p2p

	sam.txPool = structs.NewTxPool(config.BlockSize)

	return sam
}

func (sam *ProposeTxsAuxiliaryMod) RegisterHandlers() {
	sam.p2pMod.RegisterHandler(message.MsgInject, sam.handleInject)
}

// get the txs from the txPool and propose them to the shard
func (sam *ProposeTxsAuxiliaryMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			utils.LoggerInstance.Info("Stop the ProposeTxsAuxiliaryMod")
			return
		default:
			txs := sam.txPool.GetBatchofTxs()
			if txs == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			req := message.NewRequest(sam.nodeAttr.Sid, message.ReqVerifyTxs, utils.Encode(txs))

			msg := message.Message{
				MsgType: message.MsgPropose,
				Content: utils.Encode(req),
			}
			sam.p2pMod.MsgHandlerMap[message.MsgPropose](&msg)
		}
	}
}

// receive the txs from the client and add them to the txPool
func (sam *ProposeTxsAuxiliaryMod) handleInject(msg *message.Message) {
	utils.LoggerInstance.Debug("handle inject")

	txs := []structs.Transaction{}
	err := utils.Decode(msg.Content, &txs)

	if err != nil || len(txs) == 0 {
		utils.LoggerInstance.Error("error decoding the txs")
		return
	}

	utils.LoggerInstance.Debug("Received txs len: %d", len(txs))
	sam.txPool.AddTxs(txs)
}
