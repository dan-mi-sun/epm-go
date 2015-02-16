package epm

import (
	"fmt"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/modules/types"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/thelonious/monklog"
	"github.com/eris-ltd/epm-go/utils"
	//	"github.com/eris-ltd/lllc-server"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

var logger *monklog.Logger = monklog.NewLogger("EPM")

var (
	StateDiffOpen  = "!{"
	StateDiffClose = "!}"
)

var GOPATH = os.Getenv("GOPATH")

var (
	ContractPath = path.Join(utils.ErisLtd, "epm-go", "cmd", "tests", "contracts")
	TestPath     = path.Join(utils.ErisLtd, "epm-go", "cmd", "tests", "definitions")

	EpmDir  = utils.Epm
	LogFile = path.Join(utils.Logs, "epm", "log")
)

type KeyManager interface {
	ActiveAddress() string
	Address(n int) (string, error)
	SetAddress(addr string) error
	SetAddressN(n int) error
	NewAddress(set bool) string
	AddressCount() int
}

type DecerverModule interface {
	Init() error
	Start() error

	ReadConfig(string) error
	WriteConfig(string) error
	SetProperty(string, interface{}) error
	Property(string) interface{}
}

type Blockchain interface {
	KeyManager
	DecerverModule
	ChainId() (string, error)
	WorldState() *types.WorldState
	State() *types.State
	Storage(target string) *types.Storage
	Account(target string) *types.Account
	StorageAt(target, storage string) string
	BlockCount() int
	LatestBlock() string
	Block(hash string) *types.Block
	IsScript(target string) bool
	Tx(addr, amt string) (string, error)
	Msg(addr string, data []string) (string, error)
	Script(code string) (string, error)
	Subscribe(name, event, target string) chan types.Event
	UnSubscribe(name string)
	Commit()
	AutoCommit(toggle bool)
	IsAutocommit() bool

	Shutdown() error
	WaitForShutdown()
}

// EPM object. Maintains list of jobs and a symbols table
type EPM struct {
	chain Blockchain

	jobs []Job
	vars map[string]string

	pkgdef string
	Diff   bool
	states map[string]types.State

	//map job numbers to names of diffs invoked after that job
	diffName map[int][]string
	//map job numbers to diff actions (save or diff ie 0 or 1)
	diffSched map[int][]int

	log string
}

// New empty EPM
func NewEPM(chain Blockchain, log string) (*EPM, error) {
	//lllcserver.URL = LLLURL
	//logger.Infoln("url: ", LLLURL)
	e := &EPM{
		chain:     chain,
		jobs:      []Job{},
		vars:      make(map[string]string),
		log:       ".epm-log",
		Diff:      false, // off by default
		states:    make(map[string]types.State),
		diffName:  make(map[int][]string),
		diffSched: make(map[int][]int),
	}
	// temp dir
	err := CopyContractPath()
	return e, err
}

func (e *EPM) Stop() {
	e.chain.Shutdown()
}

// Parse a pdx file into a series of EPM jobs
func (e *EPM) Parse(filename string) error {
	logger.Infoln("Parsing ", filename)
	// set current file to parse
	e.pkgdef = filename
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	// TODO: diffs
	//diffmap := make(map[string]bool)

	p := Parse(string(b))
	if err := p.run(); err != nil {
		return err
	}
	e.jobs = p.jobs
	return nil
}

// New EPM Job
func NewJob(cmd string, args []*tree) *Job {
	j := new(Job)
	j.cmd = cmd
	j.args = [][]*tree{}
	for _, a := range args {
		j.args = append(j.args, []*tree{a})
	}
	return j
}

// Add job to EPM jobs
func (e *EPM) AddJob(j *Job) {
	e.jobs = append(e.jobs, *j)
}

func (e *EPM) VarSub(id string) (string, error) {
	if strings.HasPrefix(id, "{{") && strings.HasSuffix(id, "}}") {
		id = id[2 : len(id)-2]
	}
	v, ok := e.vars[id]
	if !ok {
		return "", fmt.Errorf("Unknown variable %s", id)
	}
	return v, nil
}

// replaces any {{varname}} args with the variable value
/*func (e *EPM) VarSub(args []string) []string {
	r, _ := regexp.Compile(`\{\{(.+?)\}\}`)
	for i, a := range args {
		// if its a known var, replace it
		// else, leave alone
		args[i] = r.ReplaceAllStringFunc(a, func(s string) string {
			k := s[2 : len(s)-2] // shave the brackets
			v, ok := e.vars[k]
			if ok {
				return v
			} else {
				return s
			}
		})
	}
	return args
}*/

// Read EPM variables in from a file
func (e *EPM) ReadVars(file string) error {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	sp := strings.Split(string(f), "\n")
	for _, kv := range sp {
		kvsp := strings.Split(kv, ":")
		if len(kvsp) != 2 {
			return fmt.Errorf("Invalid variable formatting in %s", file)
		}
		k := kvsp[0]
		v := kvsp[1]
		e.vars[k] = v
	}
	return nil
}

// Write EPM variables to file
func (e *EPM) WriteVars(file string) error {
	vars := e.Vars()
	s := ""
	for k, v := range vars {
		s += k + ":" + v + "\n"
	}
	if len(s) == 0 {
		return nil
	}
	// remove final new line
	s = s[:len(s)-1]
	err := ioutil.WriteFile(file, []byte(s), 0600)
	return err
}

// Return map of EPM variables.
func (e *EPM) Vars() map[string]string {
	return e.vars
}

// Return list of jobs
func (e *EPM) Jobs() []Job {
	return e.jobs
}

// Store a variable (strips {{ }} from key if necessary)
func (e *EPM) StoreVar(key, val string) {

	if len(key) > 4 && key[:2] == "{{" && key[len(key)-2:] == "}}" {
		key = key[2 : len(key)-2]
	}
	e.vars[key] = utils.Coerce2Hex(val)
}

func CopyContractPath() error {
	// copy the current dir into scratch/epm. Necessary for finding include files after a modify. :sigh:
	root := path.Base(ContractPath)
	p := path.Join(EpmDir, root)
	// TODO: should we delete and copy even if it does exist?
	// we might miss changed otherwise
	if _, err := os.Stat(p); err != nil {
		cmd := exec.Command("cp", "-r", ContractPath, p)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("error copying working dir into tmp: %s", err.Error())
		}
	}
	return nil
}
