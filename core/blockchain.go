package core

import (
	"fmt"
	"github.com/XiaoYao-0/memory-blockchain/common"
)

type Blockchain struct {
	Tip            common.Hash // the hash of the last block in a chain
	BlocksDB       *BlocksDB
	AccountsDB     *AccountsDB
	TransactionsDB *TransactionsDB
	TxsPoolDB      *TxsPoolDB
}

func NewBlockchain() (*Blockchain, error) {
	blocksDB, err := NewBlocksDB()
	if err != nil {
		return nil, fmt.Errorf("NewBlockchain error: %v", err)
	}
	accountsDB, err := NewAccountsDB()
	if err != nil {
		return nil, fmt.Errorf("NewBlockchain error: %v", err)
	}
	transactionsDB, err := NewTransactionsDB()
	if err != nil {
		return nil, fmt.Errorf("NewBlockchain error: %v", err)
	}
	txsPool, err := NewTxsPoolDB()
	if err != nil {
		return nil, fmt.Errorf("NewBlockchain error: %v", err)
	}
	tip, err := blocksDB.GetLastBlockHash()
	if err != nil {
		return nil, fmt.Errorf("NewBlockchain error: %v", err)
	}
	bc := Blockchain{
		Tip:            tip,
		BlocksDB:       blocksDB,
		AccountsDB:     accountsDB,
		TransactionsDB: transactionsDB,
		TxsPoolDB:      txsPool,
	}
	return &bc, nil
}

func (bc *Blockchain) CloseDB() {
	bc.BlocksDB.Close()
	bc.AccountsDB.Close()
}

func (bc *Blockchain) SendTransaction(tx *Transaction) error {
	// Check if there is enough balance in the account to pay the handling fee and transfer amount
	balance, err := bc.AccountsDB.GetBalanceOf(tx.From)
	if err != nil {
		return fmt.Errorf("SendTransaction error: "+
			"fail to check if there is enough balance in the account to pay the handling fee and transfer amount: %v", err)
	}
	if balance < tx.Amount+tx.Fee {
		return fmt.Errorf("SendTransaction error: "+
			"your balance (%v) is not enough to cover the handling fee (%v) and amount (%v) you want to transfer",
			balance, tx.Fee, tx.Amount)
	}
	bc.TxsPoolDB.AddTxs([]*Transaction{tx})
	fmt.Printf("💰 Transaction send!\n")
	fmt.Println(tx.Output())
	fmt.Println("Transaction is waiting for packaged...")
	return nil
}

func (bc *Blockchain) MineBlock(miner common.Address) error {
	txs := bc.TxsPoolDB.GetSomeTxs(DefaultNumberOfTxsInBlock)
	if len(txs) == 0 {
		fmt.Println("❌ There is no tx in pool")
		return fmt.Errorf("there is no tx in pool")
	}
	realTxsCount := len(txs)
	block := NewBlock(txs, bc.Tip)
	notPackagedTxs, err := block.BePackaged(miner, bc.BlocksDB, bc.AccountsDB)
	if err != nil {
		fmt.Println("❌ Failed to mine new block")
		return fmt.Errorf("MineBlock error: %v", err)
	}
	bc.TxsPoolDB.DeleteSomeTxs(realTxsCount)
	bc.TxsPoolDB.LeftAddTxs(notPackagedTxs)
	fmt.Printf("🔨 New Block Mined!\n")
	fmt.Println(block.Output())
	return nil
}

type BlocksIterator struct {
	currentHash common.Hash
	db          *BlocksDB
}

func (bc *Blockchain) BlocksIterator() *BlocksIterator {
	bci := &BlocksIterator{bc.Tip, bc.BlocksDB}
	return bci
}

func (i *BlocksIterator) Next() (*Block, error) {
	block, err := i.db.GetBlock(i.currentHash)
	if err != nil {
		return nil, fmt.Errorf("Next error: %v", err)
	}
	i.currentHash = block.PrevBlockHash
	return block, nil
}
