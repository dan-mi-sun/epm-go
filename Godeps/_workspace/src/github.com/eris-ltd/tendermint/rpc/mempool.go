package rpc

import (
	"net/http"

	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/binary"
	blk "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/block"
	. "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/common"
)

func BroadcastTxHandler(w http.ResponseWriter, r *http.Request) {
	txJSON := GetParam(r, "tx")
	var err error
	var tx blk.Tx
	binary.ReadJSON(&tx, []byte(txJSON), &err)
	if err != nil {
		WriteAPIResponse(w, API_INVALID_PARAM, Fmt("Invalid tx: %v", err))
		return
	}

	err = mempoolReactor.BroadcastTx(tx)
	if err != nil {
		WriteAPIResponse(w, API_ERROR, Fmt("Error broadcasting transaction: %v", err))
		return
	}

	WriteAPIResponse(w, API_OK, "")
	return
}

/*
curl -H 'content-type: text/plain;' http://127.0.0.1:8888/submit_tx?tx=...
*/
