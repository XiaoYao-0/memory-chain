package core

import (
	"fmt"
	"github.com/XiaoYao-0/memory-blockchain/common"
	"github.com/boltdb/bolt"
)

type TransactionsDB struct {
	DB *bolt.DB
}

const (
	TransactionsDBFile = "./data/transactions.db"
	TransactionsBucket = "transactions_bucket"
)

func NewTransactionsDB() (*TransactionsDB, error) {
	db, err := bolt.Open(TransactionsDBFile, 0666, nil)
	if err != nil {
		return nil, fmt.Errorf("NewTransactionsDB error: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TransactionsBucket))

		if b == nil {
			var txError error
			b, txError = tx.CreateBucket([]byte(TransactionsBucket))
			if txError != nil {
				return txError
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("NewTransactionsDB error: %v", err)
	}
	return &TransactionsDB{
		DB: db,
	}, nil
}

func (db *TransactionsDB) Close() {
	_ = db.DB.Close()
}

func (db *TransactionsDB) GetTransaction(hash common.Hash) (*Transaction, error) {
	var transaction *Transaction
	err := db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TransactionsBucket))

		if b == nil {
			return fmt.Errorf("bucket %v do not exist", TransactionsBucket)
		}
		var txError error
		encodedTransaction := b.Get(hash.Serialize())
		if encodedTransaction == nil {
			return fmt.Errorf("transaction %v do not exist", hash.Hex(true))
		}
		transaction, txError = DeserializeTransaction(encodedTransaction)
		if txError != nil {
			return txError
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("GetTransaction error: %v", err)
	}
	return transaction, nil
}

func (db *TransactionsDB) AddTransaction(transaction *Transaction) error {
	var err error
	err = db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TransactionsBucket))

		if b == nil {
			return fmt.Errorf("bucket %v do not exist", TransactionsBucket)
		}
		var txError error
		txError = b.Put(transaction.Hash.Serialize(), transaction.Serialize())
		if txError != nil {
			return txError
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("AddTransaction error: %v", err)
	}
	return nil
}

func (db *TransactionsDB) AddTransactionWithRetry(transaction *Transaction, maxRetry int) error {
	var err error
	for i := 0; i < maxRetry; i++ {
		err = db.AddTransaction(transaction)
		if err != nil {
			continue
		}
		break
	}
	return err
}
