package mint

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strconv"

	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/go-ethereum/logger"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/modules/types"
	//"github.com/eris-ltd/modules/types"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/account"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/config"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/consensus"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/daemon"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/tendermint/p2p"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/confer"
)

var (
	GAS      = "10000"
	GASPRICE = "500"
)

//Logging
var mintlogger *logger.Logger = logger.NewLogger("MintLogger")

// implements decerver-interfaces Blockchain
// this will get passed to Otto (javascript vm)
// as such, it does not have "administrative" methods
type MintModule struct {
	Config         *ChainConfig
	ConsensusState *consensus.ConsensusState
	App            *confer.Config
	tendermint     *daemon.Node
	started        bool

	node     *daemon.Node
	listener p2p.Listener
}

/*
   First, the functions to satisfy Module
*/

func NewMint() *MintModule {
	m := new(MintModule)
	m.Config = DefaultConfig
	m.started = false
	return m
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
	if c.UseSeed {
		app.SetDefault("SeedNode", c.RemoteHost+":"+strconv.Itoa(c.RemotePort))
	}
	app.SetDefault("GenesisFile", path.Join(c.RootDir, "genesis.json"))
	app.SetDefault("AddrBookFile", path.Join(c.RootDir, "addrbook.json"))
	app.SetDefault("PrivValidatorfile", path.Join(c.RootDir, "priv_validator.json"))
	config.SetApp(app)
}

// initialize an chain
func (mod *MintModule) Init() error {
	// config should be loaded by epm

	// transform epm json based config to tendermint config
	mod.Config2Config()

	// Create & start node
	n := daemon.NewNode()
	l := p2p.NewDefaultListener("tcp", config.App().GetString("ListenAddr"), false)
	n.AddListener(l)

	mod.listener = l
	mod.node = n
	mod.App = config.App()

	return nil
}

// start the tendermint node
func (mod *MintModule) Start() error {
	mod.node.Start()

	// If seedNode is provided by config, dial out.
	if config.App().GetString("SeedNode") != "" {
		mod.node.DialSeed()
	}

	// Run the RPC server.
	if config.App().GetString("RPC.HTTP.ListenAddr") != "" {
		mod.node.StartRpc()
	}

	mod.started = true

	if mod.Config.Mining {
		// TODO
	}
	return nil
}

func (mod *MintModule) Shutdown() error {
	mod.node.Stop()
	return nil
}

