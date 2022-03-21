package common

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

type Address [20]byte

// NewAddress with prefix "0x"
func NewAddress(hexS string) (Address, error) {
	if !strings.HasPrefix(hexS, "0x") {
		return Address{}, errors.New("address hex should start with '0x'")
	}
	if len(hexS) != 42 {
		return Address{}, errors.New("address hex should be 40-bit hex number")
	}
	bytes, err := hex.DecodeString(hexS[2:])
	if err != nil || len(bytes) != 20 {
		return Address{}, errors.New("not a hex number")
	}
	addr := [20]byte{}
	for i := 0; i < 20; i++ {
		addr[i] = bytes[i]
	}
	return addr, nil
}

func ZeroAddress() Address {
	return [20]byte{}
}

func (addr Address) Hex(withPrefix bool) string {
	if withPrefix {
		return fmt.Sprintf("0x%x", addr)
	}
	return fmt.Sprintf("%x", addr)
}

func (addr Address) Bytes20() [20]byte {
	return addr
}

func (addr Address) Bytes() []byte {
	bytes := make([]byte, 20)
	bytes20 := addr.Bytes20()
	for i := 0; i < 20; i++ {
		bytes[i] = bytes20[i]
	}
	return bytes
}

func (addr Address) Serialize() []byte {
	return addr.Bytes()
}

func DeserializeAddress(d []byte) (Address, error) {
	var addr Address
	if len(d) != 20 {
		return Address{}, fmt.Errorf("DeserializeAddress error: data should be 20 bytes instead of %v bytes", len(d))
	}
	return addr, nil
}
