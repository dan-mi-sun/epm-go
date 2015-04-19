package commands

import (
	"fmt"
	"log"

	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/modules/mint"
	"github.com/eris-ltd/epm-go/epm"
)

func NewChain(chainType string, rpc bool) epm.Blockchain {
	switch chainType {
	case "tendermint", "mint":
		if rpc {
			log.Fatal("Tendermint rpc not implemented yet")
		} else {
			return mint.NewMint()
		}
	}
	return nil

}

func ChainSpecificDeploy(chain epm.Blockchain, deployGen, root string, novi bool) error {
	return nil
}

func Fetch(chainType, peerserver string) ([]byte, error) {
	return nil, fmt.Errorf("Fetch not supported for mint")
}
