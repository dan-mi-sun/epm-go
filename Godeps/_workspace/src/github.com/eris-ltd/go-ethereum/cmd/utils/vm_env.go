package utils

import (
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/go-ethereum/chain/types"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/go-ethereum/state"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/go-ethereum/vm"
	"math/big"
)

type VMEnv struct {
	state *state.State
	block *types.Block

	transactor []byte
	value      *big.Int
}

func NewEnv(state *state.State, block *types.Block, transactor []byte, value *big.Int) *VMEnv {
	return &VMEnv{
		state:      state,
		block:      block,
		transactor: transactor,
		value:      value,
	}
}

func (self *VMEnv) Origin() []byte        { return self.transactor }
func (self *VMEnv) BlockNumber() *big.Int { return self.block.Number }
func (self *VMEnv) PrevHash() []byte      { return self.block.PrevHash }
func (self *VMEnv) Coinbase() []byte      { return self.block.Coinbase }
func (self *VMEnv) Time() int64           { return self.block.Time }
func (self *VMEnv) Difficulty() *big.Int  { return self.block.Difficulty }
func (self *VMEnv) BlockHash() []byte     { return self.block.Hash() }
func (self *VMEnv) Value() *big.Int       { return self.value }
func (self *VMEnv) State() *state.State   { return self.state }
func (self *VMEnv) GasLimit() *big.Int    { return self.block.GasLimit }
func (self *VMEnv) AddLog(*state.Log)     {}
func (self *VMEnv) Transfer(from, to vm.Account, amount *big.Int) error {
	return vm.Transfer(from, to, amount)
}
