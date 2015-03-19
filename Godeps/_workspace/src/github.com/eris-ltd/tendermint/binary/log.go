package binary

import (
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/logger"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/log15"
)

var log = logger.New("module", "binary")

func init() {
	log.SetHandler(
		log15.LvlFilterHandler(
			log15.LvlWarn,
			//log15.LvlDebug,
			logger.RootHandler(),
		),
	)
}
