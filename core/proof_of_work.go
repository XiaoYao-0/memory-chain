package core

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/XiaoYao-0/memory-blockchain/common"
	"math"
	"math/big"
	"strconv"
)

var targetBits = 24 // 挖矿难度

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int64) []byte {
	txsHash := make([][]byte, len(pow.block.Txs))
	for i, tx := range pow.block.Txs {
		txsHash[i] = tx.Hash.Bytes()
	}
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash.Bytes(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(nonce),
			pow.block.Miner.Bytes(),
		},
		bytes.Join(txsHash, []byte{}),
	)
	return data
}

func (pow *ProofOfWork) Run() (int64, common.Hash) {
	var hashInt big.Int
	var hash [32]byte
	nonce := int64(0)
	maxNonce := math.MaxInt64
	fmt.Printf("Mining block...\n")
	for nonce < int64(maxNonce) {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\r%x", hash)
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}
