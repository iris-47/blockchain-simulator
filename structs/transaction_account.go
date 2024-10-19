// Description: This file contains the implementation of ETH-like transaction.
package structs

import (
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

var _ Transaction = &AccountTransaction{}

func init() {
	gob.Register(&AccountTransaction{}) // register the AccountTransaction struct for utils.Encode()
}

type AccountTransaction struct {
	// Global
	Sender    Address
	Recipient Address
	Nounce    int64
	Value     *big.Int
	TxHash    []byte

	Time time.Time

	// used in some consensus algorithm
	Relayed        bool
	OriginalSender Address
	FinalRecipient Address

	Siganature []byte // not implemented yet
}

func (tx *AccountTransaction) Type() string {
	return AccountTransactionType
}

func (tx *AccountTransaction) ID() []byte {
	return tx.TxHash
}

func (tx *AccountTransaction) From() []Address {
	return []Address{tx.Sender}
}

func (tx *AccountTransaction) To() []Address {
	return []Address{tx.Recipient}
}

func (tx *AccountTransaction) GetTime() time.Time {
	return tx.Time
}

func (tx *AccountTransaction) Hash() []byte {
	return tx.TxHash
}

func (tx *AccountTransaction) IsCoinBase() bool {
	return false
}

func (tx *AccountTransaction) GetNonce() int64 {
	return tx.Nounce
}

func (tx *AccountTransaction) SetTime(time time.Time) {
	tx.Time = time
}

func (tx AccountTransaction) String() string {
	floatVal, _ := tx.Value.Float64()
	return fmt.Sprintf("TxId: %s: [%s-->%s:%f]", hex.EncodeToString(tx.ID()), tx.Sender, tx.Recipient, floatVal)
}
