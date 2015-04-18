package commands

import (
	"github.com/eris-ltd/epm-go/epm"
	"log"

	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/modules/eth"
)

func NewChain(chainType string, rpc bool) epm.Blockchain {
	switch chainType {
	case "eth", "ethereum":
		if rpc {
			log.Fatal("Eth rpc not implemented yet")
		} else {
			return eth.NewEth(nil)
		}
	}
	return nil

}

func ChainSpecificDeploy(chain epm.Blockchain, deployGen, root string, novi bool) error {
	return nil
}
