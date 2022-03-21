package core

import (
	"fmt"
	"github.com/XiaoYao-0/memory-blockchain/common"
	"github.com/boltdb/bolt"
)

type BlocksDB struct {
	DB *bolt.DB
}

const (
	BlocksDBFile  = "./data/blocks.db"
	BlocksBucket  = "blocks_bucket"
	LastBlockHash = "last_block_hash"
)

func NewBlocksDB() (*BlocksDB, error) {
	db, err := bolt.Open(BlocksDBFile, 0666, nil)
	if err != nil {
		return nil, fmt.Errorf("NewBlocksDB error: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))

		if b == nil {
			genesis := NewGenesisBlock()
			var txError error
			b, txError = tx.CreateBucket([]byte(BlocksBucket))
			if txError != nil {
				return txError
			}
			txError = b.Put(genesis.Hash.Bytes(), genesis.Serialize())
			if txError != nil {
				return txError
			}
			txError = b.Put([]byte(LastBlockHash), genesis.Hash.Bytes())
			if txError != nil {
				return txError
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("NewBlocksDB error: %v", err)
	}
	return &BlocksDB{
		DB: db,
	}, nil
}

func (db *BlocksDB) Close() {
	_ = db.DB.Close()
}

func (db *BlocksDB) GetLastBlockHash() (common.Hash, error) {
	var hash common.Hash
	err := db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))

		if b == nil {
			return fmt.Errorf("bucket %v do not exist", BlocksBucket)
		}
		var txError error
		hash, txError = common.DeserializeHash(b.Get([]byte(LastBlockHash)))
		if txError != nil {
			return txError
		}
		return nil
	})
	if err != nil {
		return common.Hash{}, fmt.Errorf("GetLastBlockHash error: %v", err)
	}
	return hash, nil
}

func (db *BlocksDB) GetBlock(hash common.Hash) (*Block, error) {
	var block *Block
	err := db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))

		if b == nil {
			return fmt.Errorf("bucket %v do not exist", BlocksBucket)
		}
		var txError error
		encodedBlock := b.Get(hash.Serialize())
		if encodedBlock == nil {
			return fmt.Errorf("block %v do not exist", hash.Hex(true))
		}
		block, txError = DeserializeBlock(encodedBlock)
		if txError != nil {
			return txError
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("GetBlock error: %v", err)
	}
	return block, nil
}

func (db *BlocksDB) AddBlock(block *Block) error {
	var err error
	err = db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))

		if b == nil {
			return fmt.Errorf("bucket %v do not exist", BlocksBucket)
		}
		var txError error
		txError = b.Put(block.Hash.Serialize(), block.Serialize())
		if txError != nil {
			return txError
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("AddBlock error: %v", err)
	}
	return nil
}

func (db *BlocksDB) AddBlockWithRetry(block *Block, maxRetry int) error {
	var err error
	for i := 0; i < maxRetry; i++ {
		err = db.AddBlock(block)
		if err != nil {
			continue
		}
		break
	}
	return err
}
