package crypto

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/randentropy"
	"reflect"
	"testing"
)

func TestKeyStorePlain(t *testing.T) {
	ks := NewKeyStorePlain(common.DefaultDataDir())
	pass := "" // not used but required by API
	k1, err := ks.GenerateNewKey(randentropy.Reader, pass)
	if err != nil {
		t.Fatal(err)
	}

	k2 := new(Key)
	k2, err = ks.GetKey(k1.Address, pass)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(k1.Address, k2.Address) {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(k1.PrivateKey, k2.PrivateKey) {
		t.Fatal(err)
	}

	err = ks.DeleteKey(k2.Address, pass)
	if err != nil {
		t.Fatal(err)
	}
}

func TestKeyStorePassphrase(t *testing.T) {
	ks := NewKeyStorePassphrase(common.DefaultDataDir())
	pass := "foo"
	k1, err := ks.GenerateNewKey(randentropy.Reader, pass)
	if err != nil {
		t.Fatal(err)
	}
	k2 := new(Key)
	k2, err = ks.GetKey(k1.Address, pass)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(k1.Address, k2.Address) {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(k1.PrivateKey, k2.PrivateKey) {
		t.Fatal(err)
	}

	err = ks.DeleteKey(k2.Address, pass) // also to clean up created files
	if err != nil {
		t.Fatal(err)
	}
}

func TestKeyStorePassphraseDecryptionFail(t *testing.T) {
	ks := NewKeyStorePassphrase(common.DefaultDataDir())
	pass := "foo"
	k1, err := ks.GenerateNewKey(randentropy.Reader, pass)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ks.GetKey(k1.Address, "bar") // wrong passphrase
	if err == nil {
		t.Fatal(err)
	}

	err = ks.DeleteKey(k1.Address, "bar") // wrong passphrase
	if err == nil {
		t.Fatal(err)
	}

	err = ks.DeleteKey(k1.Address, pass) // to clean up
	if err != nil {
		t.Fatal(err)
	}
}
