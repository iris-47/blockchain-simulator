package clientMod

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"context"
	"math/big"
	"sync"
	"time"
)

var _ runningModInterface.RunningMod = &sendMimicAccountTxsMod{}

// just for test use, this mod sends Txs every 3 seconds
type sendMimicAccountTxsMod struct {
	nodeAttr *nodeattr.NodeAttr
	p2pMod   *p2p.P2PMod
}

// just for test use, this mod sends Txs every 3 seconds
func NewsendMimicAccountTxsMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	smatm := new(sendMimicAccountTxsMod)
	smatm.nodeAttr = attr
	smatm.p2pMod = p2p

	return smatm
}

func (smatm *sendMimicAccountTxsMod) RegisterHandlers() {
}

func (smatm *sendMimicAccountTxsMod) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	// wait for the system to start
	if !p2p.WaitForAllIPsReady(10 * time.Second) {
		utils.LoggerInstance.Error("Wait for all IPs ready timeout")
		return
	}
	utils.LoggerInstance.Info("All IPs are ready, start to send txs")
	// generate mimic contract txs and send it according to the config.TxInjectSpeed
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			utils.LoggerInstance.Debug("Try to send mimic contract txs")
			txs := generateMimicAccountTxs()

			txsToSend := make(map[int][]structs.Transaction) // key: sid, value: txs
			for _, tx := range txs {
				sid := utils.Addr2Shard(tx.To()[0])
				if _, ok := txsToSend[sid]; !ok {
					txsToSend[sid] = make([]structs.Transaction, 0)
				}
				txsToSend[sid] = append(txsToSend[sid], tx)
			}

			// send the txs to the corresponding shard
			for sid, txs := range txsToSend {
				msg := message.Message{
					MsgType: message.MsgInject,
					Content: utils.Encode(txs),
				}
				utils.LoggerInstance.Debug("send txs to shard %v, len %v", sid, len(txs))
				smatm.p2pMod.ConnMananger.Send(config.IPMap[sid][0], msg.JsonEncode())
			}
		}
	}
}

// generate config.TxInjectSpeed of mimic contract
func generateMimicAccountTxs() []structs.Transaction {
	txs := make([]structs.Transaction, 0)
	for i := 0; i < config.TxInjectSpeed; i++ {
		tx := structs.NewAccountTransaction(
			randomAddr(),
			randomAddr(),
			0,
			big.NewInt(0),
		)

		txs = append(txs, tx)
	}

	return txs
}
