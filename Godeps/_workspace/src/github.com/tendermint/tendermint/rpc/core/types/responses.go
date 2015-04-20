package core_types

import (
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/account"
	sm "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

type ResponseGenPrivAccount struct {
	PrivAccount *account.PrivAccount
}

type ResponseGetAccount struct {
	Account *account.Account
}

type ResponseGetStorage struct {
	Key   []byte
	Value []byte
}

type ResponseCall struct {
	Return  []byte
	GasUsed uint64
	// TODO ...
}

type ResponseListAccounts struct {
	BlockHeight uint
	Accounts    []*account.Account
}

type StorageItem struct {
	Key   []byte
	Value []byte
}

type ResponseDumpStorage struct {
	StorageRoot  []byte
	StorageItems []StorageItem
}

type ResponseBlockchainInfo struct {
	LastHeight uint
	BlockMetas []*types.BlockMeta
}

type ResponseGetBlock struct {
	BlockMeta *types.BlockMeta
	Block     *types.Block
}

// curl -H 'content-type: text/plain;' http://127.0.0.1:8888/submit_tx?tx=...
type ResponseBroadcastTx struct {
	Receipt Receipt
}

type Receipt struct {
	TxHash          []byte
	CreatesContract uint8
	ContractAddr    []byte
}

type ResponseStatus struct {
	GenesisHash       []byte
	Network           string
	LatestBlockHash   []byte
	LatestBlockHeight uint
	LatestBlockTime   int64 // nano
}

type ResponseNetInfo struct {
	Moniker   string
	Network   string
	Listening bool
	Listeners []string
	Peers     []Peer
}

type Peer struct {
	Host       string // ip
	P2PPort    uint16
	RPCPort    uint16
	IsOutbound bool
}

type ResponseSignTx struct {
	Tx types.Tx
}

type ResponseListValidators struct {
	BlockHeight         uint
	BondedValidators    []*sm.Validator
	UnbondingValidators []*sm.Validator
}
