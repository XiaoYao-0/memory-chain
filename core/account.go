package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/XiaoYao-0/memory-blockchain/common"
	"strings"
)

type Account struct {
	Address  common.Address
	Balance  int64
	Messages [][]byte
}

func NewAccount(addr common.Address, balance int64) *Account {
	return &Account{
		Address:  addr,
		Balance:  balance,
		Messages: [][]byte{},
	}
}

func (account *Account) Output() string {
	msgs := make([]string, len(account.Messages))
	for i := 0; i < len(msgs); i++ {
		msgs[i] = string(account.Messages[i])
	}
	msgsOutput := strings.Join(msgs, "\n      ")
	return fmt.Sprintf("Account %v\n"+
		"  Address: %v\n"+
		"  Balance: %v\n"+
		"  Messages: %v\n",
		account.Address.Hex(true),
		account.Address.Hex(true),
		account.Balance,
		msgsOutput)
}

func (account *Account) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	_ = encoder.Encode(account)

	return result.Bytes()
}

func DeserializeAccount(d []byte) (*Account, error) {
	var account Account

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&account)
	if err != nil {
		return nil, fmt.Errorf("DeserializeAccount error: %v", err)
	}
	return &account, nil
}
