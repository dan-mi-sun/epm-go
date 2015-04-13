package server

import (
	"bytes"
	"fmt"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/go-martini/martini"
	// "github.com/eris-ltd/epm-go/commands"
	"github.com/eris-ltd/epm-go/chains"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"strconv"
	// "syscall"
	"time"
)

const EPM_HELP = "Wrong command, asshole."

type HttpService struct {
	// Maybe keep track of some statistics if this is used to create chains
	// via some Eris web service later, like it works with the compilers.
}

// Create a new http service
func NewHttpService() *HttpService {
	return &HttpService{}
}

// Handler for not found.
func (this *HttpService) handleNotFound(w http.ResponseWriter, r *http.Request) {
	this.writeMsg(w, 404, EPM_HELP)
}

// Handler for echo.
func (this *HttpService) handleEcho(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.writeMsg(w, 200, params["message"])
}

// Utility method for responding with an error.
func (this *HttpService) writeMsg(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, msg)
}

func (this *HttpService) handlePlop(params martini.Params, w http.ResponseWriter, r *http.Request) {
	if params["toPlop"] == "key" {
		this.writeMsg(w, 401, "Key is not exportable")
		return
	}
	cmd := exec.Command("epm", "--chain", params["chainName"], "plop", params["toPlop"])
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}
	this.writeMsg(w, 200, out.String())
}

func (this *HttpService) handleLsRefs(params martini.Params, w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("epm", "refs", "ls")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}
	this.writeMsg(w, 200, out.String())
}

func (this *HttpService) handleAddRefs(params martini.Params, w http.ResponseWriter, r *http.Request) {
	toAdd := params["chainType"] + "/" + params["chainType"]
	cmd := exec.Command("epm", "refs", "add", toAdd, params["chainName"])
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}
	this.writeMsg(w, 200, out.String())
}

func (this *HttpService) handleRmRefs(params martini.Params, w http.ResponseWriter, r *http.Request) {
	toAdd := params["chainName"]
	cmd := exec.Command("epm", "refs", "rm", toAdd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}
	this.writeMsg(w, 200, out.String())
}

func (this *HttpService) handleCheckout(params martini.Params, w http.ResponseWriter, r *http.Request) {
	toAdd := params["chainName"]
	cmd := exec.Command("epm", "checkout", toAdd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}
	this.writeMsg(w, 200, out.String())
}

func (this *HttpService) handleClean(params martini.Params, w http.ResponseWriter, r *http.Request) {
	toAdd := params["chainName"]
	cmd := exec.Command("epm", "rm", "--force", toAdd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}

	this.writeMsg(w, 200, out.String())
}

func (this *HttpService) handleFetch(params martini.Params, w http.ResponseWriter, r *http.Request) {
	fmt.Println("[server] Incoming: Fetching... ")
	toAdd := params["fetchIP"] + ":" + params["fetchPort"]
	chainName := params["chainName"]
	var cmd *exec.Cmd

	toCheckout := r.URL.Query().Get("checkout")
	cmd = exec.Command("epm", "fetch", toAdd, "--checkout")
	if chainName != "" {
		cmd.Args = append(cmd.Args, "--force-name", chainName)
	}
	if toCheckout == "false" {
		cmd.Args = append(cmd.Args[:3], cmd.Args[4:]...)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}

	this.writeMsg(w, 200, out.String())
}

func (this *HttpService) handleNewChain(params martini.Params, w http.ResponseWriter, r *http.Request) {
	fmt.Println("[server] Incoming: New chain... ")
	// Read the genesis.json passed in to a temp file
	var hasGenesis bool
	genesis, err := ioutil.ReadAll(r.Body)
	if err != nil {
		this.writeMsg(w, 400, "Error reading json body for genesis")
		return
	}
	defer r.Body.Close()

	var genesisFile *os.File
	defer genesisFile.Close()

	if len(genesis) != 0 {
		hasGenesis = true

		genesisFile, err = ioutil.TempFile(os.TempDir(), "epm-serve-")
		if err != nil {
			this.writeMsg(w, 500, "Internal server error. Cannot write to temp directory")
			return
		}

		err = ioutil.WriteFile(genesisFile.Name(), genesis, 644)
		if err != nil {
			this.writeMsg(w, 500, "Internal server error. Cannot write to temp directory")
			return
		}

	} else {
		hasGenesis = false
	}

	// Set up the command
	var cmd *exec.Cmd

	chainName := params["chainName"]
	chainType := r.URL.Query().Get("type")
	toCheckout := r.URL.Query().Get("checkout")
	cmd = exec.Command("epm", "new", "--no-edit", "--checkout")
	if chainName != "" {
		cmd.Args = append(cmd.Args, "--force-name", chainName)
	}
	if chainType == "" {
		cmd.Args = append(cmd.Args, "--type", "thelonious")
	} else {
		cmd.Args = append(cmd.Args, "--type", chainType)
	}
	if hasGenesis {
		cmd.Args = append(cmd.Args, "--genesis", genesisFile.Name())
	}
	if toCheckout == "false" {
		cmd.Args = append(cmd.Args[:3], cmd.Args[4:]...)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}

	this.writeMsg(w, 200, out.String())
}

