package core

import (
	"fmt"
	"github.com/XiaoYao-0/memory-blockchain/common"
	"github.com/boltdb/bolt"
	"math"
)

type AccountsDB struct {
	DB *bolt.DB
}

const (
	AccountsFile   = "./data/accounts.db"
	AccountsBucket = "accounts_bucket"
)

func NewAccountsDB() (*AccountsDB, error) {
	db, err := bolt.Open(AccountsFile, 0666, nil)
	if err != nil {
		return nil, fmt.Errorf("NewAccountsDB error: %v", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(AccountsBucket))

		if b == nil {
			var txError error
			b, txError = tx.CreateBucket([]byte(AccountsBucket))
			if txError != nil {
				return txError
			}
			for _, account := range initAccounts() {
				txError = b.Put(account.Address.Serialize(), account.Serialize())
				if txError != nil {
					return txError
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("NewAccountsDB error: %v", err)
	}
	return &AccountsDB{
		DB: db,
	}, nil
}

func (db *AccountsDB) Close() {
	_ = db.DB.Close()
}

var InitialAccounts map[common.Address]int64

// InitAccounts allocate initial funds
func initAccounts() []*Account {
	var accounts []*Account
	InitialAccounts = make(map[common.Address]int64)
	addr1, _ := common.NewAddress("0x0000000000000000000000000000000000000001")
	addr2, _ := common.NewAddress("0x0000000000000000000000000000000000000002")
	addr3, _ := common.NewAddress("0x0000000000000000000000000000000000000003")
	addr4, _ := common.NewAddress("0x0000000000000000000000000000000000000004")
	addr5, _ := common.NewAddress("0x0000000000000000000000000000000000000005")
	InitialAccounts[addr1] = int64(math.Pow10(10))
	InitialAccounts[addr2] = int64(math.Pow10(10))
	InitialAccounts[addr3] = int64(math.Pow10(10))
	InitialAccounts[addr4] = int64(math.Pow10(10))
	InitialAccounts[addr5] = int64(math.Pow10(10))
	for addr, balance := range InitialAccounts {
		if balance < 0 {
			balance = 0
		}
		accounts = append(accounts, NewAccount(addr, balance))
	}
	return accounts
}

func (db *AccountsDB) GetAccountOf(addr common.Address) (*Account, error) {
	var err error
	var account *Account
	account = &Account{
		Address:  addr,
		Balance:  0,
		Messages: [][]byte{},
	}
	err = db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(AccountsBucket))

		if b == nil {
			return fmt.Errorf("bucket %v do not exist", AccountsBucket)
		}
		var txError error
		encodedAccount := b.Get(addr.Bytes())
		if encodedAccount == nil {
			txError = b.Put(addr.Serialize(), account.Serialize())
			if txError != nil {
				return txError
			}
		} else {
			account, txError = DeserializeAccount(encodedAccount)
			if txError != nil {
				return txError
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("GetAccountOf error: %v", err)
	}
	return account, nil
}

func (db *AccountsDB) GetBalanceOf(addr common.Address) (int64, error) {
	account, err := db.GetAccountOf(addr)
	if err != nil {
		return 0, fmt.Errorf("GetBalanceOf error: %v", err)
	}
	return account.Balance, nil
}

func (db *AccountsDB) IncreaseBalanceOf(addr common.Address, amount int64) error {
	if amount < 0 {
		return fmt.Errorf("IncreaseBalanceOf error: amount=%v<0", amount)
	}
	account, err := db.GetAccountOf(addr)
	if err != nil {
		return fmt.Errorf("IncreaseBalanceOf error: %v", err)
	}
	if account.Balance+amount < account.Balance {
		return fmt.Errorf("IncreaseBalanceOf error: integer overflow: %v+%v->%v", account.Balance, amount, account.Balance+amount)
	}
	account.Balance += amount
	err = db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(AccountsBucket))

		if b == nil {
			return fmt.Errorf("bucket %v do not exist", AccountsBucket)
		}
		var txError error
		txError = b.Put(addr.Serialize(), account.Serialize())
		if txError != nil {
			return txError
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("IncreaseBalanceOf error: %v", err)
	}
	return nil
}

func (db *AccountsDB) DecreaseBalanceOf(addr common.Address, amount int64) error {
	if amount < 0 {
		return fmt.Errorf("DecreaseBalanceOf error: amount=%v<0", amount)
	}
	account, err := db.GetAccountOf(addr)
	if err != nil {
		return fmt.Errorf("DecreaseBalanceOf error: %v", err)
	}
	if account.Balance-amount > account.Balance {
		return fmt.Errorf("DecreaseBalanceOf error: integer underflow: %v-%v->%v", account.Balance, amount, account.Balance+amount)
	}
	account.Balance -= amount
	err = db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(AccountsBucket))
		if b == nil {
			return fmt.Errorf("bucket %v do not exist", AccountsBucket)
		}
		var txError error
		txError = b.Put(addr.Serialize(), account.Serialize())
		if txError != nil {
			return txError
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("DecreaseBalanceOf error: %v", err)
	}
	return nil
}

func (db *AccountsDB) GetMessagesOf(addr common.Address) ([][]byte, error) {
	account, err := db.GetAccountOf(addr)
	if err != nil {
		return [][]byte{}, fmt.Errorf("GetMessagesOf error: %v", err)
	}
	return account.Messages, nil
}

func (db *AccountsDB) PutMessagesTo(addr common.Address, message []byte) error {
	account, err := db.GetAccountOf(addr)
	if err != nil {
		return fmt.Errorf("PutMessagesTo error: %v", err)
	}
	account.Messages = append(account.Messages, message)
	err = db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(AccountsBucket))

		if b == nil {
			return fmt.Errorf("bucket %v do not exist", AccountsBucket)
		}
		var txError error
		txError = b.Put(addr.Serialize(), account.Serialize())
		if txError != nil {
			return txError
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("PutMessagesTo error: %v", err)
	}
	return nil
}

func (db *AccountsDB) DeleteMessageOf(addr common.Address) error {
	account, err := db.GetAccountOf(addr)
	if err != nil {
		return fmt.Errorf("DeleteMessageOf error: %v", err)
	}
	if len(account.Messages) == 0 {
		return fmt.Errorf("DeleteMessageOf error: account has no message")
	}
	account.Messages = account.Messages[:len(account.Messages)-1]
	err = db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(AccountsBucket))

		if b == nil {
			return fmt.Errorf("bucket %v do not exist", AccountsBucket)
		}
		var txError error
		txError = b.Put(addr.Serialize(), account.Serialize())
		if txError != nil {
			return txError
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("DeleteMessageOf error: %v", err)
	}
	return nil
}
