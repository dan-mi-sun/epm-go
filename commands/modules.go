package commands

import (
	"github.com/eris-ltd/epm-go/chains"
	"github.com/eris-ltd/epm-go/epm"
	"github.com/eris-ltd/epm-go/utils"
	"os"
	"path"
	"path/filepath"

	//epm-binary-generator:IMPORT
	mod "github.com/eris-ltd/epm-go/commands/modules/thelonious"
)

// chainroot is a full path to the dir
func LoadChain(c *Context, chainType, chainRoot string) epm.Blockchain {
	rpc := c.Bool("rpc")
	logger.Debugln("Loading chain ", c.String("type"))

	chain := mod.NewChain(chainType, rpc)
	setupModule(c, chain, chainRoot)
	return chain
}

// TODO: if we are passed a chainRoot but also db is set
//   we should copy from the chainroot to db
// For now, if a chainroot is provided, we use that dir directly

func configureRootDir(c *Context, m epm.Blockchain, chainRoot string) {
	// we need to overwrite the default monk config with our defaults
	root, _ := filepath.Abs(defaultDatabase)
	m.SetProperty("RootDir", root)

	// if the HEAD is set, it overrides the default
	if typ, c, err := chains.GetHead(); err == nil && c != "" {
		root, _ = chains.ResolveChainDir(typ, c, c)
		m.SetProperty("RootDir", root)
		//path.Join(utils.Blockchains, "thelonious", c)
	}

	// if the chainRoot is set, it overwrites the head
	if chainRoot != "" {
		m.SetProperty("RootDir", chainRoot)
	}

	if c.Bool("rpc") {
		r := m.Property("RootDir").(string)
		last := filepath.Base(r)
		if last != "rpc" {
			m.SetProperty("RootDir", path.Join(r, "rpc"))
		}
	}
}

func readConfigFile(c *Context, m epm.Blockchain) {
	// if there's a config file in the root dir, use that
	// else fall back on default or flag
	// TODO: switch those priorities around!
	configFlag := c.String("config")
	s := path.Join(m.Property("RootDir").(string), "config.json")
	if _, err := os.Stat(s); err == nil {
		m.ReadConfig(s)
	} else {
		m.ReadConfig(configFlag)
	}
}

func applyFlags(c *Context, m epm.Blockchain) {
	// then apply flags
	setLogLevel(c, m)
	setKeysFile(c, m)
	setGenesisPath(c, m)
	setContractPath(c, m)
	setMining(c, m)
	setRpc(c, m)
}

func setupModule(c *Context, m epm.Blockchain, chainRoot string) {
	// TODO: kinda bullshit and useless since we set log level at epm
	// m.Config.LogLevel = defaultLogLevel

	configureRootDir(c, m, chainRoot)
	readConfigFile(c, m)
	applyFlags(c, m)
	if c.Bool("config") {
		// write the config to a temp file, open in editor, reload
		tempConfig := path.Join(utils.Epm, "tempconfig.json")
		ifExit(m.WriteConfig(tempConfig))
		ifExit(utils.Editor(tempConfig))
		ifExit(m.ReadConfig(tempConfig))
	}

	logger.Infoln("Root directory: ", m.Property("RootDir").(string))

	// initialize and start
	m.Init()
	m.Start()
}
