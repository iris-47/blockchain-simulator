package structs

import (
	"BlockChainSimulator/utils"
	"crypto/sha256"
	"encoding/gob"
	"math/big"
)

var _ State = &AccountState{}

func init() {
	gob.Register(&AccountState{})
}

type AccountState struct {
	AcAddress    Address
	Nonce        int64
	Balance      *big.Int
	DirtyBalance *big.Int
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
	as.Balance.Add(as.DirtyBalance, amount)
}

func (as *AccountState) Deduct(amount *big.Int) bool {
	if as.DirtyBalance.Cmp(amount) < 0 {
		return false
	}
	as.DirtyBalance.Sub(as.DirtyBalance, amount)
	return true
}

func (as *AccountState) GetKey() []byte {
	return []byte(as.AcAddress)
}

func (as *AccountState) Update(tx Transaction) bool {
	return true
}

func (as *AccountState) Rollback() {
	as.DirtyBalance = as.Balance
}

func (as *AccountState) Commit() {
	as.Balance = as.DirtyBalance
}
