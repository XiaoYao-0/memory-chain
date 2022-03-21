package client

import (
	"fmt"
	"github.com/XiaoYao-0/memory-blockchain/common"
	"github.com/XiaoYao-0/memory-blockchain/core"
	"github.com/c-bata/go-prompt"
	"github.com/urfave/cli/v2"
	"strings"
)

type UserClient struct {
	BC   *core.Blockchain
	App  *cli.App
	User common.Address
}

func NewUserClient(bc *core.Blockchain) *UserClient {
	uCli := &UserClient{
		BC: bc,
	}
	uCli.App = &cli.App{
		Name: "blockchain user client",
		Commands: []*cli.Command{
			{
				Name:   "printchain",
				Usage:  "print data of blocks of the blockchain",
				Action: uCli.printChainAction(),
			},
			{
				Name:   "printtxspool",
				Usage:  "print transactions in Txs-Pool",
				Action: uCli.printTxsPoolAction(),
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
				Action: uCli.getBlockAction(),
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
				Action: uCli.getTransactionAction(),
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
				Action: uCli.sendTransactionAction(),
			},
		},
		ExitErrHandler: func(context *cli.Context, err error) {
			if err != nil {
				fmt.Println("ERROR", err)
			}
		},
	}
	return uCli
}

func userCompleter(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "printchain", Description: "Print data of blocks of the blockchain"},
		{Text: "printtxspool", Description: "Store the article text posted by user"},
		{Text: "getblock", Description: "Get a block by hash"},
		{Text: "gettransaction", Description: "Get a transaction by hash"},
		{Text: "sendtransaction", Description: "Send a transaction"},
		{Text: "help", Description: "Print help docs"},
		{Text: "exit", Description: "Exit the client"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func (uCli *UserClient) Run() {
	for {
		t := prompt.Input(">> ", userCompleter)
		switch t {
		case "help":
			_ = uCli.App.Run([]string{"-h"})
		case "exit":
			return
		default:
			_ = uCli.App.Run(append([]string{uCli.App.Name}, strings.Fields(t)...))
		}
	}
}

func (uCli *UserClient) printChainAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		fmt.Println("Print chain...")
		bci := uCli.BC.BlocksIterator()
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

func (uCli *UserClient) printTxsPoolAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		fmt.Println("Print Txs-Pool...")
		fmt.Println(uCli.BC.TxsPoolDB.TxsPool.Output())
		return nil
	}
}

func (uCli *UserClient) getBlockAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		hash, err := common.NewHash(c.String("hash"))
		if err != nil {
			return fmt.Errorf("illegal hash error: %v", err)
		}
		block, err := uCli.BC.BlocksDB.GetBlock(hash)
		if err != nil {
			return fmt.Errorf("getBlock error: %v", err)
		}
		fmt.Println(block.Output())
		return nil
	}
}

func (uCli *UserClient) getTransactionAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		hash, err := common.NewHash(c.String("hash"))
		if err != nil {
			return fmt.Errorf("illegal hash error: %v", err)
		}
		txsInPool := uCli.BC.TxsPoolDB.GetAllTxs()
		for _, tx := range txsInPool {
			if tx.Hash == hash {
				fmt.Println("This transaction is still waiting for packaged in Txs-Pool.")
				fmt.Println(tx.Output())
				return nil
			}
		}

		tx, err := uCli.BC.TransactionsDB.GetTransaction(hash)
		if err != nil {
			return fmt.Errorf("getBlock error: %v", err)
		}
		fmt.Println("This transaction has been packaged.")
		fmt.Println(tx.Output())
		return nil
	}
}

func (uCli *UserClient) sendTransactionAction() func(c *cli.Context) error {
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
		err = uCli.BC.SendTransaction(tx)
		if err != nil {
			return fmt.Errorf("sendTransaction error: %v", err)
		}
		return nil
	}
}