func (this *HttpService) handleConfig(params martini.Params, w http.ResponseWriter, r *http.Request) {
	configs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 400, "Could not parse the config.")
		return
	}

	cmd := exec.Command("epm", "config", "--chain", params["chainName"])
	var toAdd string
	var out bytes.Buffer
	cmd.Stdout = &out
	for k, v := range configs {
		toAdd = k + ":" + v[0]
		cmd.Args = append(cmd.Args, toAdd)
	}

	err = cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}
	this.writeMsg(w, 200, out.String())
}

func (this *HttpService) handleRawConfig(params martini.Params, w http.ResponseWriter, r *http.Request) {
	configs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 400, "Could not parse the config.")
		return
	}

	cmd := exec.Command("epm", "config", "--chain", params["chainName"])
	var toAdd string
	var out bytes.Buffer
	cmd.Stdout = &out
	for k, v := range configs {
		toAdd = k + ":" + v[0]
		cmd.Args = append(cmd.Args, toAdd)
	}

	err = cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}
	this.writeMsg(w, 200, out.String())
}

func (this *HttpService) handleStartChain(params martini.Params, w http.ResponseWriter, r *http.Request) {
	fmt.Println("[server] Incoming: Starting Chain Runner")

	// this implementation will give us zombie processes but
	//   due to go's lack of a proper unix FORK there is not
	//   much that we can do about it.
	toMine := r.URL.Query().Get("mine")
	cmd := exec.Command("epm", "run")
	if toMine == "true" {
		cmd.Args = append(cmd.Args, "--mine")
	}
	err := cmd.Start()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}
	this.writeMsg(w, 200, "Blockchain restarted.")
}

func (this *HttpService) handleStopChain(params martini.Params, w http.ResponseWriter, r *http.Request) {
	fmt.Println("[server] Incoming: Stopping Chain Runner")

	var cmd *exec.Cmd
	cmd = exec.Command("epm", "plop", "chainid")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}
	chainId := strings.TrimSpace(out.String())

	chainType := r.URL.Query().Get("type")
	if chainType == "" {
		chainType = "thelonious"
	}

	chainDir := chains.ComposeRoot(chainType, chainId)
	pidFile := path.Join(chainDir, "pid")
	if _, err := os.Stat(pidFile); err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "There was no chain running. Cannot restart.")
		return
	}
	var pid []byte
	var pidInt int
	var chainProcess *os.Process
	pid, err = ioutil.ReadFile(pidFile)
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "The PID file could not be read.")
		return
	}
	pidInt, err = strconv.Atoi(string(pid))
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "The PID could not be converted into a string.")
		return
	}
	chainProcess, err = os.FindProcess(pidInt)
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not find the process.")
		return
	}
	chainProcess.Signal(os.Interrupt)
}

func (this *HttpService) handleRestartChain(params martini.Params, w http.ResponseWriter, r *http.Request) {
	fmt.Println("[server] Incoming: Restarting Chain Runner")
	this.handleStopChain(params, w, r)
	time.Sleep(5 * time.Second)
	this.handleStartChain(params, w, r)
}

func (this *HttpService) handleKeyImport(params martini.Params, w http.ResponseWriter, r *http.Request) {
	keyFileRaw, _, err := r.FormFile("key")
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 400, "Could not parse the config.")
		return
	}

	keyFile, _ := ioutil.TempFile("", "key-")
	defer keyFile.Close()
	_, err = io.Copy(keyFile, keyFileRaw)
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not save the key properly.")
		return
	}

	newName := path.Join(os.TempDir(), params["keyName"])
	os.Rename(keyFile.Name(), newName)

	cmd := exec.Command("epm", "keys", "import", newName)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		fmt.Printf("[server] ERROR: %q\n", err)
		this.writeMsg(w, 500, "Could not execute command.")
		return
	}
	this.writeMsg(w, 200, out.String())
}