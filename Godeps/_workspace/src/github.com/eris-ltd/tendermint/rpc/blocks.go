package rpc

import (
	"net/http"

	blk "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/block"
	. "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/common"
)

func BlockchainInfoHandler(w http.ResponseWriter, r *http.Request) {
	minHeight, _ := GetParamUint(r, "min_height")
	maxHeight, _ := GetParamUint(r, "max_height")
	if maxHeight == 0 {
		maxHeight = blockStore.Height()
	} else {
		maxHeight = MinUint(blockStore.Height(), maxHeight)
	}
	if minHeight == 0 {
		minHeight = uint(MaxInt(1, int(maxHeight)-20))
	}
	log.Debug("BlockchainInfoHandler", "maxHeight", maxHeight, "minHeight", minHeight)

	blockMetas := []*blk.BlockMeta{}
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta := blockStore.LoadBlockMeta(height)
		blockMetas = append(blockMetas, blockMeta)
	}

	WriteAPIResponse(w, API_OK, struct {
		LastHeight uint
		BlockMetas []*blk.BlockMeta
	}{blockStore.Height(), blockMetas})
}

//-----------------------------------------------------------------------------

func GetBlockHandler(w http.ResponseWriter, r *http.Request) {
	height, _ := GetParamUint(r, "height")
	if height == 0 {
		WriteAPIResponse(w, API_INVALID_PARAM, "height must be greater than 1")
		return
	}
	if height > blockStore.Height() {
		WriteAPIResponse(w, API_INVALID_PARAM, "height must be less than the current blockchain height")
		return
	}

	blockMeta := blockStore.LoadBlockMeta(height)
	block := blockStore.LoadBlock(height)

	WriteAPIResponse(w, API_OK, struct {
		BlockMeta *blk.BlockMeta
		Block     *blk.Block
	}{blockMeta, block})
}
