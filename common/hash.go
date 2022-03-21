package common

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

type Hash [32]byte

// NewHash with prefix "0x"
func NewHash(hexS string) (Hash, error) {
	if !strings.HasPrefix(hexS, "0x") {
		return Hash{}, errors.New("hash hex should start with '0x'")
	}
	if len(hexS) != 66 {
		return Hash{}, errors.New("hash hex should be 64-bit hex number")
	}
	bytes, err := hex.DecodeString(hexS[2:])
	if err != nil || len(bytes) != 32 {
		return Hash{}, errors.New("not a hex number")
	}
	hash := [32]byte{}
	for i := 0; i < 32; i++ {
		hash[i] = bytes[i]
	}
	return hash, nil
}

func (hash Hash) Hex(withPrefix bool) string {
	if withPrefix {
		return fmt.Sprintf("0x%x", hash)
	}
	return fmt.Sprintf("%x", hash)
}

func (hash Hash) Bytes32() [32]byte {
	return hash
}

func (hash Hash) Bytes() []byte {
	bytes := make([]byte, 32)
	bytes32 := hash.Bytes32()
	for i := 0; i < 32; i++ {
		bytes[i] = bytes32[i]
	}
	return bytes
}

func (hash Hash) Serialize() []byte {
	return hash.Bytes()
}

func DeserializeHash(d []byte) (Hash, error) {
	var hash Hash
	if len(d) != 32 {
		return Hash{}, fmt.Errorf("DeserializeHash error: data should be 32 bytes instead of %v bytes", len(d))
	}
	return hash, nil
}
