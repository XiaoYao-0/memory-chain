package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/XiaoYao-0/memory-blockchain/common"
	"strings"
	"time"
)

const (
	MaxRetryOfAddingBlock       = 5
	MinerAwardForOneBlock int64 = 10
)

type Block struct {
	Timestamp     int64          // time when block was created
	Txs           []*Transaction // data of block
	PrevBlockHash common.Hash    // hash of previous block
	// Hash = SHA256(PrevBlockHash + Timestamp + Data)
	Hash  common.Hash // hash of block
	Nonce int64
	Miner common.Address
}

// NewBlock create a block which transactions are not packaged and proof-of-work not completed
func NewBlock(txs []*Transaction, prevBlockHash common.Hash) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Txs:           txs,
		PrevBlockHash: prevBlockHash,
		Hash:          common.Hash{},
		Nonce:         0,
	}
	return block
}

func NewGenesisBlock() *Block {
	return NewBlock([]*Transaction{}, common.Hash{})
}

// BePackaged it will be rolled back if failed
func (b *Block) BePackaged(miner common.Address, blocksDB *BlocksDB, accountsDB *AccountsDB) ([]*Transaction, error) {
	var realTxs []*Transaction
	var notPackagedTxs []*Transaction
	var err error
	for _, tx := range b.Txs {
		err = tx.Exec(accountsDB)
		if err != nil {
			notPackagedTxs = append(notPackagedTxs, tx)
			continue
		}
		realTxs = append(realTxs, tx)
	}
	if len(realTxs) == 0 {
		return []*Transaction{}, fmt.Errorf("BePackaged error: No transaction executed successfully")
	}
	oldB := *b
	b.Txs = realTxs
	pow := NewProofOfWork(b)
	nonce, hash := pow.Run()
	b.Nonce = nonce
	b.Hash = hash
	b.Miner = miner
	// award to miner
	award := MinerAwardForOneBlock
	for _, tx := range b.Txs {
		award += tx.Fee
	}
	err = accountsDB.IncreaseBalanceOf(b.Miner, award)
	if err != nil {
		for _, tx := range realTxs {
			tx.RollBack(accountsDB)
		}
		b.Txs, b.Nonce, b.Hash, b.Miner = oldB.Txs, oldB.Nonce, oldB.Hash, oldB.Miner
		return []*Transaction{}, fmt.Errorf("BePackaged error: package failed and all txs are rolled back")
	}
	// add block
	err = blocksDB.AddBlockWithRetry(b, MaxRetryOfAddingBlock)
	if err != nil {
		for _, tx := range realTxs {
			tx.RollBack(accountsDB)
		}
		b.Txs, b.Nonce, b.Hash, b.Miner = oldB.Txs, oldB.Nonce, oldB.Hash, oldB.Miner
		return []*Transaction{}, fmt.Errorf("BePackaged error: package failed and all txs are rolled back")
	}
	return notPackagedTxs, nil
}

func (b *Block) Output() string {
	txsHash := make([]string, len(b.Txs))
	for i, tx := range b.Txs {
		txsHash[i] = tx.Hash.Hex(true)
	}
	txsOutput := strings.Join(txsHash, "\n      ")
	return fmt.Sprintf("Block %v\n"+
		"  Timestamp: %v\n"+
		"  PrevBlockHash: %v\n"+
		"  Hash: %v\n"+
		"  Nonce: %v\n"+
		"  Miner: %v\n"+
		"  Txs: %v\n",
		b.Hash.Hex(true),
		time.Unix(b.Timestamp, 0).Format(time.RFC3339),
		b.PrevBlockHash.Hex(true),
		b.Hash.Hex(true),
		b.Nonce,
		b.Miner.Hex(true),
		txsOutput)
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	_ = encoder.Encode(b)

	return result.Bytes()
}

func DeserializeBlock(d []byte) (*Block, error) {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		return nil, fmt.Errorf("DeserializeBlock error: %v", err)
	}
	return &block, nil
}
