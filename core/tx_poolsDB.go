package core

import (
	"fmt"
	"github.com/boltdb/bolt"
)

const (
	TxsPoolDBFile      = "./data/txs_pool.db"
	TxsPoolBucket      = "blocks_bucket"
	TxsPoolKey         = "txs_pool"
	MaxRetryOfFlushing = 5
)

type TxsPoolDB struct {
	TxsPool *TxsPool
	DB      *bolt.DB
}

func NewTxsPoolDB() (*TxsPoolDB, error) {
	db, err := bolt.Open(TxsPoolDBFile, 0666, nil)
	if err != nil {
		return nil, fmt.Errorf("NewTxsPoolDB error: %v", err)
	}
	txsPool := &TxsPool{Txs: []*Transaction{}}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TxsPoolBucket))
		var txError error
		if b == nil {
			b, txError = tx.CreateBucket([]byte(TxsPoolBucket))
			if txError != nil {
				return txError
			}
			txError = b.Put([]byte(TxsPoolKey), txsPool.Serialize())
			if txError != nil {
				return txError
			}
			return nil
		}
		encodedTxsPool := b.Get([]byte(TxsPoolKey))
		txsPool, txError = DeserializeTxsPool(encodedTxsPool)
		if txError != nil {
			return txError
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("NewTxsPoolDB error: %v", err)
	}
	return &TxsPoolDB{
		TxsPool: txsPool,
		DB:      db,
	}, nil
}

func (db *TxsPoolDB) GetAllTxs() []*Transaction {
	return db.TxsPool.getAllTxs()
}

// GetSomeTxs please use DefaultNumberOfTxsInBlock
func (db *TxsPoolDB) GetSomeTxs(number int) []*Transaction {
	return db.TxsPool.getSomeTxs(number)
}

func (db *TxsPoolDB) AddTxs(txs []*Transaction) {
	db.TxsPool.addTxs(txs)
	db.flush()
}

func (db *TxsPoolDB) LeftAddTxs(txs []*Transaction) {
	db.TxsPool.leftAddTxs(txs)
	db.flush()
}

func (db *TxsPoolDB) DeleteSomeTxs(number int) {
	db.TxsPool.deleteSomeTxs(number)
	db.flush()
}

func (db *TxsPoolDB) flush() {
	go func() {
		var err error
		for i := 0; i < MaxRetryOfFlushing; i++ {
			err = db.DB.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(TxsPoolBucket))
				var txError error
				if b == nil {
					return fmt.Errorf("bucket %v do not exist", TxsPoolBucket)
				}
				txError = b.Put([]byte(TxsPoolKey), db.TxsPool.Serialize())
				if txError != nil {
					return txError
				}
				return nil
			})
			if err != nil {
				continue
			}
			return
		}
		if err != nil {
			panic(fmt.Errorf("flush txs_pool db error: %v", err))
		}
	}()
}
