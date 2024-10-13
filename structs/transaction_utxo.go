// Description: This file contains the implementation of BTC-like transaction.
package structs

import (
	"BlockChainSimulator/utils"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

var _ Transaction = &UTXOTransaction{}

func init() {
	gob.Register(&UTXOTransaction{}) // register the UTXOTransaction struct for utils.Encode()
}

// TxIn defines a transaction input.
type TxIn struct {
	PrevTxId []byte
	Value    big.Float
	Index    int     // The index of the Vouts in PrevTx, named as Vout in btcd
	Addr     Address // currently implement as Address
}

// TXOut represents a transaction output
type TxOut struct {
	Value big.Float
	Addr  Address // currently implement as "AddressHash"
}

// UTXOTransaction represents a BTC-like UTXO transaction
type UTXOTransaction struct {
	TxId []byte // The identifier of Tx, always Hash
	Vin  []TxIn
	Vout []TxOut

	Nonce      int64 // random number
	IsCoinbase bool

	Time time.Time // the time when Tx adding to pool
	// CommitTime time.Time
}

func (tx *UTXOTransaction) ID() []byte {
	return tx.TxId
}

func (tx *UTXOTransaction) From() []Address {
	result := make([]Address, 0)
	for _, vin := range tx.Vin {
		result = append(result, vin.Addr)
	}
	return result
}

func (tx *UTXOTransaction) To() []Address {
	result := make([]Address, 0)
	for _, vout := range tx.Vout {
		result = append(result, vout.Addr)
	}
	return result
}

func (tx *UTXOTransaction) GetTime() time.Time {
	return tx.Time
}

func (tx *UTXOTransaction) Hash() []byte {
	return utils.Hash(tx)
}

func (tx *UTXOTransaction) IsCoinBase() bool {
	return tx.IsCoinbase
}

func (tx *UTXOTransaction) GetNonce() int64 {
	return tx.Nonce
}

func (tx *UTXOTransaction) SetTime(time time.Time) {
	tx.Time = time
}

// This function is moved to package client
// func NewTx(coinbase bool) *UTXOTransaction {
//
// }

func (tin TxIn) String() string {
	floatVal, _ := tin.Value.Float64()
	return fmt.Sprintf("[%s:%f-->%s:%d]", tin.Addr, floatVal, hex.EncodeToString(tin.PrevTxId), tin.Index)
}

func (tout TxOut) String() string {
	floatVal, _ := tout.Value.Float64()
	return fmt.Sprintf("[%s:%f]", tout.Addr, floatVal)
}

func (tx *UTXOTransaction) String() string {
	str := ""
	str += fmt.Sprintf("TxId: %s", hex.EncodeToString(tx.TxId))
	str += ", Vin:"
	str += fmt.Sprint(tx.Vin)
	str += ", Vout:"
	str += fmt.Sprint(tx.Vout)
	str += ", IsCoinbase:"
	str += fmt.Sprint(tx.IsCoinbase)
	return str
}
