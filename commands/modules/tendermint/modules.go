package commands

import (
	"github.com/eris-ltd/epm-go/epm"
	"log"

	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/modules/mint"
)

func NewChain(chainType string, rpc bool) epm.Blockchain {
	switch chainType {
	case "tendermint", "mint":
		if rpc {
			log.Fatal("Tendermint rpc not implemented yet")
		} else {
			return mint.NewMintModule(nil)
		}
	}
	return nil

}

func ChainSpecificDeploy(chain epm.Blockchain, deployGen, root string, novi bool) error {
	return nil
}
