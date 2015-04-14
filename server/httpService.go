package server

import (
	"bytes"
	"fmt"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/go-martini/martini"
	"github.com/eris-ltd/epm-go/commands"
	"github.com/eris-ltd/epm-go/chains"
	"github.com/eris-ltd/epm-go/epm"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"strconv"
	"time"
)

const EPM_HELP = "That API endpoint does not exist. Please see the epm documentation.\n"


type HttpService struct {
	ChainIsRunning bool
	chainIsRestarting bool
	ChainShutDownChannel chan bool
	ChainIsShutDown chan bool
	Chain epm.Blockchain
	// Maybe keep track of some statistics if this is used to create chains
	// via some Eris web service later, like it works with the compilers.
}

// Create a new http service
func NewHttpService() *HttpService {
	h := &HttpService{}
	h.ChainIsRunning = false
	h.chainIsRestarting = false
	h.ChainShutDownChannel = make(chan bool, 1)
	h.ChainIsShutDown = make(chan bool, 1)

	chainShutDownViaOS := make(chan os.Signal, 1)
	signal.Notify(chainShutDownViaOS, os.Interrupt, os.Kill)
	go func() {
	    <-chainShutDownViaOS
	    h.CleanUpAndExit()
	}()
	return h
}

// -----------------------------------------------------------------
// ------------------- INFORMATIONAL HANDLERS ----------------------
// -----------------------------------------------------------------

func (this *HttpService) handlePlop(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Plopping")

	if params["toPlop"] == "key" {
		this.writeMsg(w, 401, "Key is not exportable")
		return
	}

	cmdRaw := []string{"--chain", params["chainName"], "plop", params["toPlop"]}
	this.executeCommand(cmdRaw, w)
}

func (this *HttpService) handleLsRefs(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("List References")
	cmdRaw := []string{"refs", "ls"}
	this.executeCommand(cmdRaw, w)
}

func (this *HttpService) handleAddRefs(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Add a Reference")
	toAdd := params["chainType"] + "/" + params["chainType"]
	cmdRaw := []string{"refs", "add", toAdd, params["chainName"]}
	this.executeCommand(cmdRaw, w)
}

func (this *HttpService) handleRmRefs(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Remove a Reference")
	cmdRaw := []string{"refs", "rm", params["chainName"]}
	this.executeCommand(cmdRaw, w)
}

// -----------------------------------------------------------------
// ------------------- CHAIN MANAGEMENT HANDLERS -------------------
// -----------------------------------------------------------------

func (this *HttpService) handleConfig(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Save Config Values")

	configs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		this.logError(w, 400, err)
		return
	}

	cmdRaw := []string{"config", "--chain", params["chainName"]}
	for k, v := range configs {
		toAdd := k + ":" + v[0]
		cmdRaw = append(cmdRaw, toAdd)
	}

	this.executeCommand(cmdRaw, w)
}

func (this *HttpService) handleRawConfig(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Save Raw Config JSON String")

	// TODO: fix this to read body and and send for config saving.
	configs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		this.logError(w, 400, err)
		return
	}

	cmdRaw := []string{"config", "--chain", params["chainName"]}
	for k, v := range configs {
		toAdd := k + ":" + v[0]
		cmdRaw = append(cmdRaw, toAdd)
	}

	this.executeCommand(cmdRaw, w)
}

func (this *HttpService) handleCheckout(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Checkout a Chain")
	cmdRaw := []string{"checkout", params["chainName"]}
	this.executeCommand(cmdRaw, w)
}

func (this *HttpService) handleClean(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Removing a Chain from the Tree")
	cmdRaw := []string{"rm", "--force", params["chainName"]}
	this.executeCommand(cmdRaw, w)
}

// -----------------------------------------------------------------
// ------------------- BLOCKCHAIN ADMIN HANDLERS -------------------
// -----------------------------------------------------------------

func (this *HttpService) handleFetchChain(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Fetchin a Blockchain")

	toAdd := params["fetchIP"] + ":" + params["fetchPort"]
	chainName := params["chainName"]
	toCheckout := r.URL.Query().Get("checkout")

	cmdRaw := []string{"fetch", toAdd, "--checkout"}
	if toCheckout == "false" {
		cmdRaw = cmdRaw[:2]
	}
	if chainName != "" {
		cmdRaw = append(cmdRaw, "--force-name", chainName)
	}

	this.executeCommand(cmdRaw, w)
}

