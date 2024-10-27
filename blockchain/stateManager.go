// Description: This file use ethereum's trie to store the state of the blockchain
package blockchain

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"golang.org/x/exp/rand"
)

type StateManager struct {
	ChainConfig *config.ChainConfig // the chain configuration

	// Used to store the state of Ethereum-like accounts and contracts
	triedb *trie.Database // middle layer to cache tries, to avoid frequent disk access
	db     ethdb.Database // the leveldb database to store tries in the disk

	// Used to store the Bitcoin-like UTXO set
	UTXOSet *UTXOSet // the UTXO set of the blockchain

	DirtyState map[string]structs.State // the state which has been updated but not committed
}

func NewStateManager(cc *config.ChainConfig, db ethdb.Database) (*StateManager, error) {
	stm := &StateManager{
		ChainConfig: cc,
		db:          db,
	}

	stm.triedb = trie.NewDatabaseWithConfig(stm.db, &trie.Config{Cache: 10})
	stm.DirtyState = make(map[string]structs.State)

	return stm, nil
}

// Check the txs and update the states to the DirtyState
func (stm *StateManager) UpdateStates(txs []structs.Transaction, stateRoot []byte) bool {
	flag := true
	if config.TxVerifyTime {
		// randomly delay the execution time of the transaction around execTimeDelay(ms)
		delay := rand.Intn(config.ExecTimeDelay*2) + config.ExecTimeDelay
		time.Sleep(time.Duration(delay) * time.Millisecond * time.Duration(len(txs)))

		return flag
	}

	st, err := trie.New(trie.TrieID(common.BytesToHash(stateRoot)), stm.triedb)
	if err != nil {
		utils.LoggerInstance.Error("Failed to create the trie")
		log.Panic(err)
	}

	for _, tx := range txs {
		if tx.Type() == structs.UTXOTransactionType {
			// update the UTXO set
		} else if tx.Type() == structs.AccountTransactionType {
			acTx := tx.(*structs.AccountTransaction)

			// update the state of the sender
			if utils.Addr2Shard(acTx.Sender) == stm.ChainConfig.ShardID {
				stm.UpdateAccountState(acTx.Sender, acTx.Value, st, false)
			}

			// update the state of the receiver
			if utils.Addr2Shard(acTx.Recipient) == stm.ChainConfig.ShardID {
				stm.UpdateAccountState(acTx.Recipient, acTx.Value, st, true)
			}
		} else if tx.Type() == structs.ETHLikeContractTransactionType {
			receiver := tx.To()[0]
			if utils.Addr2Shard(receiver) == stm.ChainConfig.ShardID {
				// update the state of the sender
				var state *structs.ContractState
				// if the state is not in the DirtyState, get the state from the trie
				if stm.DirtyState[receiver] == nil {
					stateBytes, _ := st.Get([]byte(receiver))
					if stateBytes == nil {
						// the contract has not been deployed
						utils.LoggerInstance.Error("The contract %s has not been deployed", receiver)
						continue
					}
					utils.Decode(stateBytes, state)
				} else {
					state = stm.DirtyState[receiver].(*structs.ContractState)
				}
				if !state.Update(tx) {
					utils.LoggerInstance.Info("Failed to update the state of the contract %s", receiver)
					flag = false
				}
				stm.DirtyState[string(state.GetKey())] = state // update the state in the DirtyState
			}
		}
	}

	return flag
}

// Consensus Passed, Commit the states
func (stm *StateManager) CommitStates(stateRoot []byte) []byte {
	if config.TxVerifyTime {
		return stateRoot
	}
	st, err := trie.New(trie.TrieID(common.BytesToHash(stateRoot)), stm.triedb)
	if err != nil {
		utils.LoggerInstance.Error("Failed to create the trie")
		log.Panic(err)
	}
	for key := range stm.DirtyState {
		state := stm.DirtyState[key]
		state.Commit()
		st.Update(state.GetKey(), utils.Encode(state))
	}

	rootHash, nodeSet := st.Commit(false)

	err = stm.triedb.Update(trie.NewWithNodeSet(nodeSet))
	if err != nil {
		utils.LoggerInstance.Error("Failed to update the trie")
		log.Panic(err)
	}
	err = stm.triedb.Commit(rootHash, false)
	if err != nil {
		utils.LoggerInstance.Error("Failed to commit the trie")
		log.Panic(err)
	}

	return rootHash.Bytes()
}

func (stm *StateManager) UpdateAccountState(addr string, amount *big.Int, st *trie.Trie, deposit bool) {
	// update the state of the sender
	var state *structs.AccountState

	// get the current account state
	if stm.DirtyState[addr] == nil {
		stateBytes, _ := st.Get([]byte(addr))
		if stateBytes == nil {
			// the account has not been created
			utils.LoggerInstance.Info("The account %s has not been created, now create it", addr)
			// create a new account
			state = &structs.AccountState{
				AcAddress: addr,
				Nonce:     0,
				Balance:   config.Init_Balance,
			}
		} else {
			utils.Decode(stateBytes, state)
		}
	} else {
		state = stm.DirtyState[addr].(*structs.AccountState)
	}

	// update the balance of the account state
	if deposit {
		state.Deposit(amount)
	} else {
		if !state.Deduct(amount) {
			utils.LoggerInstance.Warn("Failed to deduct the balance of the account %s, insufficient funds", addr)
		}
	}

	stm.DirtyState[string(state.GetKey())] = state // update the state in the DirtyState
}
