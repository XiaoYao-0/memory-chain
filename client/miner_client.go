package client

import (
	"fmt"
	"github.com/XiaoYao-0/memory-blockchain/common"
	"github.com/XiaoYao-0/memory-blockchain/core"
	"github.com/c-bata/go-prompt"
	"github.com/urfave/cli/v2"
	"strings"
	"time"
)

type MinerClient struct {
	BC            *core.Blockchain
	App           *cli.App
	Miner         common.Address
	IsMinerSet    bool
	IsMining      bool
	MiningChannel chan struct{}
}

func NewMinerClient(bc *core.Blockchain) *MinerClient {
	mCli := &MinerClient{
		BC: bc,
	}
	mCli.App = &cli.App{
		Name: "blockchain miner client",
		Commands: []*cli.Command{
			{
				Name:  "setminer",
				Usage: "set miner",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "miner",
						Usage:    "miner",
						Required: true,
					},
				},
				Action: mCli.setMinerAction(),
			},
			{
				Name:   "getminer",
				Usage:  "get miner",
				Action: mCli.getMinerAction(),
			},
			{
				Name:   "startmining",
				Usage:  "start mining",
				Action: mCli.startMiningAction(),
			},
			{
				Name:   "endmining",
				Usage:  "end mining",
				Action: mCli.endMiningAction(),
			},
			{
				Name:   "printchain",
				Usage:  "print data of blocks of the blockchain",
				Action: mCli.printChainAction(),
			},
			{
				Name:   "printtxspool",
				Usage:  "print transactions in Txs-Pool",
				Action: mCli.printTxsPoolAction(),
			},
			{
				Name:  "getblock",
				Usage: "get a block by hash",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "hash",
						Usage:    "hash of a block (with prefix \"0x\")",
						Required: true,
					},
				},
				Action: mCli.getBlockAction(),
			},
			{
				Name:  "gettransaction",
				Usage: "get a transaction by hash",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "hash",
						Usage:    "hash of a transaction (with prefix \"0x\")",
						Required: true,
					},
				},
				Action: mCli.getTransactionAction(),
			},
			{
				Name:  "sendtransaction",
				Usage: "send a transaction",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "from",
						Usage:    "hash of a transaction (with prefix \"0x\")",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "to",
						Usage:    "hash of a transaction (with prefix \"0x\")",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "message",
						Usage:    "message you want to send",
						Required: false,
					},
					&cli.Int64Flag{
						Name:     "amount",
						Usage:    "amount you want to transfer",
						Required: false,
					},
				},
				Action: mCli.sendTransactionAction(),
			},
			{
				Name:  "getaccount",
				Usage: "get an account by address",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "addr",
						Usage:    "address of account (with prefix \"0x\")",
						Required: true,
					},
				},
				Action: mCli.getAccountAction(),
			},
		},
		ExitErrHandler: func(context *cli.Context, err error) {
			if err != nil {
				fmt.Println("ERROR", err)
			}
		},
	}
	mCli.MiningChannel = make(chan struct{})
	return mCli
}

func minerCompleter(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "setminer", Description: "Set miner"},
		{Text: "getminer", Description: "Get miner"},
		{Text: "startmining", Description: "Start mining"},
		{Text: "endmining", Description: "End mining"},
		{Text: "printchain", Description: "Print data of blocks of the blockchain"},
		{Text: "printtxspool", Description: "Print txs in Txs-Pool"},
		{Text: "getblock", Description: "Get a block by hash"},
		{Text: "gettransaction", Description: "Get a transaction by hash"},
		{Text: "sendtransaction", Description: "Send a transaction"},
		{Text: "getaccount", Description: "Get an account by address"},
		{Text: "help", Description: "Print help docs"},
		{Text: "exit", Description: "Exit the client"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func (mCli *MinerClient) Run() {
	for {
		t := prompt.Input(">> ", minerCompleter)
		switch t {
		case "help":
			_ = mCli.App.Run([]string{"-h"})
		case "exit":
			return
		default:
			_ = mCli.App.Run(append([]string{mCli.App.Name}, strings.Fields(t)...))
		}
	}
}

func (mCli *MinerClient) setMinerAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if mCli.IsMining {
			return fmt.Errorf("please stop mining first")
		}
		miner, err := common.NewAddress(c.String("miner"))
		if err != nil {
			return err
		}
		mCli.Miner = miner
		mCli.IsMinerSet = true
		fmt.Printf("Miner is set to %v\n", miner.Hex(true))
		return nil
	}
}