func (this *HttpService) handleNewChain(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Making a new Blockchain")

	// Read the genesis.json passed in to a temp file
	var hasGenesis bool
	genesis, err := ioutil.ReadAll(r.Body)
	if err != nil {
		this.logError(w, 400, err)
		return
	}
	defer r.Body.Close()

	var genesisFile *os.File
	defer genesisFile.Close()

	if len(genesis) != 0 {
		hasGenesis = true

		genesisFile, err = ioutil.TempFile(os.TempDir(), "epm-serve-")
		if err != nil {
			this.logError(w, 500, err)
			return
		}

		err = ioutil.WriteFile(genesisFile.Name(), genesis, 644)
		if err != nil {
			this.logError(w, 500, err)
			return
		}
	} else {
		hasGenesis = false
	}

	chainName := params["chainName"]
	chainType := r.URL.Query().Get("type")
	toCheckout := r.URL.Query().Get("checkout")
	cmdRaw := []string{"new", "--no-edit", "--checkout"}
	if toCheckout == "false" {
		cmdRaw = cmdRaw[:2]
	}
	if chainName != "" {
		cmdRaw = append(cmdRaw, "--force-name", chainName)
	}
	if chainType == "" {
		cmdRaw = append(cmdRaw, "--type", "thelonious")
	} else {
		cmdRaw = append(cmdRaw, "--type", chainType)
	}
	if hasGenesis {
		cmdRaw = append(cmdRaw, "--genesis", genesisFile.Name())
	}

	this.executeCommand(cmdRaw, w)
}

func (this *HttpService) handleStartChain(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Starting Chain Runner")

	if !this.ChainIsRunning {

		c := &commands.Context{
			Arguments: []string{},
			Strings:   make(map[string]string),
			Integers:  make(map[string]int),
			Booleans:  make(map[string]bool),
			HasSet:    make(map[string]struct{}),
		}

		root, chainType, _, err := commands.ResolveRootFlag(c)
		if err != nil {
			this.logError(w, 500, err)
			return
		}

		logLevel := r.URL.Query().Get("log")
		if logLevel == "" {
			logLevel = "2"
		}
		toCommit := r.URL.Query().Get("commit")

		go func() {
			c.Integers["log"], _ = strconv.Atoi(logLevel)
			c.Set("log")
			if toCommit == "true" {
				c.Booleans["mine"] = true
				c.Set("mine")
			}

			this.logInfo("Starting Blockchain with log level: " + logLevel)
			this.Chain = commands.LoadChain(c, chainType, root)

			<-this.ChainShutDownChannel

			this.logInfo("Shutting Down Chain")
			this.Chain.Shutdown()
			this.Chain.WaitForShutdown()
			this.ChainIsShutDown <- true
		}()

	} else {
		this.writeMsg(w, 500, "A blockchain is already running.")
		return
	}

	this.ChainIsRunning = true

	if !this.chainIsRestarting {
		this.writeMsg(w, 200, "Blockchain started.")
	}
}

func (this *HttpService) handleStopChain(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Stopping Chain Runner")

	cmdRaw := []string{"plop", "chainid"}
	toTrim, err := this.executeCommandRaw(cmdRaw, w)
	if err != nil {
		this.logError(w, 400, err)
		return
	}
	chainId := strings.TrimSpace(toTrim)

	chainType := r.URL.Query().Get("type")
	if chainType == "" {
		chainType = "thelonious"
	}

	// First check if there is a running chain via in process check.
	if this.ChainIsRunning {

		this.ChainShutDownChannel <- true
		<-this.ChainIsShutDown
		this.ChainIsRunning = false

		if !this.chainIsRestarting {
			this.writeMsg(w, 200, "Blockchain stopped.")
		}

		return
	}

	// If there is no chain running then check if there is a PID.
	chainDir := chains.ComposeRoot(chainType, chainId)
	pidFile := path.Join(chainDir, "pid")
	if _, err := os.Stat(pidFile); err != nil {
		err := fmt.Errorf("There was no blockchain running.")
		this.logError(w, 500, err)
		return
	}
	var pidInt int
	var chainProcess *os.Process
	pid, err := ioutil.ReadFile(pidFile)
	if err != nil {
		this.logError(w, 500, err)
		return
	}
	pidInt, err = strconv.Atoi(string(pid))
	if err != nil {
		this.logError(w, 500, err)
		return
	}
	chainProcess, err = os.FindProcess(pidInt)
	if err != nil {
		this.logError(w, 500, err)
		return
	}
	chainProcess.Signal(os.Interrupt)

	if !this.chainIsRestarting {
		this.writeMsg(w, 200, "Blockchain stopped.")
	}
}

