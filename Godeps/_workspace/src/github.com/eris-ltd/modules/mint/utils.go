package mint

import (
	"fmt"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/go-ethereum/logger"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/confer"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/config"
	"os"
	"os/signal"
	"path"
	"strconv"
)

//var logger = logger.NewLogger("CLI")
var interruptCallbacks = []func(os.Signal){}

// Register interrupt handlers callbacks
func RegisterInterrupt(cb func(os.Signal)) {
	interruptCallbacks = append(interruptCallbacks, cb)
}

// go routine that call interrupt handlers in order of registering
func HandleInterrupt() {
	c := make(chan os.Signal, 1)
	go func() {
		signal.Notify(c, os.Interrupt)
		for sig := range c {
			mintlogger.Errorf("Shutting down (%v) ... \n", sig)
			RunInterruptCallbacks(sig)
		}
	}()
}

func RunInterruptCallbacks(sig os.Signal) {
	for _, cb := range interruptCallbacks {
		cb(sig)
	}
}

func AbsolutePath(Datadir string, filename string) string {
	if path.IsAbs(filename) {
		return filename
	}
	return path.Join(Datadir, filename)
}

func openLogFile(Datadir string, filename string) *os.File {
	path := AbsolutePath(Datadir, filename)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening log file '%s': %v", filename, err))
	}
	return file
}

func confirm(message string) bool {
	fmt.Println(message, "Are you sure? (y/n)")
	var r string
	fmt.Scanln(&r)
	for ; ; fmt.Scanln(&r) {
		if r == "n" || r == "y" {
			break
		} else {
			fmt.Printf("Yes or no?", r)
		}
	}
	return r == "y"
}

func InitDataDir(Datadir string) {
	_, err := os.Stat(Datadir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Data directory '%s' doesn't exist, creating it\n", Datadir)
			os.Mkdir(Datadir, 0777)
		}
	}
}

func exit(err error) {
	status := 0
	if err != nil {
		fmt.Println(err)
		mintlogger.Errorln("Fatal: ", err)
		status = 1
	}
	logger.Flush()
	os.Exit(status)
}

func int2Level(i int) string {
	switch i {
	case 0:
		return "crit"
	case 1:
		return "error"
	case 2:
		return "warn"
	case 3:
		return "info"
	case 4:
		return "debug"
	case 5:
		return "debug"
	default:
		return "info"
	}
}

func (mod *MintModule) Config2Config() {
	c := mod.Config
	app := confer.NewConfig()
	app.SetDefault("Network", "tendermint_testnet0")
	app.SetDefault("ListenAddr", c.ListenHost+":"+strconv.Itoa(c.ListenPort))
	app.SetDefault("DB.Backend", "leveldb")
	app.SetDefault("DB.Dir", path.Join(c.RootDir, c.DbName))
	app.SetDefault("Log.Stdout.Level", int2Level(c.LogLevel))
	app.SetDefault("Log.File.Dir", path.Join(c.RootDir, c.DebugFile))
	app.SetDefault("Log.File.Level", "debug")
	app.SetDefault("RPC.HTTP.ListenAddr", c.RpcHost+":"+strconv.Itoa(c.RpcPort))
	app.SetDefault("Mining", c.Mining)
	if c.UseSeed {
		app.SetDefault("SeedNode", c.RemoteHost+":"+strconv.Itoa(c.RemotePort))
	}
	app.SetDefault("GenesisFile", path.Join(c.RootDir, "genesis.json"))
	app.SetDefault("AddrBookFile", path.Join(c.RootDir, "addrbook.json"))
	app.SetDefault("PrivValidatorfile", path.Join(c.RootDir, "priv_validator.json"))
	config.SetApp(app)
}

func trapSignal(cb func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			mintlogger.Infoln(fmt.Sprintf("captured %v, exiting..", sig))
			cb()
			os.Exit(1)
		}
	}()
	select {}
}
