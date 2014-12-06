package epm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/project-douglas/lllc-server"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var GOPATH = os.Getenv("GOPATH")

// TODO: Should be set to the "current" directory if using epm-cli
var (
	ContractPath = path.Join(GOPATH, "src", "github.com", "eris-ltd", "epm-go", "cmd", "tests", "contracts")
	TestPath     = path.Join(GOPATH, "src", "github.com", "eris-ltd", "epm-go", "cmd", "tests", "definitions")
	EPMDIR       = path.Join(usr(), ".epm-go")
)

// What to do if a job errs
const (
	PersistOnErr int = iota
	ReturnOnErr
	FailOnErr
)

// Default to fail on error
var ErrMode = FailOnErr

// Commit changes to the db (ie. mine a block)
func (e *EPM) Commit() {
	e.chain.Commit()
}

// Execute parsed jobs
func (e *EPM) ExecuteJobs() error {
	if e.Diff {
		e.checkTakeStateDiff(0)
	}
	// TODO: set gendoug...
	//gendougaddr, _:= e.eth.Get("gendoug", nil)
	//e.StoreVar("GENDOUG", gendougaddr)

	for i, j := range e.jobs {
		err := e.ExecuteJob(j)
		if e.Diff {
			e.checkTakeStateDiff(i + 1)
		}

		if err != nil {
			switch ErrMode {
			case ReturnOnErr:
				return err
			case FailOnErr:
				log.Fatal(err)
			case PersistOnErr:
				continue
			}
		}

		// time.Sleep(time.Second) // this was not necessary for epm but was when called from CI. not sure why :(
		// otherwise, tx reactors get blocked;
	}
	if e.Diff {
		e.checkTakeStateDiff(len(e.jobs))
	}
	return nil
}

// Job switch
// Args are still raw input from user (but only 2 or 3)
func (e *EPM) ExecuteJob(job Job) error {
	job.args = e.VarSub(job.args) // substitute vars
	switch job.cmd {
	case "deploy":
		return e.Deploy(job.args)
	case "modify-deploy":
		return e.ModifyDeploy(job.args)
	case "transact":
		return e.Transact(job.args)
	case "query":
		return e.Query(job.args)
	case "log":
		return e.Log(job.args)
	case "set":
		return e.Set(job.args)
	case "endow":
		return e.Endow(job.args)
	case "test":
		e.chain.Commit()
		err := e.ExecuteTest(job.args[0], 0)
		if err != nil {
			logger.Errorln(err)
			return err
		}
	case "epm":
		return e.EPMx(job.args[0])
	}
	return nil
}

// Deploy a pdx from a pdx
func (e *EPM) EPMx(filename string) error {
	// save the old jobs, empty the job list
	oldjobs := e.jobs
	e.jobs = []Job{}

	if err := e.Parse(filename); err != nil {
		logger.Errorln("failed to parse pdx file:", filename, err)
		return err
	}

	err := e.ExecuteJobs()
	if err != nil {
		return err
	}
	// return to old jobs
	e.jobs = oldjobs
	return nil
}

// Deploy a contract and save its address
func (e *EPM) Deploy(args []string) error {
	contract := args[0]
	key := args[1]
	contract = strings.Trim(contract, "\"")
	var p string
	// compile contract
	if filepath.IsAbs(contract) {
		p = contract
	} else {
		p = path.Join(ContractPath, contract)
	}
	b, err := lllcserver.Compile(p, false)
	if err != nil {
		return fmt.Errorf("Error compiling %s: %s", contract, err.Error())
	}
	// deploy contract
	addr, _ := e.chain.Script("0x"+hex.EncodeToString(b), "bytes")
	// save contract address
	e.StoreVar(key, addr)
	return nil
}

// Modify lines in the contract prior to deploy, and save its address
func (e *EPM) ModifyDeploy(args []string) error {
	contract := args[0]
	key := args[1]
	args = args[2:]

	contract = strings.Trim(contract, "\"")
	newName, err := Modify(path.Join(ContractPath, contract), args)
	if err != nil {
		return err
	}
	return e.Deploy([]string{newName, key})
}

// Send a transaction with data to a contract
func (e *EPM) Transact(args []string) error {
	to := args[0]
	dataS := args[1]
	data := strings.Split(dataS, " ")
	data = DoMath(data)
	e.chain.Msg(to, data)
	return nil
}

// Issue a query.
// TODO: Not currently functional, but not really necessary either
func (e *EPM) Query(args []string) error {
	addr := args[0]
	storage := args[1]
	varName := args[2]

	v := e.chain.StorageAt(addr, storage)
	e.StoreVar(varName, v)
	logger.Errorf("\tresult: %s = %s\n", varName, v)
	return nil
}

// Log something to the log file
func (e *EPM) Log(args []string) error {
	k := args[0]
	v := args[1]

	_, err := os.Stat(e.log)
	var f *os.File
	if err != nil {
		f, err = os.Create(e.log)
		if err != nil {
			return err
		}
	} else {
		f, err = os.OpenFile(e.log, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("%s : %s", k, v)); err != nil {
		return err
	}
	return nil
}

// Set an epm variable
func (e *EPM) Set(args []string) error {
	k := args[0]
	v := args[1]
	e.StoreVar(k, v)
	return nil
}

// Send a basic transaction transfering value.
func (e *EPM) Endow(args []string) error {
	addr := args[0]
	value := args[1]
	e.chain.Tx(addr, value)
	return nil
}

// Apply substitution: replace pairs from args to contract
// and save in a temporary file
func Modify(contract string, args []string) (string, error) {
	b, err := ioutil.ReadFile(contract)
	if err != nil {
		return "", fmt.Errorf("Could not open file %s: %s", contract, err.Error())
	}

	lll := string(b)

	// when we modify a contract, we save it in the .tmp dir in the same relative path as its original root.
	// eg. if ContractPath is ~/ponos and we modify ponos/projects/issue.lll then the modified version will be found in
	//  EPMDIR/.tmp/ponos/projects/somehash.lll
	dirC := path.Dir(contract) // absolute path
	l := len(ContractPath)
	var dir string
	if dirC != ContractPath {
		dir = dirC[l+1:] //this is relative to the contract root (ie. projects/issue.lll)
	} else {
		dir = ""
	}
	root := path.Base(ContractPath) // base of the ContractPath should be the root dir of the contracts
	dir = path.Join(root, dir)      // add in the root (ie. ponos/projects/issue.lll)

	for len(args) > 0 {
		sub := args[0]
		rep := args[1]

		lll = strings.Replace(lll, sub, rep, -1)
		args = args[2:]
	}

	hash := sha256.Sum256([]byte(lll))
	newPath := path.Join(EPMDIR, ".tmp", dir, hex.EncodeToString(hash[:])+".lll")
	err = ioutil.WriteFile(newPath, []byte(lll), 0644)
	if err != nil {
		return "", fmt.Errorf("Could not write file %s: %s", newPath, err.Error())
	}
	return newPath, nil
}
