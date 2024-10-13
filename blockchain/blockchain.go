package blockchain

import (
	"BlockChainSimulator/config"
	"BlockChainSimulator/storage"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/triedb"
	"golang.org/x/exp/rand"
)

type BlockChain struct {
	CurrentBlock *structs.Block // the current block

	ChainConfig *config.ChainConfig // the chain configuration
	Storage     *storage.Storage    // Storage is the bolt-db to store the blocks
	db          ethdb.Database      // the leveldb database to store in the disk, for status trie(only for account model)
	triedb      *triedb.Database    // the trie database which helps to store the status trie(only for account model)
	// triedb := triedb.NewDatabase(db, cacheConfig.triedbConfig(genesis != nil && genesis.IsVerkle()))
}

func NewBlockChain(cc *config.ChainConfig, db ethdb.Database) (*BlockChain, error) {
	bc := &BlockChain{
		db:          db,
		ChainConfig: cc,
		Storage:     storage.NewStorage(cc),
	}

	curHash, err := bc.Storage.GetNewestBlockHash()
	if err != nil {
		// no blockchain in the storage
		if err.Error() == "cannot find the newest block hash" {
			genesisBlock := bc.NewGenisisBlock()
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

	// TODO: further config should be considered
	bc.triedb = triedb.NewDatabase(db, &triedb.Config{
		Preimages: true,
	})

	_, err = trie.New(trie.TrieID(common.BytesToHash(curBlock.Header.StateRoot)), bc.triedb)
	if err != nil {
		log.Panic(err)
	}
	utils.LoggerInstance.Info("Create the blockchain successfully")
	return bc, nil
}

func (bc *BlockChain) NewGenisisBlock() *structs.Block {
	txs := make([]structs.Transaction, 0)
	// TODO: further config should be considered
	triedb := triedb.NewDatabase(bc.db, &triedb.Config{
		Preimages: true,
	})
	bc.triedb = triedb

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

func (bc *BlockChain) AddBlock(b *structs.Block) {
	if b.Header.Miner != bc.ChainConfig.NodeID && config.Transaction_Method != "UTXO" {
		bc.UpdateStatusTrie(b.Transactions)
	}
	bc.CurrentBlock = b
	bc.Storage.AddBlock(b)
}

// Update the status through a Account-type-transaction
func (bc *BlockChain) UpdateAccountStatus(t structs.Transaction, st *trie.Trie) bool {
	tx := t.(*structs.AccountTransaction)
	result := false
	if !tx.Relayed && (Addr2Shard(tx.Sender) == bc.ChainConfig.ShardID) {
		s_state_enc, _ := st.Get([]byte(tx.Sender))
		var s_state *structs.AccountState
		if s_state_enc == nil { // the account is not exist, create a new one
			ib := new(big.Int)
			ib.Add(ib, config.Init_Balance)
			randomGenerator := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
			s_state = &structs.AccountState{
				Nonce:   randomGenerator.Int63(),
				Balance: ib,
			}
		} else { // the account exist
			utils.Decode(s_state_enc, s_state)
		}
		s_balance := s_state.Balance
		if s_balance.Cmp(tx.Value) == -1 {
			utils.LoggerInstance.Info("Tx {%v} is invalid, the balance is not enough", tx)
			return false
		}
		s_state.Deduct(tx.Value)
		st.Update([]byte(tx.Sender), utils.Encode(s_state))
		result = true
	}
	if Addr2Shard(tx.Recipient) == bc.ChainConfig.ShardID {
		// modify local state
		r_state_enc, _ := st.Get([]byte(tx.Recipient))
		var r_state *structs.AccountState
		if r_state_enc == nil {
			ib := new(big.Int)
			ib.Add(ib, config.Init_Balance)
			randomGenerator := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
			r_state = &structs.AccountState{
				Nonce:   randomGenerator.Int63(),
				Balance: ib,
			}
		} else {
			utils.Decode(r_state_enc, r_state)
		}
		r_state.Deposit(tx.Value)
		st.Update([]byte(tx.Recipient), utils.Encode(r_state))
		result = true
	}
	return result
}

func (bc *BlockChain) UpdateStatusTrie(txs []structs.Transaction) []byte {
	// 	// empty block
	// 	if len(txs) == 0 {
	// 		return bc.CurrentBlock.Header.StateRoot
	// 	}

	// st, err := trie.New(trie.StateTrieID(common.BytesToHash(bc.CurrentBlock.Header.StateRoot)), bc.triedb)
	//
	//	if err != nil {
	//		log.Panic(err)
	//	}
	//
	// isProcessed := false
	//
	//	for _, tx := range txs {
	//		isProcessed = isProcessed || bc.UpdateAccountStatus(tx, st)
	//	}
	//
	// // no transaction is processed
	//
	//	if !isProcessed {
	//		return bc.CurrentBlock.Header.StateRoot
	//	}
	//
	// rt, ns := st.Commit(false)
	// err = bc.triedb.Update(trie.NewWithNodeSet(ns))
	//
	//	if err != nil {
	//		log.Panic()
	//	}
	//
	// err = bc.triedb.Commit(rt, false)
	//
	//	if err != nil {
	//		log.Panic(err)
	//	}
	//
	// fmt.Println("modified account number is ", cnt)

	// return rt
	return nil
}

// To check the Txs integrity of the block
func GetTxTreeRoot(txs []structs.Transaction) []byte {
	triedb := triedb.NewDatabase(rawdb.NewMemoryDatabase(), nil)
	txTree := trie.NewEmpty(triedb)
	for _, tx := range txs {
		txTree.Update(tx.Hash(), utils.Encode(tx))
	}
	return txTree.Hash().Bytes()
}
