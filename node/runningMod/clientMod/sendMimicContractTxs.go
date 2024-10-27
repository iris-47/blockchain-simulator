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
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"sync"
	"time"
)

var _ runningModInterface.RunningMod = &sendMimicContractTxsMod{}

// just for test use, this mod sends Txs every 3 seconds
type sendMimicContractTxsMod struct {
	nodeAttr *nodeattr.NodeAttr
	p2pMod   *p2p.P2PMod
}

// just for test use, this mod sends Txs every 3 seconds
func NewSendMimicContractTxsMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	smctm := new(sendMimicContractTxsMod)
	smctm.nodeAttr = attr
	smctm.p2pMod = p2p

	return smctm
}

func (smctm *sendMimicContractTxsMod) RegisterHandlers() {
}

func (smctm *sendMimicContractTxsMod) Run(ctx context.Context, wg *sync.WaitGroup) {
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
			txs := generateMimicContractTxs()

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
				smctm.p2pMod.ConnMananger.Send(config.IPMap[sid][0], msg.JsonEncode())
			}
		}
	}
}

// generate config.TxInjectSpeed of mimic contract txs
func generateMimicContractTxs() []structs.Transaction {
	crossShardRatio := 0.5 // cross shard txs ratio
	txs := make([]structs.Transaction, 0)
	for i := 0; i < config.TxInjectSpeed; i++ {
		isCrossShard := false
		if randFloat64() < crossShardRatio {
			isCrossShard = true
		}

		tx := structs.NewContractTransaction(
			randomAddr(),
			randomAddr(),
			0,
			time.Now(),
			[]byte("code.."),
			[]string{},
			isCrossShard,
		)

		txs = append(txs, tx)
	}

	return txs
}

// generate float64 number in [0, 1)
func randFloat64() float64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<53)) // 53 bits for float64 precision
	if err != nil {
		return 0
	}
	return float64(n.Int64()) / (1 << 53)
}

// A random ETH-like address
func randomAddr() string {
	address := make([]byte, 20) // address length is 20
	_, err := rand.Read(address)
	if err != nil {
		utils.LoggerInstance.Error("error generating random address")
	}

	return hex.EncodeToString(address)
}
