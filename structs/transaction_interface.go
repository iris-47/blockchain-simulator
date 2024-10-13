// Description: This file contains the Transaction interface,
// for different types of transactions should have a unified abstract representation at higher levels, such as in blocks
package structs

import (
	"time"
)

// TODO: add the implementation of public key signature method in the future

type Address = string

type Transaction interface {
	// <---Get value from Transaction--->
	// returns the transaction ID, usually the same to Hash()
	ID() []byte
	// returns the sender's address
	From() []Address
	// returns the recipient's address
	To() []Address
	// returns the time of the transaction
	GetTime() time.Time
	// returns the Hash of the transaction
	Hash() []byte
	// is the transaction a coinbase transaction
	IsCoinBase() bool
	// returns the nonce of the transaction
	GetNonce() int64

	// <---Set value to Transaction--->
	SetTime(time time.Time)
}
