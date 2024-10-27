package storage

import (
	rfccode "BlockChainSimulator/addon/coding"
	"BlockChainSimulator/config"
	"BlockChainSimulator/structs"
	"BlockChainSimulator/utils"
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/boltdb/bolt"
)

type Storage struct {
	filePath string

	// bucket names
	//<--global-->
	blockBucket           string // bucket to store the block
	blockHeaderBucket     string // bucket to store the block head
	newestBlockHashBucket string // bucket to store the newest block hash
	UTXOBucket            string // bucket to store the utxoset, {txid: outputs[]}, ignore the inputs in Txs

	//<--used in cshard-->
	codingSchemaBucket string // use in cshard
	chunkBucket        string // store the rfc encoding result

	DataBase *bolt.DB
}

// Create a new storage based on boltDB
func NewStorage(cfg *config.ChainConfig) *Storage {
	_, err := os.Stat(config.StoragePath)
	if os.IsNotExist(err) {
		errMkdir := os.Mkdir(config.StoragePath, os.ModePerm)
		if errMkdir != nil {
			utils.LoggerInstance.Error("Failed to create the storage dir:%v", errMkdir)
			log.Panic(errMkdir)
		}
	} else if err != nil {
		utils.LoggerInstance.Error("Failed to check the storage dir:%v", err)
		log.Panic(err)
	}

	s := &Storage{
		filePath:              config.StoragePath + "/" + strconv.Itoa(cfg.ShardID) + "_" + strconv.Itoa(cfg.NodeID) + ".data",
		blockBucket:           "block",
		blockHeaderBucket:     "blockHeader",
		newestBlockHashBucket: "newestBlockHash",
		UTXOBucket:            "UTXO",
		codingSchemaBucket:    "codingSchema",
		chunkBucket:           "chunk",
	}

	db, err := bolt.Open(s.filePath, 0600, nil)
	if err != nil {
		utils.LoggerInstance.Error("Failed to open the database:%v", err)
		log.Panic(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(s.blockBucket))
		if err != nil {
			utils.LoggerInstance.Error("Failed to create the block bucket:%v", err)
			log.Panic("create blocksBucket failed")
		}

		_, err = tx.CreateBucketIfNotExists([]byte(s.blockHeaderBucket))
		if err != nil {
			utils.LoggerInstance.Error("Failed to create the block header bucket:%v", err)
			log.Panic("create blockHeaderBucket failed")
		}

		_, err = tx.CreateBucketIfNotExists([]byte(s.newestBlockHashBucket))
		if err != nil {
			utils.LoggerInstance.Error("Failed to create the newest block hash bucket:%v", err)
			log.Panic("create newestBlockHashBucket failed")
		}

		_, err = tx.CreateBucketIfNotExists([]byte(s.UTXOBucket))
		if err != nil {
			utils.LoggerInstance.Error("Failed to create the UTXO bucket:%v", err)
			log.Panic("create utxoBucket failed")
		}

		_, err = tx.CreateBucketIfNotExists([]byte(s.codingSchemaBucket))
		if err != nil {
			utils.LoggerInstance.Error("Failed to create the coding schema bucket:%v", err)
			log.Panic("create codingSchemaBucket failed")
		}

		_, err = tx.CreateBucketIfNotExists([]byte(s.chunkBucket))
		if err != nil {
			utils.LoggerInstance.Error("Failed to create the chunk bucket:%v", err)
			log.Panic("create chunkBucket failed")
		}
		return nil
	})
	s.DataBase = db
	return s
}

// update the newest block in the database
func (s *Storage) UpdateNewestBlockHash(newestbhash []byte) {
	err := s.DataBase.Update(func(tx *bolt.Tx) error {
		nbhBucket := tx.Bucket([]byte(s.newestBlockHashBucket))
		err := nbhBucket.Put([]byte("theNewestBlockHash"), newestbhash)
		if err != nil {
			utils.LoggerInstance.Error("Failed to update the newest block hash:%v", err)
			log.Panic()
		}
		return nil
	})
	if err != nil {
		utils.LoggerInstance.Error("Failed to update the newest block hash:%v", err)
		log.Panic()
	}
}

