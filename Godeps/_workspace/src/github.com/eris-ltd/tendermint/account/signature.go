package account

import (
	"fmt"

	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/binary"
	. "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/common"
)

// Signature is a part of Txs and consensus Votes.
type Signature interface {
	TypeByte() byte
}

// Types of Signature implementations
const (
	SignatureTypeNil     = byte(0x00)
	SignatureTypeEd25519 = byte(0x01)
)

// for binary.readReflect
var _ = binary.RegisterInterface(
	struct{ Signature }{},
	binary.ConcreteType{SignatureNil{}},
	binary.ConcreteType{SignatureEd25519{}},
)

//-------------------------------------

// Implements Signature
type SignatureNil struct{}

func (sig SignatureNil) TypeByte() byte { return SignatureTypeNil }

func (sig SignatureNil) IsNil() bool { return true }

func (sig SignatureNil) String() string { return "SignatureNil{}" }

//-------------------------------------

// Implements Signature
type SignatureEd25519 []byte

func (sig SignatureEd25519) TypeByte() byte { return SignatureTypeEd25519 }

func (sig SignatureEd25519) IsNil() bool { return false }

func (sig SignatureEd25519) IsZero() bool { return len(sig) == 0 }

func (sig SignatureEd25519) String() string { return fmt.Sprintf("%X", Fingerprint(sig)) }
