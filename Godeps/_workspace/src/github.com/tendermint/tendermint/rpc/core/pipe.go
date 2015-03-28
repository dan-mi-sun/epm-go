package core

import (
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/consensus"
	mempl "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/mempool"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/p2p"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

var blockStore *types.BlockStore
var consensusState *consensus.ConsensusState
var mempoolReactor *mempl.MempoolReactor
var p2pSwitch *p2p.Switch

func SetPipeBlockStore(bs *types.BlockStore) {
	blockStore = bs
}

func SetPipeConsensusState(cs *consensus.ConsensusState) {
	consensusState = cs
}

func SetPipeMempoolReactor(mr *mempl.MempoolReactor) {
	mempoolReactor = mr
}

func SetPipeSwitch(sw *p2p.Switch) {
	p2pSwitch = sw
}
