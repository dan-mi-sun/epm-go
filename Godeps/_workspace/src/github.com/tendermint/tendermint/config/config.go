package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	flag "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/spf13/pflag"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/confer"
)

var app *confer.Config
var appMtx sync.Mutex

func App() *confer.Config {
	appMtx.Lock()
	defer appMtx.Unlock()
	if app == nil {
		Init("")
	}
	return app
}

func SetApp(a *confer.Config) {
	appMtx.Lock()
	defer appMtx.Unlock()
	app = a
}

// NOTE: If you change this, maybe also change initDefaults()
var defaultConfig = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

Moniker = "anonymous"
Network = "tendermint_testnet0"
ListenAddr = "0.0.0.0:8080"
# First node to connect to.  Command-line overridable.
SeedNode = "188.166.55.222:8080"

[DB]
# The only other available backend is "memdb"
Backend = "leveldb"
# Dir = "~/.tendermint/data"

[Log.Stdout]
Level = "info"

[Log.File]
Level = "debug"
# Dir = "~/.tendermint/log"

[RPC.HTTP]
# For the RPC API HTTP server.  Port required.
ListenAddr = "127.0.0.1:8081"

[Alert]
# TODO: Document options

[SMTP]
# TODO: Document options
`

var DefaultGenesis = `{
    "Accounts": [
        {
            "Address": "69988763FCF806AC35D1A2F9C4885B7DD7B0599C",
            "Amount": 2099600000000000 
        }
    ],
    "Validators": [
        {
            "PubKey": [1, "323A31EB01877858592AB7D593E9447110AFCD3ACF280D60C4F8E7C04FACC955"],
            "Amount": 100000000000,
            "UnbondTo": [
            	{
            		"Address": "69988763FCF806AC35D1A2F9C4885B7DD7B0599C",
            		"Amount":  100000000000
            	}
            ]
        },
        {
            "PubKey": [1, "DD2206E8F889EED3ABAAECEB2D18962D062A887346241820493FFE3B1DEF255D"],
            "Amount": 100000000000,
            "UnbondTo": [
            	{
            		"Address": "69988763FCF806AC35D1A2F9C4885B7DD7B0599C",
            		"Amount":  100000000000
            	}
            ]
        },
        {
            "PubKey": [1, "1B3256A3754FC6AB01110C166199A2F619E2D76DB3EE751E376FE404AC9FDCFF"],
            "Amount": 100000000000,
            "UnbondTo": [
            	{
            		"Address": "69988763FCF806AC35D1A2F9C4885B7DD7B0599C",
            		"Amount":  100000000000
            	}
            ]
        },
        {
            "PubKey": [1, "62CF1048BAEBB4FFFF360D5E896E3F4EC72D03D55183596931ED14995D512926"],
            "Amount": 100000000000,
            "UnbondTo": [
            	{
            		"Address": "69988763FCF806AC35D1A2F9C4885B7DD7B0599C",
            		"Amount":  100000000000
            	}
            ]
        }
    ]
}`

// NOTE: If you change this, maybe also change defaultConfig
func initDefaults(rootDir string) {
	app.SetDefault("Moniker", "anonymous")
	app.SetDefault("Network", "tendermint_testnet0")
	app.SetDefault("ListenAddr", "0.0.0.0:8080")
	app.SetDefault("DB.Backend", "leveldb")
	app.SetDefault("DB.Dir", rootDir+"/data")
	app.SetDefault("Log.Stdout.Level", "info")
	app.SetDefault("Log.File.Dir", rootDir+"/log")
	app.SetDefault("Log.File.Level", "debug")
	app.SetDefault("RPC.HTTP.ListenAddr", "0.0.0.0:8081")

	app.SetDefault("GenesisFile", rootDir+"/genesis.json")
	app.SetDefault("AddrBookFile", rootDir+"/addrbook.json")
	app.SetDefault("PrivValidatorfile", rootDir+"/priv_validator.json")

	app.SetDefault("FastSync", false)
}

func Init(rootDir string) {

	// Get rootdir
	if rootDir == "" {
		rootDir = os.Getenv("TMROOT")
	}
	if rootDir == "" {
		rootDir = os.Getenv("HOME") + "/.tendermint"
	}
	configFile := path.Join(rootDir, "config.toml")
	genesisFile := path.Join(rootDir, "genesis.json")

	// Write default config file if missing.
	checkWriteFile(configFile, defaultConfig)
	checkWriteFile(genesisFile, DefaultGenesis)

	// Initialize Config
	app = confer.NewConfig()
	initDefaults(rootDir)
	paths := []string{configFile}
	if err := app.ReadPaths(paths...); err != nil {
		log.Warn("Error reading configuration", "paths", paths, "error", err)
	}

	// Confused?
	// app.Debug()
}

// Check if a file exists; if not, ensure the directory is made and write the file
func checkWriteFile(configFile, contents string) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if strings.Index(configFile, "/") != -1 {
			err := os.MkdirAll(filepath.Dir(configFile), 0700)
			if err != nil {
				fmt.Printf("Could not create directory: %v", err)
				os.Exit(1)
			}
		}
		err := ioutil.WriteFile(configFile, []byte(contents), 0600)
		if err != nil {
			fmt.Printf("Could not write config file: %v", err)
			os.Exit(1)
		}
		fmt.Printf("Config file written to %v.\n", configFile)
	}
}

func ParseFlags(args []string) {
	var flags = flag.NewFlagSet("main", flag.ExitOnError)
	var printHelp = false

	// Declare flags
	flags.BoolVar(&printHelp, "help", false, "Print this help message.")
	flags.String("listen_addr", app.GetString("ListenAddr"), "Listen address. (0.0.0.0:0 means any interface, any port)")
	flags.String("seed_node", app.GetString("SeedNode"), "Address of seed node")
	flags.String("rpc_http_listen_addr", app.GetString("RPC.HTTP.ListenAddr"), "RPC listen address. Port required")
	flags.Bool("fast_sync", app.GetBool("FastSync"), "Fast blockchain syncing")
	flags.String("log_stdout_level", app.GetString("Log.Stdout.Level"), "Stdout log level")
	flags.Parse(args)
	if printHelp {
		flags.PrintDefaults()
		os.Exit(0)
	}

	// Merge parsed flag values onto app.
	app.BindPFlag("ListenAddr", flags.Lookup("listen_addr"))
	app.BindPFlag("SeedNode", flags.Lookup("seed_node"))
	app.BindPFlag("FastSync", flags.Lookup("fast_sync"))
	app.BindPFlag("RPC.HTTP.ListenAddr", flags.Lookup("rpc_http_listen_addr"))
	app.BindPFlag("Log.Stdout.Level", flags.Lookup("log_stdout_level"))

	// Confused?
	//app.Debug()
}
