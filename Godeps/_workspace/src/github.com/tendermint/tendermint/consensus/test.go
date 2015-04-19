package consensus

import (
	"sort"

	bc "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/blockchain"
	dbm "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/db"
	mempl "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/mempool"
	sm "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
)

func randConsensusState() (*ConsensusState, []*sm.PrivValidator) {
	state, _, privValidators := sm.RandGenesisState(20, false, 1000, 10, false, 1000)
	blockStore := bc.NewBlockStore(dbm.NewMemDB())
	mempool := mempl.NewMempool(state)
	mempoolReactor := mempl.NewMempoolReactor(mempool)
	cs := NewConsensusState(state, blockStore, mempoolReactor)
	return cs, privValidators
}

func randVoteSet(height uint, round uint, type_ byte, numValidators int, votingPower uint64) (*VoteSet, *sm.ValidatorSet, []*sm.PrivValidator) {
	vals := make([]*sm.Validator, numValidators)
	privValidators := make([]*sm.PrivValidator, numValidators)
	for i := 0; i < numValidators; i++ {
		_, val, privValidator := sm.RandValidator(false, votingPower)
		vals[i] = val
		privValidators[i] = privValidator
	}
	valSet := sm.NewValidatorSet(vals)
	sort.Sort(sm.PrivValidatorsByAddress(privValidators))
	return NewVoteSet(height, round, type_, valSet), valSet, privValidators
}