func (mCli *MinerClient) getMinerAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if !mCli.IsMinerSet {
			return fmt.Errorf("please set miner first")
		}
		fmt.Printf("Miner is set to %v\n", mCli.Miner.Hex(true))
		return nil
	}
}

func (mCli *MinerClient) startMiningAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if !mCli.IsMinerSet {
			return fmt.Errorf("miner is not set")
		}
		if mCli.IsMining {
			return fmt.Errorf("mining has been started")
		}
		mCli.IsMining = true
		fmt.Println("Start Mining...")
		go func(ch chan struct{}) {
			for {
				select {
				case _ = <-ch:
					return
				default:
					err := mCli.BC.MineBlock(mCli.Miner)
					if err != nil {
						time.Sleep(time.Second * 10)
					}
				}
			}
		}(mCli.MiningChannel)
		return nil
	}
}

func (mCli *MinerClient) endMiningAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if !mCli.IsMining {
			return fmt.Errorf("mining is not yet started")
		}
		fmt.Println("End Mining...")
		mCli.MiningChannel <- struct{}{}
		return nil
	}
}

func (mCli *MinerClient) printChainAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		fmt.Println("Print chain...")
		bci := mCli.BC.BlocksIterator()
		for {
			block, err := bci.Next()
			if err != nil {
				return err
			}
			fmt.Println(block.Output())
			if block.PrevBlockHash == core.NewGenesisBlock().Hash {
				fmt.Println("Genesis Block")
				break
			}
		}
		return nil
	}
}

func (mCli *MinerClient) printTxsPoolAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		fmt.Println("Print Txs-Pool...")
		fmt.Println(mCli.BC.TxsPoolDB.TxsPool.Output())
		return nil
	}
}

func (mCli *MinerClient) getBlockAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		hash, err := common.NewHash(c.String("hash"))
		if err != nil {
			return fmt.Errorf("illegal hash error: %v", err)
		}
		block, err := mCli.BC.BlocksDB.GetBlock(hash)
		if err != nil {
			return fmt.Errorf("getBlock error: %v", err)
		}
		fmt.Println(block.Output())
		return nil
	}
}

func (mCli *MinerClient) getTransactionAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		hash, err := common.NewHash(c.String("hash"))
		if err != nil {
			return fmt.Errorf("illegal hash error: %v", err)
		}
		txsInPool := mCli.BC.TxsPoolDB.GetAllTxs()
		for _, tx := range txsInPool {
			if tx.Hash == hash {
				fmt.Println("This transaction is still waiting for packaged in Txs-Pool.")
				fmt.Println(tx.Output())
				return nil
			}
		}

		tx, err := mCli.BC.TransactionsDB.GetTransaction(hash)
		if err != nil {
			return fmt.Errorf("getBlock error: %v", err)
		}
		fmt.Println("This transaction has been packaged.")
		fmt.Println(tx.Output())
		return nil
	}
}

func (mCli *MinerClient) sendTransactionAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		from, err := common.NewAddress(c.String("from"))
		if err != nil {
			return fmt.Errorf("illegal from address error: %v", err)
		}
		to, err := common.NewAddress(c.String("to"))
		if err != nil {
			return fmt.Errorf("illegal to address error: %v", err)
		}
		message := c.String("message")
		amount := c.Int64("amount")
		if message == "" && amount == 0 {
			return fmt.Errorf("amount and message cannot be both empty")
		}
		if amount < 0 {
			return fmt.Errorf("amount should be more than 0")
		}
		tx, err := core.NewTransaction(from, to, message, amount)
		if err != nil {
			return fmt.Errorf("sendTransaction error: %v", err)
		}
		err = mCli.BC.SendTransaction(tx)
		if err != nil {
			return fmt.Errorf("sendTransaction error: %v", err)
		}
		return nil
	}
}

func (mCli *MinerClient) getAccountAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		addr, err := common.NewAddress(c.String("addr"))
		if err != nil {
			return fmt.Errorf("illegal address error: %v", err)
		}
		account, err := mCli.BC.AccountsDB.GetAccountOf(addr)
		if err != nil {
			return fmt.Errorf("getAccount error: %v", err)
		}
		fmt.Println(account.Output())
		return nil
	}
}
