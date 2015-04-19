package core

import (
	ctypes "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/rpc/core/types"
	sm "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
)

//-----------------------------------------------------------------------------

func ListValidators() (*ctypes.ResponseListValidators, error) {
	var blockHeight uint
	var bondedValidators []*sm.Validator
	var unbondingValidators []*sm.Validator

	state := consensusState.GetState()
	blockHeight = state.LastBlockHeight
	state.BondedValidators.Iterate(func(index uint, val *sm.Validator) bool {
		bondedValidators = append(bondedValidators, val)
		return false
	})
	state.UnbondingValidators.Iterate(func(index uint, val *sm.Validator) bool {
		unbondingValidators = append(unbondingValidators, val)
		return false
	})

	return &ctypes.ResponseListValidators{blockHeight, bondedValidators, unbondingValidators}, nil
}
