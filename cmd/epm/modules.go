package main

import (
	"github.com/eris-ltd/epm-go/epm"
	"github.com/eris-ltd/epm-go/chains"
	"log"
	"os"
	"path"
	"path/filepath"

    "github.com/codegangsta/cli"

	// modules
	_"github.com/eris-ltd/decerver-interfaces/glue/eth"
	_"github.com/eris-ltd/decerver-interfaces/glue/genblock"
	_"github.com/eris-ltd/decerver-interfaces/glue/monkrpc"
	"github.com/eris-ltd/thelonious/monk"
)

// chainroot is a full path to the dir
func loadChain(c *cli.Context, chainRoot string) epm.Blockchain {
    rpc := c.Bool("rpc")
	logger.Debugln("Loading chain ", c.String("type"))
	switch c.String("type"){
	case "thel", "thelonious", "monk":
		if rpc {
			//return NewMonkRpcModule(c, chainRoot)
		} else {
		    return NewMonkModule(c, chainRoot)
		}
	case "btc", "bitcoin":
		if rpc {
			log.Fatal("Bitcoin rpc not implemented yet")
		} else {
			log.Fatal("Bitcoin not implemented yet")
		}
	case "eth", "ethereum":
		if rpc {
			log.Fatal("Eth rpc not implemented yet")
		} else {
		//	return NewEthModule(c, chainRoot)
		}
	case "gen", "genesis":
		//return NewGenModule(c, chainRoot)
	}
	return nil
}

// TODO: if we are passed a chainRoot but also db is set
//   we should copy from the chainroot to db
// For now, if a chainroot is provided, we use that dir directly


func configureRootDir(c *cli.Context, m epm.Blockchain, chainRoot string){
	// we need to overwrite the default monk config with our defaults
	root, _ := filepath.Abs(defaultDatabase)
    m.SetProperty("RootDir", root)

	// if the HEAD is set, it overrides the default
	if c, err := chains.GetHead(); err == nil && c != "" {
		root, _ = chains.ResolveChain("thelonious", c, c)
        m.SetProperty("RootDir", root)
		//path.Join(utils.Blockchains, "thelonious", c)
	}

	// if the chainRoot is set, it overwrites the head
	if chainRoot != "" {
        m.SetProperty("RootDir", chainRoot)
	}
}

func readConfigFile(c *cli.Context, m epm.Blockchain){
	// if there's a config file in the root dir, use that
	// else fall back on default or flag
    configFlag := c.String("config")
	s := path.Join(m.Property("RootDir").(string), "config.json")
	if _, err := os.Stat(s); err == nil {
		m.ReadConfig(s)
	} else {
		m.ReadConfig(configFlag)
	}
}

func applyFlags(c *cli.Context, m epm.Blockchain){
	// then apply cli flags
	setLogLevel(c, m)
	setKeysFile(c, m)
	setGenesisPath(c, m)
	setContractPath(c, m)
}

func setupModule(c *cli.Context, m epm.Blockchain, chainRoot string) {
    // TODO: kinda bullshit and useless since we set log level at epm
    // m.Config.LogLevel = defaultLogLevel

    configureRootDir(c, m, chainRoot)
    readConfigFile(c, m)
    applyFlags(c, m)

	logger.Infoln("Root directory: ", m.Property("RootDir").(string))

	// initialize and start
	m.Init()
	m.Start()
}


// configure and start an in-process thelonious  node
// all paths should be made absolute
func NewMonkModule(c *cli.Context, chainRoot string) epm.Blockchain {
	m := monk.NewMonk(nil)
    setupModule(c, m, chainRoot)
    return m
}

