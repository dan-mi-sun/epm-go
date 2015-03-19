package consensus

import (
	"errors"
	"fmt"
	"io"

	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/account"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/binary"
	blk "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/block"
)

var (
	ErrInvalidBlockPartSignature = errors.New("Error invalid block part signature")
	ErrInvalidBlockPartHash      = errors.New("Error invalid block part hash")
)

type Proposal struct {
	Height     uint
	Round      uint
	BlockParts blk.PartSetHeader
	POLParts   blk.PartSetHeader
	Signature  account.SignatureEd25519
}

func NewProposal(height uint, round uint, blockParts, polParts blk.PartSetHeader) *Proposal {
	return &Proposal{
		Height:     height,
		Round:      round,
		BlockParts: blockParts,
		POLParts:   polParts,
	}
}

func (p *Proposal) String() string {
	return fmt.Sprintf("Proposal{%v/%v %v %v %v}", p.Height, p.Round,
		p.BlockParts, p.POLParts, p.Signature)
}

func (p *Proposal) WriteSignBytes(w io.Writer, n *int64, err *error) {
	binary.WriteUvarint(p.Height, w, n, err)
	binary.WriteUvarint(p.Round, w, n, err)
	binary.WriteBinary(p.BlockParts, w, n, err)
	binary.WriteBinary(p.POLParts, w, n, err)
}
