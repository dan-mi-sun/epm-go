package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/eris-ltd/epm-go/crypto/secp256k1"
)

func Test0Key(t *testing.T) {
	t.Skip()
	key := common.Hex2Bytes("1111111111111111111111111111111111111111111111111111111111111111")

	p, err := secp256k1.GeneratePubKey(key)
	addr := Sha3(p[1:])[12:]
	fmt.Printf("%x\n", p)
	fmt.Printf("%v %x\n", err, addr)
}

func TestInvalidSign(t *testing.T) {
	_, err := Sign(make([]byte, 1), nil)
	if err == nil {
		t.Errorf("expected sign with hash 1 byte to error")
	}

	_, err = Sign(make([]byte, 33), nil)
	if err == nil {
		t.Errorf("expected sign with hash 33 byte to error")
	}
}
