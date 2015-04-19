package rpc

import (
	"bytes"
	"encoding/hex"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/account"
	. "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/common"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/config"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/consensus"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/logger"
	nm "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/node"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/p2p"
	ctypes "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/rpc/core/types"
	cclient "github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/rpc/core_client"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
	"testing"
	"time"
)

// global variables for use across all tests
var (
	rpcAddr       = "127.0.0.1:8089"
	requestAddr   = "http://" + rpcAddr + "/"
	websocketAddr = "ws://" + rpcAddr + "/events"

	node *nm.Node

	mempoolCount = 0

	userAddr                   = "D7DFF9806078899C8DA3FE3633CC0BF3C6C2B1BB"
	userPriv                   = "FDE3BD94CB327D19464027BA668194C5EFA46AE83E8419D7542CFF41F00C81972239C21C81EA7173A6C489145490C015E05D4B97448933B708A7EC5B7B4921E3"
	userPub                    = "2239C21C81EA7173A6C489145490C015E05D4B97448933B708A7EC5B7B4921E3"
	userByteAddr, userBytePriv = initUserBytes()

	clients = map[string]cclient.Client{
		"JSONRPC": cclient.NewClient(requestAddr, "JSONRPC"),
		"HTTP":    cclient.NewClient(requestAddr, "HTTP"),
	}
)

// returns byte versions of address and private key
// type [64]byte needed by account.GenPrivAccountFromKey
func initUserBytes() ([]byte, [64]byte) {
	byteAddr, _ := hex.DecodeString(userAddr)
	var byteKey [64]byte
	userPrivByteSlice, _ := hex.DecodeString(userPriv)
	copy(byteKey[:], userPrivByteSlice)
	return byteAddr, byteKey
}

func decodeHex(hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return bytes
}

// create a new node and sleep forever
func newNode(ready chan struct{}) {
	// Create & start node
	node = nm.NewNode()
	l := p2p.NewDefaultListener("tcp", config.App().GetString("ListenAddr"), false)
	node.AddListener(l)
	node.Start()

	// Run the RPC server.
	node.StartRPC()
	ready <- struct{}{}

	// Sleep forever
	ch := make(chan struct{})
	<-ch
}

// initialize config and create new node
func init() {
	rootDir := ".tendermint"
	config.Init(rootDir)
	app := config.App()
	app.Set("SeedNode", "")
	app.Set("DB.Backend", "memdb")
	app.Set("RPC.HTTP.ListenAddr", rpcAddr)
	app.Set("GenesisFile", rootDir+"/genesis.json")
	app.Set("PrivValidatorFile", rootDir+"/priv_validator.json")
	app.Set("Log.Stdout.Level", "debug")
	config.SetApp(app)
	logger.Reset()

	// Save new priv_validator file.
	priv := &state.PrivValidator{
		Address: decodeHex(userAddr),
		PubKey:  account.PubKeyEd25519(decodeHex(userPub)),
		PrivKey: account.PrivKeyEd25519(decodeHex(userPriv)),
	}
	priv.SetFile(rootDir + "/priv_validator.json")
	priv.Save()

	consensus.RoundDuration0 = 3 * time.Second
	consensus.RoundDurationDelta = 1 * time.Second

	// start a node
	ready := make(chan struct{})
	go newNode(ready)
	<-ready
}

//-------------------------------------------------------------------------------
// make transactions

// make a send tx (uses get account to figure out the nonce)
func makeSendTx(t *testing.T, typ string, from, to []byte, amt uint64) *types.SendTx {
	acc := getAccount(t, typ, from)
	nonce := 0
	if acc != nil {
		nonce = int(acc.Sequence) + 1
	}
	bytePub, err := hex.DecodeString(userPub)
	if err != nil {
		t.Fatal(err)
	}
	tx := &types.SendTx{
		Inputs: []*types.TxInput{
			&types.TxInput{
				Address:   from,
				Amount:    amt,
				Sequence:  uint(nonce),
				Signature: account.SignatureEd25519{},
				PubKey:    account.PubKeyEd25519(bytePub),
			},
		},
		Outputs: []*types.TxOutput{
			&types.TxOutput{
				Address: to,
				Amount:  amt,
			},
		},
	}
	return tx
}

// make a call tx (uses get account to figure out the nonce)
func makeCallTx(t *testing.T, typ string, from, to, data []byte, amt, gaslim, fee uint64) *types.CallTx {
	acc := getAccount(t, typ, from)
	nonce := 0
	if acc != nil {
		nonce = int(acc.Sequence) + 1
	}

	bytePub, err := hex.DecodeString(userPub)
	if err != nil {
		t.Fatal(err)
	}
	tx := &types.CallTx{
		Input: &types.TxInput{
			Address:   from,
			Amount:    amt,
			Sequence:  uint(nonce),
			Signature: account.SignatureEd25519{},
			PubKey:    account.PubKeyEd25519(bytePub),
		},
		Address:  to,
		GasLimit: gaslim,
		Fee:      fee,
		Data:     data,
	}
	return tx
}

// make transactions
//-------------------------------------------------------------------------------
// rpc call wrappers

// get the account
func getAccount(t *testing.T, typ string, addr []byte) *account.Account {
	client := clients[typ]
	ac, err := client.GetAccount(addr)
	if err != nil {
		t.Fatal(err)
	}
	return ac.Account
}

