package auxiliaryMod

import (
	"BlockChainSimulator/message"
	"BlockChainSimulator/node/nodeattr"
	"BlockChainSimulator/node/p2p"
	"BlockChainSimulator/node/runningMod/runningModInterface"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"math/big"
	"time"
)

var _ runningModInterface.RunningMod = &sendTxTestMod{}

// just for test use, this mod sends Txs every 3 seconds
type sendTxTestMod struct {
	nodeAttr *nodeattr.NodeAttr
	p2pMod   *p2p.P2PMod

	txPool structs.TxPool
}

// just for test use, this mod sends Txs every 3 seconds
func NewTestAuxiliaryMod(attr *nodeattr.NodeAttr, p2p *p2p.P2PMod) runningModInterface.RunningMod {
	sttm := new(sendTxTestMod)
	sttm.nodeAttr = attr
	sttm.p2pMod = p2p

	sttm.txPool = structs.TxPool{}

	return sttm
}

func (sttm *sendTxTestMod) RegisterHandlers() {

}

func (sttm *sendTxTestMod) Run() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			txs := make([]structs.Transaction, 0)
			txs = append(txs, &structs.UTXOTransaction{
				TxId:       []byte("txid1"),
				Vin:        []structs.TxIn{{Addr: "addr1", Value: *big.NewFloat(10)}},
				Vout:       []structs.TxOut{{Addr: "addr2", Value: *big.NewFloat(7)}, {Addr: "addr3", Value: *big.NewFloat(3)}},
				Nonce:      123,
				IsCoinbase: false,
			})
			txs = append(txs, &structs.UTXOTransaction{
				TxId:       []byte("txid2"),
				Vin:        []structs.TxIn{{Addr: "addr3", Value: *big.NewFloat(101)}},
				Vout:       []structs.TxOut{{Addr: "addr4", Value: *big.NewFloat(71)}, {Addr: "addr3", Value: *big.NewFloat(22)}},
				Nonce:      1233,
				IsCoinbase: false,
			})

			msg := message.Message{
				MsgType: message.MsgStop,
				Content: utils.Encode(txs),
			}

			sttm.p2pMod.ConnMananger.Send("127.0.0.1:10001", msg.JsonEncode())
		}
	}
}
