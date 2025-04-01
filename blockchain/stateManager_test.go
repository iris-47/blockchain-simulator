package blockchain

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/structs"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a mock ChainConfig
func createMockChainConfig() *config.ChainConfig {
	return &config.ChainConfig{
		ShardID: 1,
		NodeID:  1,

		ShardNum: 4,
		NodeNum:  4,

		BlockSize: 100,
	}
}

// Helper function to create a mock transaction
func createMockTransaction(form string, to string, value *big.Int) structs.Transaction {
	return structs.NewAccountTransaction(form, to, 0, value)
}

func createMockCoinbae(to string, value *big.Int) structs.Transaction {
	return structs.NewAcconutCoinbase(to, value)
}

func TestNewStateManager(t *testing.T) {
	cc := createMockChainConfig()
	db, err := rawdb.NewLevelDBDatabase(config.StoragePath+strconv.Itoa(cc.ShardID)+"_"+strconv.Itoa(cc.NodeID)+".state", 0, 1, "state", false)
	assert.NoError(t, err)
	defer db.Close()

	stm, err := NewStateManager(cc, db)
	assert.NoError(t, err)
	assert.NotNil(t, stm)

	assert.Equal(t, cc, stm.ChainConfig)
	assert.NotNil(t, stm.triedb)
	assert.Equal(t, db, stm.db)
}

func TestUpdateStates(t *testing.T) {
	cc := createMockChainConfig()
	db, err := rawdb.NewLevelDBDatabase(config.StoragePath+strconv.Itoa(cc.ShardID)+"_"+strconv.Itoa(cc.NodeID)+".state", 0, 1, "state", false)

	assert.NoError(t, err)
	defer db.Close()
	stm, _ := NewStateManager(cc, db)

	txs := []structs.Transaction{
		createMockTransaction("sender1", "receiver1", big.NewInt(100)),
		createMockTransaction("sender2", "receiver2", big.NewInt(200)),
	}

	stateRoot := trie.NewEmpty(stm.triedb).Hash().Bytes()
	success := stm.UpdateStates(txs, stateRoot)
	assert.True(t, success)
}

// func TestCommitStates(t *testing.T) {
// 	db := memorydb.New()
// 	cc := createMockChainConfig()
// 	stm, _ := NewStateManager(cc, db)

// 	// Mock a DirtyState
// 	mockState := &structs.ContractState{}
// 	mockKey := []byte("mockContract")
// 	mockStateBytes := utils.Encode(mockState)

// 	stm.DirtyState[string(mockKey)] = mockState

// 	// Set a stateRoot
// 	stateRoot := common.Hex2Bytes("1234")
// 	newRoot := stm.CommitStates(stateRoot)

// 	// Verify that the trie has been updated
// 	st, err := trie.New(trie.TrieID(common.BytesToHash(newRoot)), stm.triedb)
// 	assert.NoError(t, err)

// 	value, err := st.Get(mockKey)
// 	assert.NoError(t, err)
// 	assert.True(t, bytes.Equal(value, mockStateBytes))
// }

// func TestCommitStates_EmptyDirtyState(t *testing.T) {
// 	db := memorydb.New()
// 	cc := createMockChainConfig()
// 	stm, _ := NewStateManager(cc, db)

// 	// Set a stateRoot
// 	stateRoot := common.Hex2Bytes("1234")
// 	newRoot := stm.CommitStates(stateRoot)

// 	// Root should remain unchanged as there is no DirtyState
// 	assert.Equal(t, stateRoot, newRoot)
// }

// func TestNewStateManager_Error(t *testing.T) {
// 	// Simulate error in creating a new StateManager (e.g., nil database)
// 	cc := createMockChainConfig()
// 	stm, err := NewStateManager(cc, nil)
// 	assert.Error(t, err)
// 	assert.Nil(t, stm)
// }

// func TestUpdateStates_ErrorHandling(t *testing.T) {
// 	db := memorydb.New()
// 	cc := createMockChainConfig()
// 	stm, _ := NewStateManager(cc, db)

// 	// Mock invalid stateRoot
// 	invalidStateRoot := []byte{}
// 	txs := []structs.Transaction{
// 		createMockTransaction(structs.ETHLikeContractTransactionType, "receiver1"),
// 	}

// 	success := stm.UpdateStates(txs, invalidStateRoot)
// 	assert.False(t, success)
// }
