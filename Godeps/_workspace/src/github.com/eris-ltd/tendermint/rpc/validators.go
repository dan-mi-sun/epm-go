package rpc

import (
	"net/http"

	sm "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/state"
)

func ListValidatorsHandler(w http.ResponseWriter, r *http.Request) {
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

	WriteAPIResponse(w, API_OK, struct {
		BlockHeight         uint
		BondedValidators    []*sm.Validator
		UnbondingValidators []*sm.Validator
	}{blockHeight, bondedValidators, unbondingValidators})
}