func (this *HttpService) handleRestartChain(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Restarting Chain Runner")

	if this.ChainIsRunning {

		this.chainIsRestarting = true
		this.handleStopChain(params, w, r)
		time.Sleep(5 * time.Second)
		this.handleStartChain(params, w, r)
		this.chainIsRestarting = false

	} else {

		err := fmt.Errorf("There was no blockchain running.")
		this.logError(w, 500, err)
		return

	}

	this.writeMsg(w, 200, "Blockchain restarted.")
}

func (this *HttpService) handleChainStatus(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Chain Running Status")

	if this.ChainIsRunning {
		this.writeMsg(w, 200, "true")
	} else {
		this.writeMsg(w, 200, "false")
	}

}

// -----------------------------------------------------------------
// ------------------- KEYS HANDLERS -------------------------------
// -----------------------------------------------------------------

func (this *HttpService) handleKeyImport(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Keys Import")

	keyFileRaw, _, err := r.FormFile("key")
	if err != nil {
		this.logError(w, 400, err)
		return
	}

	keyFile, _ := ioutil.TempFile("", "key-")
	defer keyFile.Close()
	_, err = io.Copy(keyFile, keyFileRaw)
	if err != nil {
		this.logError(w, 500, err)
		return
	}

	newName := path.Join(os.TempDir(), params["keyName"])
	os.Rename(keyFile.Name(), newName)

	cmdRaw := []string{"keys", "import", newName}
	this.executeCommand(cmdRaw, w)
}

// -----------------------------------------------------------------
// ------------------- HELPER FUNCTIONS ----------------------------
// -----------------------------------------------------------------

// Helper function to ensure if a chain is running that it has the time to shut
//   down before the parent process exits.
func (this *HttpService) CleanUpAndExit() {
	logger.Errorln("Shutdown Signal Received")
	if this.ChainIsRunning {
		this.ChainShutDownChannel <- true
		<-this.ChainIsShutDown
	}
	os.Exit(0)
}

// Log an incoming request
func (this *HttpService) logIncoming(incoming string) {
	logger.Warnln("Incoming Handle Request: " + incoming)
}

// Log some information
func (this *HttpService) logInfo(incoming string) {
	logger.Warnln(incoming)
}

// Log that an error has happened
func (this *HttpService) logError(w http.ResponseWriter, code int, err error) {
	errString := fmt.Sprintf("%v", err)
	logger.Errorf("ERROR :(\treturning http code: %v\tbecause: %s\n", code, errString)
	this.writeMsg(w, code, errString)
}

func (this *HttpService) executeCommand(cmdRaw []string, w http.ResponseWriter) {
	product, err := this.executeCommandRaw(cmdRaw, w)
	if err != nil {
		this.logError(w, 500, err)
		return
	}
	this.writeMsg(w, 200, product)
}

func (this *HttpService) executeCommandRaw(cmdRaw []string, w http.ResponseWriter) (string, error) {
	var cmd *exec.Cmd
	cmd = exec.Command("epm")
	for _, ele := range cmdRaw {
		cmd.Args = append(cmd.Args, ele)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		this.logError(w, 500, err)
		return "", err
	}
	return out.String(), nil
}

// Handler for not found.
func (this *HttpService) handleNotFound(w http.ResponseWriter, r *http.Request) {
	this.logIncoming("404! No handler found for that endpoint.")
	this.writeMsg(w, 404, EPM_HELP)
}

// Handler for echo.
func (this *HttpService) handleEcho(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.logIncoming("Echo")
	this.writeMsg(w, 200, params["message"])
}

// Utility method for responding with an error.
func (this *HttpService) writeMsg(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, msg)
}