// add a blockheader into the database
func (s *Storage) AddBlockHeader(blockhash []byte, bh *structs.BlockHeader) {
	err := s.DataBase.Update(func(tx *bolt.Tx) error {
		bhbucket := tx.Bucket([]byte(s.blockHeaderBucket))
		err := bhbucket.Put(blockhash, utils.Encode(bh))
		if err != nil {
			utils.LoggerInstance.Error("Failed to add the block header:%v", err)
			log.Panic()
		}
		return nil
	})
	if err != nil {
		utils.LoggerInstance.Error("Failed to add the block header:%v", err)
		log.Panic()
	}
}

// add a block into the database
func (s *Storage) AddBlock(b *structs.Block) {
	err := s.DataBase.Update(func(tx *bolt.Tx) error {
		bbucket := tx.Bucket([]byte(s.blockBucket))
		err := bbucket.Put(b.Hash, utils.Encode(b))
		if err != nil {
			utils.LoggerInstance.Error("Failed to add the block:%v", err)
			log.Panic()
		}
		return nil
	})
	if err != nil {
		utils.LoggerInstance.Error("Failed to add the block:%v", err)
		log.Panic()
	}
	s.AddBlockHeader(b.Hash, b.Header)
	s.UpdateNewestBlockHash(b.Hash)
}

// read a blockheader from the database
func (s *Storage) GetBlockHeader(bhash []byte) (*structs.BlockHeader, error) {
	var res structs.BlockHeader
	err := s.DataBase.View(func(tx *bolt.Tx) error {
		bhbucket := tx.Bucket([]byte(s.blockHeaderBucket))
		bh_encoded := bhbucket.Get(bhash)
		if bh_encoded == nil {
			return errors.New("the block is not existed")
		}
		utils.Decode(bh_encoded, &res)
		return nil
	})
	return &res, err
}

// read a block from the database
func (s *Storage) GetBlock(bhash []byte) (*structs.Block, error) {
	var res structs.Block
	err := s.DataBase.View(func(tx *bolt.Tx) error {
		bbucket := tx.Bucket([]byte(s.blockBucket))
		b_encoded := bbucket.Get(bhash)
		if b_encoded == nil {
			return errors.New("the block is not existed")
		}
		utils.Decode(b_encoded, &res)
		return nil
	})
	return &res, err
}

// read the Newest block hash
func (s *Storage) GetNewestBlockHash() ([]byte, error) {
	var nhb []byte
	err := s.DataBase.View(func(tx *bolt.Tx) error {
		bhbucket := tx.Bucket([]byte(s.newestBlockHashBucket))
		// the bucket has the only key "theNewestBlockHash"
		nhb = bhbucket.Get([]byte("theNewestBlockHash"))
		if nhb == nil {
			return errors.New("cannot find the newest block hash")
		}
		return nil
	})
	return nhb, err
}

func (s *Storage) SaveCodingSchema(schema *rfccode.CodingSchema) {
	err := s.DataBase.Update(func(tx *bolt.Tx) error {
		bbucket := tx.Bucket([]byte(s.codingSchemaBucket))

		err := bbucket.Put(utils.Hash(schema), utils.Encode(schema))
		if err != nil {
			utils.LoggerInstance.Error("Failed to save the coding schema:%v", err)
			log.Panic()
		}
		return nil
	})
	if err != nil {
		utils.LoggerInstance.Error("Failed to save the coding schema:%v", err)
		log.Panic()
	}
}

func (s *Storage) SaveChunks(chunk *rfccode.Chunk) {
	err := s.DataBase.Update(func(tx *bolt.Tx) error {
		bbucket := tx.Bucket([]byte(s.codingSchemaBucket))

		err := bbucket.Put(utils.Hash(chunk), utils.Encode(chunk))
		if err != nil {
			utils.LoggerInstance.Error("Failed to save the chunk:%v", err)
			log.Panic()
		}
		return nil
	})
	if err != nil {
		utils.LoggerInstance.Error("Failed to save the chunk:%v", err)
		log.Panic()
	}
}