/*

// Deploy genesis blocks using EPM
func NewGenModule(c *cli.Context, chainRoot string) epm.Blockchain {
	// empty ethchaIn object
	// note this will load `eth-config.json` into Config if it exists
	m := genblock.NewGenBlockModule(nil)

	// we need to overwrite the default monk config with our defaults
	m.Config.RootDir, _ = filepath.Abs(defaultDatabase)
	m.Config.LogLevel = defaultLogLevel

	// if the HEAD is set, it overrides the default
	if c, err := utils.GetHead(); err != nil && c != "" {
		m.Config.RootDir, _ = utils.ResolveChain("thelonious", c, "")
	}

    config := c.String("config")
    database  := c.String("db")
    logLevel := c.Int("log")
    keys :=  c.String("keys")
    contractPath := c.String("c")

	// then try to read local config file to overwrite defaults
	// (if it doesnt exist, it will be saved)
	m.ReadConfig(config)

	// then apply cli flags
	setDb(c, &(m.Config.RootDir), database)
	setLogLevel(c, &(m.Config.LogLevel), logLevel)
	setKeysFile(c, &(m.Config.KeyFile), keys)
	setContractPath(c, &(m.Config.ContractPath), contractPath)

	if chainRoot != "" {
		m.Config.RootDir = chainRoot
	}

	// set LLL path
	epm.LLLURL = m.Config.LLLPath

	// initialize and start
	m.Init()
	m.Start()
	return m
}

// Rpc module for talking to running thelonious node supporting rpc server
func NewMonkRpcModule(c *cli.Context, chainRoot string) epm.Blockchain {
	// empty ethchain object
	// note this will load `eth-config.json` into Config if it exists
	m := monkrpc.NewMonkRpcModule()

	// we need to overwrite the default monk config with our defaults
	m.Config.RootDir, _ = filepath.Abs(defaultDatabase)
	m.Config.LogLevel = defaultLogLevel

	// if the HEAD is set, it overrides the default
	if c, err := utils.GetHead(); err != nil && c != "" {
		m.Config.RootDir, _ = utils.ResolveChain("thelonious", c, "")
	}

    config := c.String("config")
    database  := c.String("db")
    logLevel := c.Int("log")
    keys :=  c.String("keys")
    contractPath := c.String("c")

	// then try to read local config file to overwrite defaults
	// (if it doesnt exist, it will be saved)
	m.ReadConfig(config)

	// then apply cli flags
	setDb(c, &(m.Config.RootDir), database)
	setLogLevel(c, &(m.Config.LogLevel), logLevel)
	setKeysFile(c, &(m.Config.KeyFile), keys)
	setContractPath(c, &(m.Config.ContractPath), contractPath)

	if chainRoot != "" {
		m.Config.RootDir = chainRoot
	}

	// set LLL path
	epm.LLLURL = m.Config.LLLPath

	// initialize and start
	m.Init()
	m.Start()
	return m
}

// configure and start an in-process eth node
func NewEthModule(c *cli.Context, chainRoot string) epm.Blockchain {
	// empty ethchain object
	m := eth.NewEth(nil)

	// we need to overwrite the default monk config with our defaults
	m.Config.RootDir, _ = filepath.Abs(defaultDatabase)
	m.Config.LogLevel = defaultLogLevel

	// if the HEAD is set, it overrides the default
	if c, err := utils.GetHead(); err != nil && c != "" {
		m.Config.RootDir, _ = utils.ResolveChain("ethereum", c, "")
	}

    config := c.String("config")
    database  := c.String("db")
    logLevel := c.Int("log")
    keys :=  c.String("keys")
    contractPath := c.String("c")

	// then try to read local config file to overwrite defaults
	// (if it doesnt exist, it will be saved)
	m.ReadConfig(config)

	// then apply cli flags
	setDb(c, &(m.Config.RootDir), database)
	setLogLevel(c, &(m.Config.LogLevel), logLevel)
	setKeysFile(c, &(m.Config.KeyFile), keys)
	setContractPath(c, &(m.Config.ContractPath), contractPath)

	if chainRoot != "" {
		m.Config.RootDir = chainRoot
	}

	// set LLL path
	epm.LLLURL = m.Config.LLLPath

	// initialize and start
	m.Init()
	m.Start()
	return m
}
*/
