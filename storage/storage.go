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
			log.Panic(errMkdir)
		}
	} else if err != nil {
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
		log.Panic(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(s.blockBucket))
		if err != nil {
			log.Panic("create blocksBucket failed")
		}

		_, err = tx.CreateBucketIfNotExists([]byte(s.blockHeaderBucket))
		if err != nil {
			log.Panic("create blockHeaderBucket failed")
		}

		_, err = tx.CreateBucketIfNotExists([]byte(s.newestBlockHashBucket))
		if err != nil {
			log.Panic("create newestBlockHashBucket failed")
		}

		_, err = tx.CreateBucketIfNotExists([]byte(s.UTXOBucket))
		if err != nil {
			log.Panic("create utxoBucket failed")
		}

		_, err = tx.CreateBucketIfNotExists([]byte(s.codingSchemaBucket))
		if err != nil {
			log.Panic("create codingSchemaBucket failed")
		}

		_, err = tx.CreateBucketIfNotExists([]byte(s.chunkBucket))
		if err != nil {
			log.Panic("create chunkBucket failed")
		}
		return nil
	})
	s.DataBase = db
	return s
}

// update the newest block in the database
func (s *Storage) UpdateNewestBlock(newestbhash []byte) {
	err := s.DataBase.Update(func(tx *bolt.Tx) error {
		nbhBucket := tx.Bucket([]byte(s.newestBlockHashBucket))
		err := nbhBucket.Put([]byte("theNewestBlock"), newestbhash)
		if err != nil {
			log.Panic()
		}
		return nil
	})
	if err != nil {
		log.Panic()
	}
}

// add a blockheader into the database
func (s *Storage) AddBlockHeader(blockhash []byte, bh *structs.BlockHeader) {
	err := s.DataBase.Update(func(tx *bolt.Tx) error {
		bhbucket := tx.Bucket([]byte(s.blockHeaderBucket))
		err := bhbucket.Put(blockhash, utils.Encode(bh))
		if err != nil {
			log.Panic()
		}
		return nil
	})
	if err != nil {
		log.Panic()
	}
}

// add a block into the database
func (s *Storage) AddBlock(b *structs.Block) {
	err := s.DataBase.Update(func(tx *bolt.Tx) error {
		bbucket := tx.Bucket([]byte(s.blockBucket))
		err := bbucket.Put(b.Hash, utils.Encode(b))
		if err != nil {
			log.Panic()
		}
		return nil
	})
	if err != nil {
		log.Panic()
	}
	s.AddBlockHeader(b.Hash, b.Header)
	s.UpdateNewestBlock(b.Hash)
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
		// the bucket has the only key "OnlyNewestBlock"
		nhb = bhbucket.Get([]byte("OnlyNewestBlock"))
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
			log.Panic()
		}
		return nil
	})
	if err != nil {
		log.Panic()
	}
}

func (s *Storage) SaveChunks(chunk *rfccode.Chunk) {
	err := s.DataBase.Update(func(tx *bolt.Tx) error {
		bbucket := tx.Bucket([]byte(s.codingSchemaBucket))

		err := bbucket.Put(utils.Hash(chunk), utils.Encode(chunk))
		if err != nil {
			log.Panic()
		}
		return nil
	})
	if err != nil {
		log.Panic()
	}
}
