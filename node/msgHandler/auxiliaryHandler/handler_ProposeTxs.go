package auxiliaryhandler

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/msgHandler/msgHandlerInterface"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"time"
)

var _ msgHandlerInterface.MsgHandlerMod = &ProposeTxsAuxiliaryMod{}

// this mod will receive the txs from client and propose them to the shard
type ProposeTxsAuxiliaryMod struct {
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node

	txPool *structs.TxPool
}

// this mod will receive the txs from client and propose them to the shard
func NewProposeTxsAuxiliaryMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) msgHandlerInterface.MsgHandlerMod {
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
func (sam *ProposeTxsAuxiliaryMod) Run() {
	for {
		txs := sam.txPool.WaitTxs()
		if len(txs) < config.BlockSize {
			utils.LoggerInstance.Error("No txs, something wrong")
		}

		req := message.Request{
			ReqType: message.ReqVerifyTxs,
			Content: utils.Encode(txs),
			ReqTime: time.Now(),
		}
		req.CalDigest()

		msg := message.Message{
			MsgType: message.MsgPropose,
			Content: utils.Encode(req),
		}
		sam.p2pMod.MsgHandlerMap[message.MsgPropose](&msg)

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

	utils.LoggerInstance.Debug("Received txs: %v", txs)
	sam.txPool.AddTxs(txs)
}
