package main

import (
	"github.com/XiaoYao-0/memory-blockchain/client"
	"github.com/XiaoYao-0/memory-blockchain/core"
	"log"
)

func main() {
	bc, err := core.NewBlockchain()
	if err != nil {
		log.Fatal(err)
	}
	defer bc.CloseDB()

	cli := client.NewMinerClient(bc)
	cli.Run()
}
