package auxiliaryMod

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/consensusMod/pbft"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"context"
	"sync"
	"time"
)

var _ runningModInterface.RunningMod = &ProposeBlock2ChannelAuxiliaryMod{}

// init the channel ip map
func init() {
	config.IPMap[pbft.ChannelId] = make(map[int]string)

	// The channel consists of all view nodes.
	for i := 0; i < config.ShardNum; i++ {
		config.IPMap[pbft.ChannelId][i] = config.IPMap[i][0]
	}
}

// this Mod is used with the PbftTBDAddon
// this mod will receive the txs from client, pack them to a block and propose to the shard, or propose them to channel if cross-shard
type ProposeBlock2ChannelAuxiliaryMod struct {
	nodeAttr *nodeattr.NodeAttr // the attribute of the belonging node
	p2pMod   *p2p.P2PMod        // the p2p network module of the belonging node

	txPool        *structs.TxPool
	channelTxPool *structs.TxPool
}

// this mod will receive the txs from client, pack them to a block and propose to the shard
func NewProposeBlock2ChannelAuxiliaryMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	pbcm := new(ProposeBlock2ChannelAuxiliaryMod)
	pbcm.nodeAttr = attr
	pbcm.p2pMod = p2p

	pbcm.txPool = structs.NewTxPool(config.BlockSize)
	pbcm.channelTxPool = structs.NewTxPool(config.BlockSize)

	return pbcm
}

func (pbcm *ProposeBlock2ChannelAuxiliaryMod) RegisterHandlers() {
	pbcm.p2pMod.RegisterHandler(message.MsgInject, pbcm.handleInject)
}

// get the txs from the txPool, pack them to a block and propose to the shard
func (pbcm *ProposeBlock2ChannelAuxiliaryMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			utils.LoggerInstance.Info("Stop the ProposeTxsAuxiliaryMod")
			return
		default:
			pbcm.TryProposeBlock(pbcm.txPool, pbcm.nodeAttr.Sid)
			pbcm.TryProposeBlock(pbcm.channelTxPool, pbft.ChannelId)
		}
	}
}

func (pbcm *ProposeBlock2ChannelAuxiliaryMod) TryProposeBlock(txPool *structs.TxPool, sid int) {
	txs := txPool.GetEnoughTxs(20, 200)
	if txs != nil {
		b := pbcm.nodeAttr.CurChain.NewBlock(txs)
		req := message.NewRequest(sid, message.ReqVerifyTxs, utils.Encode(b))

		msg := message.Message{
			MsgType: message.MsgPropose,
			Content: utils.Encode(req),
		}

		utils.LoggerInstance.Info("Propose block with %d txs to shard %d", len(txs), sid)
		pbcm.p2pMod.MsgHandlerMap[message.MsgPropose](&msg)
	}
	time.Sleep(100 * time.Millisecond)
}

// receive the txs from the client and add them to the txPool
func (pbcm *ProposeBlock2ChannelAuxiliaryMod) handleInject(msg *message.Message) {
	utils.LoggerInstance.Debug("handle inject")

	txs := []structs.Transaction{}
	err := utils.Decode(msg.Content, &txs)

	if err != nil || len(txs) == 0 {
		utils.LoggerInstance.Error("error decoding the txs")
		return
	}

	utils.LoggerInstance.Info("Receive %d txs", len(txs))

	for _, tx := range txs {
		ctx := tx.(*structs.ContractTransaction)
		// if the tx is cross-shard, add it to the channelTxPool
		if ctx.IsCrossShard {
			pbcm.channelTxPool.AddTxs([]structs.Transaction{ctx})
		} else {
			pbcm.txPool.AddTxs([]structs.Transaction{ctx})
		}
	}
}
