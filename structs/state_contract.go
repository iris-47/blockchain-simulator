package structs

import (
	"encoding/gob"
	"time"

	"golang.org/x/exp/rand"
)

var execTimeDelay = 50 // mimic the execution time of a transaction

var _ State = &ContractState{}

func init() {
	gob.Register(&ContractState{})
}

type ContractState struct {
	Addr            Address
	relatedContract []Address
	IsCrossShard    bool

	Variables          map[string]string
	BackupVariablesmap map[string]string
}

func (cs *ContractState) GetKey() []byte {
	return []byte(cs.Addr)
}

// Run and verify the transaction
func (cs *ContractState) Update(tx Transaction) bool {
	// randomly delay the execution time of the transaction around execTimeDelay(ms)
	delay := rand.Intn(100) + execTimeDelay
	time.Sleep(time.Duration(delay) * time.Millisecond)

	return true
}

func (cs *ContractState) Commit() {
	cs.BackupVariablesmap = cs.Variables
}

func (cs *ContractState) Rollback() {
	cs.Variables = cs.BackupVariablesmap
}
