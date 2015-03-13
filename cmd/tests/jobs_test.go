package main

import (
	"encoding/hex"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/thelonious/monk"
	"github.com/eris-ltd/epm-go/epm"
	"path"
	"testing"
	//"github.com/eris-ltd/thelonious/monkutil"
	"os"
)

var GoPath = os.Getenv("GOPATH")

func NewMonkModule() *monk.MonkModule {
	epm.ErrMode = epm.FailOnErr
	m := monk.NewMonk(nil)
	m.Config.RootDir = ".ethchain"
	m.Config.LogLevel = 5
	m.Config.GenesisConfig = "genesis.json"
	g := m.LoadGenesis(m.Config.GenesisConfig)
	g.Difficulty = 14
	m.SetGenesis(g)
	m.Init()
	m.Config.Mining = false
	m.Start()
	return m
}

/*
   For direct coding of hardcoded contracts and test results.
   See definitions and contracts for context
*/

func newEpmTest(t *testing.T, pdx string) (*epm.EPM, epm.Blockchain) {
	defaultContractPath := epm.ContractPath
	m := NewMonkModule()
	epm.ContractPath = defaultContractPath
	e, err := epm.NewEPM(m, ".epm-log-test")
	if err != nil {
		t.Error(err)
	}

	if err := e.Parse(pdx); err != nil {
		t.Error(err)
	}

	if err := e.ExecuteJobs(); err != nil {
		t.Error(err)
	}

	return e, m
}

func TestDeploy(t *testing.T) {
	e, m := newEpmTest(t, path.Join(epm.TestPath, "test_deploy.epm"))

	addr := e.Vars()["addr"]
	//fmt.Println("addr", addr)
	//0x60, 5050

	e.Commit()
	got := m.StorageAt(addr, "0x60")
	if got != "5050" {
		t.Error("got:", got, "expected:", "0x5050")
	}
	m.Shutdown()
}

func TestModifyDeploy(t *testing.T) {
	e, m := newEpmTest(t, path.Join(epm.TestPath, "test_modify_deploy.epm"))

	addr := e.Vars()["doug"]
	addr2 := e.Vars()["doug2"]
	//fmt.Println("doug addr", addr)
	//fmt.Println("doug addr2", addr2)
	//0x60, 0x5050

	e.Commit()
	got1 := m.StorageAt(addr, "0x60")
	if got1 != "5050" {
		t.Error("got:", got1, "expected:", "0x5050")
	}
	got2 := m.StorageAt(addr2, "0x60")
	if len(got2) < 2 || got2 != addr[2:] {
		t.Error("got:", got2, "expected:", addr)
	}
	m.Shutdown()
}

// doesn't work unless we wait a block until actually making the query
// not going to fly here
func iTestQuery(t *testing.T) {
	e, _ := newEpmTest(t, path.Join(epm.TestPath, "test_query.epm"))

	e.Commit()
	a := e.Vars()["B"]
	if a != "0x5050" {
		t.Error("got:", a, "expecxted:", "0x5050")
	}
}

func TestStack(t *testing.T) {
	e, m := newEpmTest(t, path.Join(epm.TestPath, "test_parse.epm"))

	addr1 := e.Vars()["A"]
	addr2 := e.Vars()["B"]
	addr3 := e.Vars()["D"]
	// fmt.Println("addr", addr1)
	// fmt.Println("addr2", addr2)
	// fmt.Println("addr3", addr3)
	//0x60, 0x5050

	e.Commit()
	got := m.StorageAt(addr2, addr1)
	if got != "15" {
		t.Error("got:", got, "expected:", "0x15")
	}
	got = m.StorageAt(addr3, "0x43")
	if got != "8080" {
		t.Error("got:", got, "expected:", "0x8080")
	}
	got = m.StorageAt(addr3, addr1)
	if got != "15" {
		t.Error("got:", got, "expected:", "0x15")
	}
	got = m.StorageAt(addr2, "0x12")
	exp := hex.EncodeToString([]byte("ethan"))
	if got != exp {
		t.Error("got:", got, "expected:", exp)
	}
	m.Shutdown()
}

func TestKV(t *testing.T) {
	e, m := newEpmTest(t, path.Join(epm.TestPath, "test_kv.epm"))

	addr1 := e.Vars()["A"]
	e.Commit()
	got := m.StorageAt(addr1, "0x5")
	exp := hex.EncodeToString([]byte("ethan"))
	if got != exp {
		t.Error("got:", got, "expected:", exp)
	}
	m.Shutdown()

}

// not a real test since the diffs just print we don't have access to them programmatically yet
// TODO>..
func iTestDiff(t *testing.T) {
	m := NewMonkModule()
	e, _ := epm.NewEPM(m, ".epm-log-test")

	if err := e.Parse(path.Join(epm.TestPath, "test_diff.epm")); err != nil {
		t.Error(err)
	}

	e.Diff = true
	e.ExecuteJobs()

	e.Commit()
	m.Shutdown()
}

func TestPdt(t *testing.T) {
	e, m := newEpmTest(t, path.Join(epm.TestPath, "test_parse.epm"))
	e.Commit()
	_, err := e.Test(path.Join(epm.TestPath, "test_parse.epm-check"))
	if err != nil {
		t.Fatal(err)
	}
	m.Shutdown()
}

func TestInclude(t *testing.T) {
	e, m := newEpmTest(t, path.Join(epm.TestPath, "test_path_include.epm"))

	addr := e.Vars()["addr"]
	addr2 := e.Vars()["addr2"]
	addr3 := e.Vars()["addr3"]
	//fmt.Println("addr", addr)
	//0x60, 5050

	e.Commit()
	got := m.StorageAt(addr, "0x60")
	if got != "5050" {
		t.Error("got:", got, "expected:", "0x5050")
	}
	got = m.StorageAt(addr2, "0x60")
	if got != "5060" {
		t.Error("got:", got, "expected:", "0x5060")
	}
	got = m.StorageAt(addr3, "0x60")
	if got != "5060" {
		t.Error("got:", got, "expected:", "0x5060")
	}
	m.Shutdown()
}

func TestEPMxNamespace(t *testing.T) {
	e, m := newEpmTest(t, path.Join(epm.TestPath, "test_epmx_deploy.epm"))

	addr := e.Vars()["dep.addr"]
	e.Commit()
	got := m.StorageAt(addr, "0x3")
	if got != "04" {
		t.Error("got:", got, "expected:", "0x04")
	}
	m.Shutdown()
}
