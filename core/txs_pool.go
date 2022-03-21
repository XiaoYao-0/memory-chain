package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"
)

const (
	DefaultNumberOfTxsInBlock = 10
)

type TxsPool struct {
	Txs []*Transaction
}

func (p *TxsPool) getAllTxs() []*Transaction {
	return p.Txs
}

// GetSomeTxs please use DefaultNumberOfTxsInBlock
func (p *TxsPool) getSomeTxs(number int) []*Transaction {
	if len(p.Txs) <= number {
		return p.getAllTxs()
	}
	return p.Txs[:number]
}

func (p *TxsPool) addTxs(txs []*Transaction) {
	p.Txs = append(p.Txs, txs...)
}

func (p *TxsPool) leftAddTxs(txs []*Transaction) {
	p.Txs = append(txs, p.Txs...)
}

func (p *TxsPool) deleteSomeTxs(number int) {
	if len(p.Txs) <= number {
		p.Txs = []*Transaction{}
		return
	}
	p.Txs = p.Txs[number:]
}

func (p *TxsPool) Output() string {
	txsOutput := make([]string, len(p.Txs))
	for i, tx := range p.Txs {
		txsOutput[i] = tx.Output()
	}
	return fmt.Sprintf("TxsPool with %v Txs:\n%v\n", len(p.Txs), strings.Join(txsOutput, "\n"))
}

func (p *TxsPool) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	_ = encoder.Encode(p)

	return result.Bytes()
}

func DeserializeTxsPool(d []byte) (*TxsPool, error) {
	var txsPool TxsPool

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&txsPool)
	if err != nil {
		return nil, fmt.Errorf("DeserializeBlock error: %v", err)
	}
	return &txsPool, nil
}