func (mod *MintModule) WaitForShutdown() {
	// Sleep forever and then...
	trapSignal(func() {
		mod.node.Stop()
	})
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

// ReadConfig and WriteConfig implemented in config.go

// What module is this?
func (mod *MintModule) Name() string {
	return "tendermint"
}

/*
 *  Implement Blockchain
 */

func (mint *MintModule) ChainId() (string, error) {
	// TODO: genhash + network
	return "TODO" + mint.App.GetString("Network"), nil
}

func (mint *MintModule) WorldState() *types.WorldState {
	stateMap := &types.WorldState{make(map[string]*types.Account), []string{}}
	var accounts []*types.Account
	state := mint.ConsensusState.GetState()
	//blockHeight = state.LastBlockHeight
	state.GetAccounts().Iterate(func(key interface{}, value interface{}) bool {
		acc := value.(*account.Account)
		hexAddr := hex.EncodeToString(acc.Address)
		stateMap.Order = append(stateMap.Order, hexAddr)
		accTy := &types.Account{
			Address: hexAddr,
			Balance: strconv.Itoa(int(acc.Balance)),
			Nonce:   strconv.Itoa(int(acc.Sequence)),
			// TODO:
			//Script:   script,
			//Storage:  storage,
			//IsScript: isscript,
		}
		accounts = append(accounts, accTy)
		return true
	})

	return stateMap
}

// tendermint/tendermint/merkel/iavl_node.go
// traverse()
func (mint *MintModule) State() *types.State {
	return nil
}

func (mint *MintModule) Storage(addr string) *types.Storage {
	return nil
}

func (mint *MintModule) Account(target string) *types.Account {
	return nil
}

func (mint *MintModule) StorageAt(contract_addr string, storage_addr string) string {
	return ""
}

func (mint *MintModule) BlockCount() int {
	return 0
}

func (mint *MintModule) LatestBlock() string {
	return ""
}

func (mint *MintModule) Block(hash string) *types.Block {
	return nil
}

func (mint *MintModule) IsScript(target string) bool {
	return true
}

// send a tx
func (mint *MintModule) Tx(addr, amt string) (string, error) {
	//keys := eth.fetchKeyPair()
	//addr = ethutil.StripHex(addr)
	if addr[:2] == "0x" {
		addr = addr[2:]
	}
	return "", nil
}

// send a message to a contract
// data is prepacked by epm
func (mint *MintModule) Msg(addr string, data []string) (string, error) {
	return "", nil
}

func (mint *MintModule) Script(script string) (string, error) {
	return "", nil
}

// returns a chanel that will fire when address is updated
func (mint *MintModule) Subscribe(name, event, target string) chan types.Event {
	return nil
}

func (mint *MintModule) UnSubscribe(name string) {
}

// Mine a block
func (m *MintModule) Commit() {
}

// start and stop continuous mining
func (m *MintModule) AutoCommit(toggle bool) {
	if toggle {
		m.StartMining()
	} else {
		m.StopMining()
	}
}

func (m *MintModule) IsAutocommit() bool {
	return false
}

/*
   Blockchain interface should also satisfy KeyManager
   All values are hex encoded
*/

// Return the active address
func (mint *MintModule) ActiveAddress() string {
	return ""
}

// Return the nth address in the ring
func (mint *MintModule) Address(n int) (string, error) {
	return "", nil
}

// Set the address
func (mint *MintModule) SetAddress(addr string) error {
	return nil
}

// Set the address to be the nth in the ring
func (mint *MintModule) SetAddressN(n int) error {
	return nil
}

// Generate a new address
func (mint *MintModule) NewAddress(set bool) string {
	return ""
}

// Return the number of available addresses
func (mint *MintModule) AddressCount() int {
	return 0
}

/*
   Helper functions
*/

func (mint *MintModule) StartMining() bool {
	return false
}

func (mint *MintModule) StopMining() bool {
	return false
}

func (mint *MintModule) StartListening() {
	//eth.ethereum.StartListening()
}

func (mint *MintModule) StopListening() {
	//eth.ethereum.StopListening()
}

/*
   some key management stuff
*/

func (mint *MintModule) fetchPriv() string {
	return ""
}

// get users home directory
func homeDir() string {
	usr, _ := user.Current()
	return usr.HomeDir
}

// convert ethereum block to types block
/*
func convertBlock(block *ethtypes.Block) *types.Block {
		if block == nil {
			return nil
		}
		b := &types.Block{}
		b.Coinbase = hex.EncodeToString(block.Coinbase())
		b.Difficulty = block.Difficulty().String()
		b.GasLimit = block.GasLimit().String()
		b.GasUsed = block.GasUsed().String()
		b.Hash = hex.EncodeToString(block.Hash())
		//b.MinGasPrice = block.MinGasPrice.String()
		b.Nonce = hex.EncodeToString(block.Nonce())
		b.Number = block.Number().String()
		b.PrevHash = hex.EncodeToString(block.ParentHash())
		b.Time = int(block.Time())
		txs := make([]*types.Transaction, len(block.Transactions()))
		for idx, tx := range block.Transactions() {
			txs[idx] = convertTx(tx)
		}
		b.Transactions = txs
		b.TxRoot = hex.EncodeToString(block.TxHash())
		b.UncleRoot = hex.EncodeToString(block.UncleHash())
		b.Uncles = make([]string, len(block.Uncles()))
		for idx, u := range block.Uncles() {
			b.Uncles[idx] = hex.EncodeToString(u.Hash())
		}

		return b
}*/

// convert ethereum tx to types tx
/*
func convertTx(ethTx *ethtypes.Transaction) *types.Transaction {
		tx := &types.Transaction{}
		tx.ContractCreation = ethtypes.IsContractAddr(ethTx.To())
		tx.Gas = ethTx.Gas().String()
		tx.GasCost = ethTx.GasPrice().String()
		tx.Hash = hex.EncodeToString(ethTx.Hash())
		tx.Nonce = fmt.Sprintf("%d", ethTx.Nonce)
		tx.Recipient = hex.EncodeToString(ethTx.To())
		tx.Sender = hex.EncodeToString(ethTx.From())
		tx.Value = ethTx.Value().String()
}
*/
