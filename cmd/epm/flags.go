package main

import (
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/eris-ltd/epm-go/commands"
)

var (
	nameFlag = cli.StringFlag{
		Name:   "name, n",
		Value:  "",
		Usage:  "specify a ref name",
		EnvVar: "",
	}

	forceNameFlag = cli.StringFlag{
		Name:   "force-name, N",
		Value:  "",
		Usage:  "force a ref name (even if already taken)",
		EnvVar: "",
	}

	chainFlag = cli.StringFlag{
		Name:   "chain",
		Value:  "",
		Usage:  "set the chain by <ref name> or by <type>/<id>",
		EnvVar: "",
	}

	multiFlag = cli.StringFlag{
		Name:  "multi",
		Value: "",
		Usage: "use another version of a chain with the same id",
	}

	typeFlag = cli.StringFlag{
		Name:   "type",
		Value:  "thelonious",
		Usage:  "set the chain type (thelonious, genesis, bitcoin, ethereum)",
		EnvVar: "",
	}

	interactiveFlag = cli.BoolFlag{
		Name:   "i",
		Usage:  "run epm in interactive mode",
		EnvVar: "",
	}

	diffFlag = cli.BoolFlag{
		Name:   "diff",
		Usage:  "show a diff of all contract storage",
		EnvVar: "",
	}

	dontClearFlag = cli.BoolFlag{
		Name:   "dont-clear",
		Usage:  "stop epm from clearing the epm cache on startup",
		EnvVar: "",
	}

	contractPathFlag = cli.StringFlag{
		Name:  "contracts, c",
		Value: commands.DefaultContractPath,
		Usage: "set the contract path",
	}

	pdxPathFlag = cli.StringFlag{
		Name:   "p",
		Value:  ".",
		Usage:  "specify a .pdx file to deploy",
		EnvVar: "DEPLOY_PDX",
	}

	logLevelFlag = cli.IntFlag{
		Name:   "log",
		Value:  2,
		Usage:  "set the log level",
		EnvVar: "EPM_LOG",
	}

	mineFlag = cli.BoolFlag{
		Name:  "mine, commit",
		Usage: "commit blocks",
	}

	bareFlag = cli.BoolFlag{
		Name:  "bare",
		Usage: "only copy the config",
	}

	rpcFlag = cli.BoolFlag{
		Name:   "rpc",
		Usage:  "run commands over rpc",
		EnvVar: "",
	}

	rpcHostFlag = cli.StringFlag{
		Name:  "host",
		Value: "localhost",
		Usage: "set the rpc host",
	}

	rpcPortFlag = cli.IntFlag{
		Name:  "port",
		Value: 5,
		Usage: "set the rpc port",
	}

	rpcLocalFlag = cli.BoolFlag{
		Name:  "local",
		Usage: "let the rpc server handle keys (sign txs)",
	}

	newCheckoutFlag = cli.BoolFlag{
		Name:  "checkout, o",
		Usage: "checkout the chain into head",
	}

	newConfigFlag = cli.StringFlag{
		Name:  "config, c",
		Usage: "specify config file",
	}

	newGenesisFlag = cli.StringFlag{
		Name:  "genesis, g",
		Usage: "specify genesis file",
	}

	viFlag = cli.BoolFlag{
		Name:  "vi",
		Usage: "edit the config in a vim window",
	}

	editConfigFlag = cli.BoolFlag{
		Name:  "edit-config",
		Usage: "open the config in an editor on epm new",
	}

	runConfigFlag = cli.BoolFlag{
		Name:  "config",
		Usage: "run time config edits",
	}

	noEditFlag = cli.BoolFlag{
		Name:  "no-edit",
		Usage: "prevent genesis.json from popping up (will use the default genesis.json)",
	}

	editGenesisFlag = cli.BoolFlag{
		Name:  "edit, e",
		Usage: "edit the genesis.json even if it is provided",
	}

	noImportFlag = cli.BoolFlag{
		Name:  "no-import",
		Usage: "stop epm from importing the generated key into chain's config",
	}

	noNewChainFlag = cli.BoolFlag{
		Name:  "no-new",
		Usage: "dont deploy a new chain for installation of the dapp",
	}

	compilerFlag = cli.StringFlag{
		Name:  "compiler",
		Usage: "specify <host>:<port> to use for compile server",
	}

	forceRmFlag = cli.BoolFlag{
		Name:   "force",
		Usage:  "delete directories without confirming (dangerous)",
	}
)
