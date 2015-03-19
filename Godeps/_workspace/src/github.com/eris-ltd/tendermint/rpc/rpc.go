package rpc

import (
	blk "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/block"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/consensus"
	mempl "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/mempool"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/p2p"
)

var blockStore *blk.BlockStore
var consensusState *consensus.ConsensusState
var mempoolReactor *mempl.MempoolReactor
var p2pSwitch *p2p.Switch

func SetRPCBlockStore(bs *blk.BlockStore) {
	blockStore = bs
}

func SetRPCConsensusState(cs *consensus.ConsensusState) {
	consensusState = cs
}

func SetRPCMempoolReactor(mr *mempl.MempoolReactor) {
	mempoolReactor = mr
}

func SetRPCSwitch(sw *p2p.Switch) {
	p2pSwitch = sw
}
