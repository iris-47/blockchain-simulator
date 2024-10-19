// Description: This file use ethereum's trie to store the state of the blockchain
package blockchain

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
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

	st, err := trie.New(trie.TrieID(common.BytesToHash(stateRoot)), stm.triedb)
	if err != nil {
		utils.LoggerInstance.Error("Failed to create the trie")
		log.Panic(err)
	}

	for _, tx := range txs {
		if tx.Type() == structs.UTXOTransactionType {
			// update the UTXO set
		} else if tx.Type() == structs.AccountTransactionType {
			// update the state of the account
		} else if tx.Type() == structs.ETHLikeContractTransactionType {
			receiver := tx.To()[0]
			if Addr2Shard(receiver) == stm.ChainConfig.ShardID {
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
	st, err := trie.New(trie.TrieID(common.BytesToHash(stateRoot)), stm.triedb)
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
