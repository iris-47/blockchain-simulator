package blockchain

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/storage"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"log"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

type BlockChain struct {
	CurrentBlock *structs.Block // the current block

	ChainConfig  *config.ChainConfig // the chain configuration
	Storage      *storage.Storage    // Storage is the bolt-db to store the blocks
	StateManager *StateManager       // the state manager to manage the states of the blockchain
}

func NewBlockChain(cc *config.ChainConfig, db ethdb.Database) (*BlockChain, error) {
	bc := &BlockChain{
		ChainConfig: cc,
		Storage:     storage.NewStorage(cc),
	}

	db, err := rawdb.NewLevelDBDatabase(config.StoragePath+strconv.Itoa(cc.ShardID)+"_"+strconv.Itoa(cc.NodeID)+".state", 0, 1, "state", false)
	if err != nil {
		utils.LoggerInstance.Error("Failed to create the database")
		log.Panic(err)
	}

	curHash, err := bc.Storage.GetNewestBlockHash()
	if err != nil {
		// no blockchain in the storage
		if err.Error() == "cannot find the newest block hash" {
			genesisBlock := bc.NewGenisisBlock(db)
			bc.AddGenesisBlock(genesisBlock)
			return bc, nil
		} else {
			log.Panic()
		}
	}

	// there is a blockchain in the storage
	curBlock, err := bc.Storage.GetBlock(curHash)
	if err != nil {
		log.Panic()
	}
	bc.CurrentBlock = curBlock
	bc.StateManager, err = NewStateManager(cc, db)

	_, err = trie.New(trie.TrieID(common.BytesToHash(curBlock.Header.StateRoot)), bc.StateManager.triedb)
	if err != nil {
		utils.LoggerInstance.Error("Failed to create the trie")
		log.Panic(err)
	}

	utils.LoggerInstance.Info("Create the blockchain successfully")
	return bc, nil
}

func (bc *BlockChain) NewGenisisBlock(db ethdb.Database) *structs.Block {
	txs := make([]structs.Transaction, 0)
	// TODO: further config should be considered
	triedb := trie.NewDatabaseWithConfig(db, &trie.Config{
		Cache:     0,
		Preimages: true,
	})

	bh := &structs.BlockHeader{
		Nonce:     0,
		Height:    0,
		TimeStamp: time.Now(),
		TxRoot:    GetTxTreeRoot(txs),
		StateRoot: trie.NewEmpty(triedb).Hash().Bytes(),
	}
	block := structs.NewBlock(bh, txs)
	block.Hash = block.Header.Hash()
	return block
}

// add the genisis block in a blockchain
func (bc *BlockChain) AddGenesisBlock(b *structs.Block) {
	bc.Storage.AddBlock(b)
	newestHash, err := bc.Storage.GetNewestBlockHash()
	if err != nil {
		utils.LoggerInstance.Error("Get newest block hash failed")
	}
	current_b, err := bc.Storage.GetBlock(newestHash)
	if err != nil {
		utils.LoggerInstance.Error("Get newest block failed")
	}
	bc.CurrentBlock = current_b
}

// Consensus done, add the block to storage and commit the states
func (bc *BlockChain) CommitBlock(b *structs.Block) {
	b.Header.StateRoot = bc.StateManager.CommitStates(bc.CurrentBlock.Header.StateRoot)
	bc.CurrentBlock = b
	bc.Storage.AddBlock(b)
}

// To check the Txs integrity of the block
func GetTxTreeRoot(txs []structs.Transaction) []byte {
	triedb := trie.NewDatabase(rawdb.NewMemoryDatabase())
	txTree := trie.NewEmpty(triedb)
	for _, tx := range txs {
		txTree.Update(tx.Hash(), utils.Encode(tx))
	}
	return txTree.Hash().Bytes()
}
