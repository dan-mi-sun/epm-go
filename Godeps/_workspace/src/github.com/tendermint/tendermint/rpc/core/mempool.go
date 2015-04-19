package core

import (
	"fmt"
	. "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/common"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/merkle"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

//-----------------------------------------------------------------------------

type Receipt struct {
	TxHash          []byte
	CreatesContract bool
	ContractAddr    []byte
}

// pass pointer?
// Note: tx must be signed
func BroadcastTx(tx types.Tx) (*Receipt, error) {
	err := mempoolReactor.BroadcastTx(tx)
	if err != nil {
		return nil, fmt.Errorf("Error broadcasting transaction: %v", err)
	}

	txHash := merkle.HashFromBinary(tx)
	var createsContract bool
	var contractAddr []byte
	// check if creates new contract
	if callTx, ok := tx.(*types.CallTx); ok {
		if callTx.Address == nil {
			createsContract = true
			contractAddr = state.NewContractAddress(callTx.Input.Address, uint64(callTx.Input.Sequence))
		}
	}
	return &Receipt{txHash, createsContract, contractAddr}, nil
}

/*
curl -H 'content-type: text/plain;' http://127.0.0.1:8888/submit_tx?tx=...
*/
