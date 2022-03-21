package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/XiaoYao-0/memory-blockchain/common"
	"math"
)

type Transaction struct {
	From   common.Address
	To     common.Address
	Data   []byte // less than 256 bytes
	Amount int64
	Fee    int64
	Hash   common.Hash
}

const (
	MaxRetryOfExecution         = 5      // max retry number of execution
	MaxLengthOfData             = 256    // max length of data in a Tx
	DataFeeRatio        float64 = 0.1    // data fee ratio of tx handling
	AmountFeeRatio      float64 = 0.0001 // amount fee ratio of tx handling
)

func NewTransaction(from, to common.Address, message string, amount int64) (*Transaction, error) {
	// validate the input
	if amount < 0 {
		return nil, fmt.Errorf("amount should not be less than 0")
	}
	data := []byte(message)
	if len(data) > MaxLengthOfData {
		return nil, fmt.Errorf("length of data should not be more than %v", MaxLengthOfData)
	}
	// calculate the fee of data
	dataFeeFloat64 := math.Floor(DataFeeRatio * float64(len(data)))
	if dataFeeFloat64 >= math.MaxInt64 || dataFeeFloat64 <= math.MinInt64 {
		return nil, fmt.Errorf("dataFeeFloat64=%v is out of int64 range", dataFeeFloat64)
	}
	dataFee := int64(dataFeeFloat64)
	if dataFee < 1 {
		dataFee = 1
	}
	// calculate the fee of amount
	amountFeeFloat64 := math.Floor(AmountFeeRatio * float64(amount))
	if amountFeeFloat64 >= math.MaxInt64 || amountFeeFloat64 <= math.MinInt64 {
		return nil, fmt.Errorf("amountFeeFloat64=%v is out of int64 range", amountFeeFloat64)
	}
	amountFee := int64(amountFeeFloat64)
	if amountFee < 1 {
		amountFee = 1
	}
	tx := &Transaction{
		From:   from,
		To:     to,
		Data:   data,
		Amount: amount,
		Fee:    dataFee + amountFee,
	}
	// calculate hash of tx
	txData := bytes.Join(
		[][]byte{
			tx.From.Bytes(),
			tx.To.Bytes(),
			tx.Data,
			IntToHex(tx.Amount),
			IntToHex(tx.Fee),
		},
		[]byte{},
	)
	tx.Hash = sha256.Sum256(txData)
	return tx, nil
}

// Exec transaction will roll back if failed
func (tx *Transaction) Exec(db *AccountsDB) error {
	var err error
	err = db.DecreaseBalanceOf(tx.From, tx.Fee+tx.Amount)
	if err != nil {
		return fmt.Errorf("Exec error: %v", err)
	}
	err = db.IncreaseBalanceOf(tx.To, tx.Amount)
	if err != nil {
		var err1 error
		for i := 0; i < MaxRetryOfExecution; i++ {
			err1 = db.IncreaseBalanceOf(tx.From, tx.Fee+tx.Amount)
			if err1 != nil {
				continue
			}
			break
		}
		if err1 != nil {
			panic(fmt.Errorf("failed to exec tx (%v) and roll back it: %v", tx.Hash.Hex(true), err1))
		}
		return fmt.Errorf("Exec error: %v", err)
	}
	if len(tx.Data) != 0 {
		err = db.PutMessagesTo(tx.To, tx.Data)
		if err != nil {
			var err1 error
			for i := 0; i < MaxRetryOfExecution; i++ {
				err1 = db.IncreaseBalanceOf(tx.From, tx.Fee+tx.Amount)
				if err1 != nil {
					continue
				}
				break
			}
			if err1 != nil {
				panic(fmt.Errorf("failed to exec tx (%v) and roll back it: %v", tx.Hash.Hex(true), err1))
			}
			for i := 0; i < MaxRetryOfExecution; i++ {
				err1 = db.DecreaseBalanceOf(tx.To, tx.Amount)
				if err1 != nil {
					continue
				}
				break
			}
			if err1 != nil {
				panic(fmt.Errorf("failed to exec tx (%v) and roll back it: %v", tx.Hash.Hex(true), err1))
			}
			return fmt.Errorf("Exec error: %v", err)
		}
	}
	return nil
}

// RollBack retry until the maximum number of retries is reached and crash
func (tx *Transaction) RollBack(db *AccountsDB) {
	var err error
	for i0 := 0; i0 < MaxRetryOfExecution; i0++ {
		err = db.IncreaseBalanceOf(tx.From, tx.Fee+tx.Amount)
		if err != nil {
			continue
		}
		err = db.DecreaseBalanceOf(tx.To, tx.Amount)
		if err != nil {
			continue
		}
		if len(tx.Data) != 0 {
			err = db.DeleteMessageOf(tx.To)
			if err != nil {
				continue
			}
		}
		break
	}
	if err != nil {
		panic(fmt.Errorf("tx roll back error\n%v", tx.Output()))
	}
}

func (tx *Transaction) Output() string {
	return fmt.Sprintf("Transaction %v\n"+
		"  From: %v\n"+
		"  To: %v\n"+
		"  Data: %s\n"+
		"  Amount: %v\n"+
		"  Fee: %v\n"+
		"  Hash: %v\n",
		tx.Hash.Hex(true),
		tx.From.Hex(true),
		tx.To.Hex(true),
		tx.Data,
		tx.Amount,
		tx.Fee,
		tx.Hash.Hex(true))
}

func (tx *Transaction) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	_ = encoder.Encode(tx)

	return result.Bytes()
}

func DeserializeTransaction(d []byte) (*Transaction, error) {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&transaction)
	if err != nil {
		return nil, fmt.Errorf("DeserializeTransaction error: %v", err)
	}
	return &transaction, nil
}
