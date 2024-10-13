package blockchain

import (
	"BlockChainSimulator/structs"
	"testing"
)

func TestGetTxTreeRoot(t *testing.T) {
	txs := []structs.Transaction{
		&structs.UTXOTransaction{
			TxId: []byte("tx1"),
			Vin:  []structs.TxIn{},
			Vout: []structs.TxOut{},
		},
		&structs.UTXOTransaction{
			TxId: []byte("tx2"),
			Vin:  []structs.TxIn{},
			Vout: []structs.TxOut{},
		},
	}

	TxRoot1 := GetTxTreeRoot(txs)
	TxRoot2 := GetTxTreeRoot(txs)

	if string(TxRoot1) != string(TxRoot2) {
		t.Errorf("GetTxTreeRoot() = %v, want %v", TxRoot1, TxRoot2)
	}
}
