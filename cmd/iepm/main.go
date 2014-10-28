package main

import (
    "fmt"
    "os"
    "path/filepath"
    "flag"
    "github.com/eris-ltd/epm-go"
    "github.com/eris-ltd/thelonious/monk"
    "github.com/eris-ltd/thelonious/ethchain"
    "github.com/eris-ltd/thelonious/ethreact"
)

var GoPath = os.Getenv("GOPATH")

// adjust these to suit all your deformed nefarious extension name desires. Muahahaha
var PkgExt = "pdx"
var TestExt = "pdt"

var (
    defaultContractPath = "."
    defaultPackagePath = "."
    defaultGenesis = ""
    defaultKeys = ""
    defaultDatabase = ".ethchain"
    defaultLogLevel = 0
    defaultDifficulty = 14
    defaultMining = false
    defaultDiffStorage = false

    contractPath = flag.String("c", defaultContractPath, "Set the contract root path")
    packagePath = flag.String("p", ".", "Set a .package-definition file")
    genesis = flag.String("g", "", "Set a genesis.json file")
    keys = flag.String("k", "", "Set a keys file")
    database = flag.String("db", ".ethchain", "Set the location of an eth-go root directory")
    logLevel = flag.Int("log", 0, "Set the eth log level")
    difficulty = flag.Int("dif", 14, "Set the mining difficulty")
    mining = flag.Bool("mine", true, "To mine or not to mine, that is the question")
    diffStorage = flag.Bool("diff", false, "Show a diff of all contract storage")
    clean = flag.Bool("clean", false, "Clear out epm related dirs")
    update = flag.Bool("update", false, "Pull and install the latest epm")
    install = flag.Bool("install", false, "Re-install epm")
)

func main(){
    flag.Parse() 

    var err error
    epm.ContractPath, err = filepath.Abs(*contractPath)
    if err != nil{
        fmt.Println(err)
        os.Exit(0)
    }

    // make ~/.epm-go and ~/.epm-go/.tmp for modified contract files
    epm.CheckMakeTmp()

    // Startup the EthChain
    // uses flag variables (global) for config
    eth := NewEthNode()
    // Create ChainInterface instance
    ethD := epm.NewEthD(eth)
    // setup EPM object with ChainInterface
    e := epm.NewEPM(ethD, ".epm-log")
    // subscribe to new blocks..
    e.Ch = Subscribe(eth, "newBlock")

    e.Diff = true
    e.Repl()
    
}


// subscribe on the channel
func Subscribe(eth *monk.EthChain, event string) chan ethreact.Event{
    ch := make(chan ethreact.Event, 1) 
    eth.Ethereum.Reactor().Subscribe(event, ch)
    return ch
}

// configure and start an in-process eth node
// all paths should be made absolute
func NewEthNode() *monk.EthChain{
    // empty ethchain object
    // note this will load `eth-config.json` into Config if it exists
    eth := monk.NewEth(nil)

    // we need to overwrite the default monk config with our defaults
    eth.Config.RootDir, _ = filepath.Abs(defaultDatabase)
    eth.Config.LogLevel = defaultLogLevel
    eth.Config.DougDifficulty = defaultDifficulty
    eth.Config.Mining = defaultMining
    // then try to read local config file to overwrite defaults
    // (if it doesnt exist, it will be saved)
    eth.ReadConfig("eth-config.json")
    // then apply cli flags

    // compute a map of the flags that have been set
    // if set, overwrite default/config-file
    setFlags := make(map[string]bool)
    flag.Visit(func (f *flag.Flag){
        setFlags[f.Name] = true
    })
    var err error
    if setFlags["db"]{
        eth.Config.RootDir, err = filepath.Abs(*database)
        if err != nil{
            fmt.Println(err)
            os.Exit(0)
        }
    }
    if setFlags["log"]{
        eth.Config.LogLevel = *logLevel
    }
    if setFlags["dif"]{
        eth.Config.DougDifficulty = *difficulty
    }
    if setFlags["mine"]{
        eth.Config.Mining = *mining
    }

    ethchain.GENDOUG = nil
    if *keys != defaultKeys {
        eth.Config.KeyFile, err = filepath.Abs(*keys)
        if err != nil{
            fmt.Println(err)
            os.Exit(0)
        }
    }
    if *genesis != defaultGenesis{
        eth.Config.GenesisConfig, err = filepath.Abs(*genesis)
        if err != nil{
            fmt.Println(err)
            os.Exit(0)
        }
        eth.Config.ContractPath, err = filepath.Abs(*contractPath)
        if err != nil{
            fmt.Println(err)
            os.Exit(0)
        }
    }


    // set LLL path
    epm.LLLURL = eth.Config.LLLPath

    // initialize and start
    eth.Init() 
    eth.Start()
    return eth
}
