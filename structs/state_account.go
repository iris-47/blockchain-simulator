package structs

import (
	"BlockChainSimulator/utils"
	"crypto/sha256"
	"math/big"
)

var _ State = &AccountState{}

type AccountState struct {
	AcAddress Address
	Nonce     int64
	Balance   *big.Int
}

func (as *AccountState) Hash() []byte {
	hash := sha256.Sum256(utils.Encode(as))
	return hash[:]
}

func (as *AccountState) GetAddress() Address {
	return as.AcAddress
}

func (as *AccountState) GetNonce() int64 {
	return as.Nonce
}

func (as *AccountState) GetBalance() *big.Int {
	return as.Balance
}

func (as *AccountState) Deposit(amount *big.Int) {
	as.Balance.Add(as.Balance, amount)
}

func (as *AccountState) Deduct(amount *big.Int) bool {
	if as.Balance.Cmp(amount) < 0 {
		return false
	}
	as.Balance.Sub(as.Balance, amount)
	return true
}
