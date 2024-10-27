package auxiliaryMod

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

// Q: there will be a lot of duplicated code, how to avoid it?

var _ runningModInterface.RunningMod = &ProposeBlockAuxiliaryMod{}

// this mod will receive the txs from client, pack them to a block and propose to the shard
type ProposeBlockAuxiliaryMod struct {
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node

	txPool *structs.TxPool
}

// this mod will receive the txs from client, pack them to a block and propose to the shard
func NewProposeBlockAuxiliaryMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	pbm := new(ProposeBlockAuxiliaryMod)
	pbm.nodeAttr = attr
	pbm.p2pMod = p2p

	pbm.txPool = structs.NewTxPool(config.BlockSize)

	return pbm
}

func (pbm *ProposeBlockAuxiliaryMod) RegisterHandlers() {
	pbm.p2pMod.RegisterHandler(message.MsgInject, pbm.handleInject)
}

// get the txs from the txPool, pack them to a block and propose to the shard
func (pbm *ProposeBlockAuxiliaryMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			utils.LoggerInstance.Info("Stop the ProposeBlockAuxiliaryMod")
			return
		default:
			pbm.TryProposeBlock()
		}
	}
}

func (pbm *ProposeBlockAuxiliaryMod) TryProposeBlock() {
	txs := pbm.txPool.GetEnoughTxs(100, 500)
	if txs != nil {
		b := pbm.nodeAttr.CurChain.NewBlock(txs)
		req := message.NewRequest(pbm.nodeAttr.Sid, message.ReqVerifyTxs, utils.Encode(b))

		msg := message.Message{
			MsgType: message.MsgPropose,
			Content: utils.Encode(req),
		}

		utils.LoggerInstance.Info("Propose block with %d txs to shard %d", len(txs), pbm.nodeAttr.Sid)
		pbm.p2pMod.MsgHandlerMap[message.MsgPropose](&msg)
	}
	time.Sleep(100 * time.Millisecond)
}

// receive the txs from the client and add them to the txPool
func (pbm *ProposeBlockAuxiliaryMod) handleInject(msg *message.Message) {
	utils.LoggerInstance.Debug("handle inject")

	txs := []structs.Transaction{}
	err := utils.Decode(msg.Content, &txs)

	if err != nil || len(txs) == 0 {
		utils.LoggerInstance.Error("error decoding the txs")
		return
	}

	utils.LoggerInstance.Info("Receive %d txs", len(txs))
	pbm.txPool.AddTxs(txs)
}
