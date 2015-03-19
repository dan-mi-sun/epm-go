package state

import (
	"bytes"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/account"
	. "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/common"
	"testing"
)

func randValidator_() *Validator {
	return &Validator{
		Address:     RandBytes(20),
		PubKey:      account.PubKeyEd25519(RandBytes(64)),
		BondHeight:  uint(RandUint32()),
		VotingPower: RandUint64(),
		Accum:       int64(RandUint64()),
	}
}

func randValidatorSet(numValidators int) *ValidatorSet {
	validators := make([]*Validator, numValidators)
	for i := 0; i < numValidators; i++ {
		validators[i] = randValidator_()
	}
	return NewValidatorSet(validators)
}

func TestCopy(t *testing.T) {
	vset := randValidatorSet(10)
	vsetHash := vset.Hash()
	if len(vsetHash) == 0 {
		t.Fatalf("ValidatorSet had unexpected zero hash")
	}

	vsetCopy := vset.Copy()
	vsetCopyHash := vsetCopy.Hash()

	if !bytes.Equal(vsetHash, vsetCopyHash) {
		t.Fatalf("ValidatorSet copy had wrong hash. Orig: %X, Copy: %X", vsetHash, vsetCopyHash)
	}
}