// make and sign transaction
func signTx(t *testing.T, typ string, fromAddr, toAddr, data []byte, key [64]byte, amt, gaslim, fee uint64) (types.Tx, *account.PrivAccount) {
	var tx types.Tx
	if data == nil {
		tx = makeSendTx(t, typ, fromAddr, toAddr, amt)
	} else {
		tx = makeCallTx(t, typ, fromAddr, toAddr, data, amt, gaslim, fee)
	}

	privAcc := account.GenPrivAccountFromKey(key)
	if bytes.Compare(privAcc.PubKey.Address(), fromAddr) != 0 {
		t.Fatal("Faield to generate correct priv acc")
	}

	client := clients[typ]
	resp, err := client.SignTx(tx, []*account.PrivAccount{privAcc})
	if err != nil {
		t.Fatal(err)
	}
	return resp.Tx, privAcc
}

// create, sign, and broadcast a transaction
func broadcastTx(t *testing.T, typ string, fromAddr, toAddr, data []byte, key [64]byte, amt, gaslim, fee uint64) (types.Tx, ctypes.Receipt) {
	tx, _ := signTx(t, typ, fromAddr, toAddr, data, key, amt, gaslim, fee)
	client := clients[typ]
	resp, err := client.BroadcastTx(tx)
	if err != nil {
		t.Fatal(err)
	}
	mempoolCount += 1
	return tx, resp.Receipt
}

// dump all storage for an account. currently unused
func dumpStorage(t *testing.T, addr []byte) ctypes.ResponseDumpStorage {
	client := clients["HTTP"]
	resp, err := client.DumpStorage(addr)
	if err != nil {
		t.Fatal(err)
	}
	return *resp
}

func getStorage(t *testing.T, typ string, addr, key []byte) []byte {
	client := clients[typ]
	resp, err := client.GetStorage(addr, key)
	if err != nil {
		t.Fatal(err)
	}
	return resp.Value
}

func callCode(t *testing.T, client cclient.Client, code, data, expected []byte) {
	resp, err := client.CallCode(code, data)
	if err != nil {
		t.Fatal(err)
	}
	ret := resp.Return
	// NOTE: we don't flip memory when it comes out of RETURN (?!)
	if bytes.Compare(ret, LeftPadWord256(expected).Bytes()) != 0 {
		t.Fatalf("Conflicting return value. Got %x, expected %x", ret, expected)
	}
}

func callContract(t *testing.T, client cclient.Client, address, data, expected []byte) {
	resp, err := client.Call(address, data)
	if err != nil {
		t.Fatal(err)
	}
	ret := resp.Return
	// NOTE: we don't flip memory when it comes out of RETURN (?!)
	if bytes.Compare(ret, LeftPadWord256(expected).Bytes()) != 0 {
		t.Fatalf("Conflicting return value. Got %x, expected %x", ret, expected)
	}
}

//--------------------------------------------------------------------------------
// utility verification function

func checkTx(t *testing.T, fromAddr []byte, priv *account.PrivAccount, tx *types.SendTx) {
	if bytes.Compare(tx.Inputs[0].Address, fromAddr) != 0 {
		t.Fatal("Tx input addresses don't match!")
	}

	signBytes := account.SignBytes(tx)
	in := tx.Inputs[0] //(*types.SendTx).Inputs[0]

	if err := in.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	// Check signatures
	// acc := getAccount(t, byteAddr)
	// NOTE: using the acc here instead of the in fails; it is nil.
	if !in.PubKey.VerifyBytes(signBytes, in.Signature) {
		t.Fatal(types.ErrTxInvalidSignature)
	}
}

// simple contract returns 5 + 6 = 0xb
func simpleContract() ([]byte, []byte, []byte) {
	// this is the code we want to run when the contract is called
	contractCode := []byte{0x60, 0x5, 0x60, 0x6, 0x1, 0x60, 0x0, 0x52, 0x60, 0x20, 0x60, 0x0, 0xf3}
	// the is the code we need to return the contractCode when the contract is initialized
	lenCode := len(contractCode)
	// push code to the stack
	//code := append([]byte{byte(0x60 + lenCode - 1)}, RightPadWord256(contractCode).Bytes()...)
	code := append([]byte{0x7f}, RightPadWord256(contractCode).Bytes()...)
	// store it in memory
	code = append(code, []byte{0x60, 0x0, 0x52}...)
	// return whats in memory
	//code = append(code, []byte{0x60, byte(32 - lenCode), 0x60, byte(lenCode), 0xf3}...)
	code = append(code, []byte{0x60, byte(lenCode), 0x60, 0x0, 0xf3}...)
	// return init code, contract code, expected return
	return code, contractCode, LeftPadBytes([]byte{0xb}, 32)
}

// simple call contract calls another contract
func simpleCallContract(addr []byte) ([]byte, []byte, []byte) {
	gas1, gas2 := byte(0x1), byte(0x1)
	value := byte(0x1)
	inOff, inSize := byte(0x0), byte(0x0) // no call data
	retOff, retSize := byte(0x0), byte(0x20)
	// this is the code we want to run (call a contract and return)
	contractCode := []byte{0x60, retSize, 0x60, retOff, 0x60, inSize, 0x60, inOff, 0x60, value, 0x73}
	contractCode = append(contractCode, addr...)
	contractCode = append(contractCode, []byte{0x61, gas1, gas2, 0xf1, 0x60, 0x20, 0x60, 0x0, 0xf3}...)

	// the is the code we need to return; the contractCode when the contract is initialized
	// it should copy the code from the input into memory
	lenCode := len(contractCode)
	memOff := byte(0x0)
	inOff = byte(0xc) // length of code before codeContract
	length := byte(lenCode)

	code := []byte{0x60, length, 0x60, inOff, 0x60, memOff, 0x37}
	// return whats in memory
	code = append(code, []byte{0x60, byte(lenCode), 0x60, 0x0, 0xf3}...)
	code = append(code, contractCode...)
	// return init code, contract code, expected return
	return code, contractCode, LeftPadBytes([]byte{0xb}, 32)
}
