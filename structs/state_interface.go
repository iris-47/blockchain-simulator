package structs

import "math/big"

type State interface {
	// Global
	Hash() []byte
	GetAddress() string
	GetNonce() int64
	GetBalance() *big.Int
	Deposit(amount *big.Int)
	Deduct(amount *big.Int) bool

	// For Smart Contract State
}
