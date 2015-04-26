/*
	This file is part of go-ethereum

	go-ethereum is free software: you can redistribute it and/or modify
	it under the terms of the GNU Lesser General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	go-ethereum is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU Lesser General Public License
	along with go-ethereum.  If not, see <http://www.gnu.org/licenses/>.
*/
/**
 * @authors
 * 	Gustav Simonsson <gustav.simonsson@gmail.com>
 *	Ethan Buchman <ethan@erisindustries.com> (adapt for ed25519 keys also)
 * @date 2015
 *
 */

package crypto

import (
	"encoding/json"
	"fmt"

	"code.google.com/p/go-uuid/uuid"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/ed25519"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/tendermint/tendermint/account"
	"github.com/eris-ltd/epm-go/crypto/randentropy"
	"github.com/eris-ltd/epm-go/crypto/secp256k1"
)

type KeyType uint8

func (k KeyType) String() string {
	switch k {
	case KeyTypeSecp256k1:
		return "secp256k1"
	case KeyTypeEd25519:
		return "ed25519"
	default:
		return "unknown"
	}
}

func keyTypeFromString(s string) (KeyType, error) {
	switch s {
	case "secp256k1":
		return KeyTypeSecp256k1, nil
	case "ed25519":
		return KeyTypeEd25519, nil
	default:
		var k KeyType
		return k, fmt.Errorf("unknown key type %s", s)
	}
}

const (
	KeyTypeSecp256k1 KeyType = iota
	KeyTypeEd25519
)

type Key struct {
	Id uuid.UUID // Version 4 "random" for unique id not derived from key data
	// key may be secp256k1 or ed25519 or potentially others
	Type KeyType
	// to simplify lookups we also store the address
	Address []byte
	// we only store privkey as pubkey/address can be derived from it
	// privkey in this struct is always in plaintext
	PrivateKey []byte
}

func (k *Key) Sign(hash []byte) ([]byte, error) {
	switch k.Type {
	case KeyTypeSecp256k1:
		return signSecp256k1(k, hash)
	case KeyTypeEd25519:
		return signEd25519(k, hash)
	}
	return nil, fmt.Errorf("invalid key type %v", k.Type)

}

func (k *Key) Pubkey() ([]byte, error) {
	switch k.Type {
	case KeyTypeSecp256k1:
		return pubKeySecp256k1(k)
	case KeyTypeEd25519:
		return pubKeyEd25519(k)
	}
	return nil, fmt.Errorf("invalid key type %v", k.Type)
}

type plainKeyJSON struct {
	Id         []byte
	Type       string
	Address    []byte
	PrivateKey []byte
}

type cipherJSON struct {
	Salt       []byte
	Nonce      []byte
	CipherText []byte
}

type encryptedKeyJSON struct {
	Id      []byte
	Type    string
	Address []byte
	Crypto  cipherJSON
}

func (k *Key) MarshalJSON() (j []byte, err error) {
	jStruct := plainKeyJSON{
		k.Id,
		k.Type.String(),
		k.Address,
		k.PrivateKey,
	}
	j, err = json.Marshal(jStruct)
	return j, err
}

func (k *Key) UnmarshalJSON(j []byte) (err error) {
	keyJSON := new(plainKeyJSON)
	err = json.Unmarshal(j, &keyJSON)
	if err != nil {
		return err
	}

	u := new(uuid.UUID)
	*u = keyJSON.Id
	k.Id = *u
	k.Address = keyJSON.Address
	k.PrivateKey = keyJSON.PrivateKey

	return err
}

func NewKey(typ KeyType) (*Key, error) {
	switch typ {
	case KeyTypeSecp256k1:
		return newKeySecp256k1(), nil
	case KeyTypeEd25519:
		return newKeyEd25519(), nil
	default:
		return nil, fmt.Errorf("Unknown key type: %v", typ)
	}
}

func newKeySecp256k1() *Key {
	priv, pub := secp256k1.GenerateKeyPair()
	return &Key{
		Id:         uuid.NewRandom(),
		Type:       KeyTypeSecp256k1,
		Address:    Sha3(pub[1:])[12:],
		PrivateKey: priv,
	}
}

func newKeyEd25519() *Key {
	privKeyBytes := new([64]byte)
	randBytes := randentropy.GetEntropyMixed(32)
	copy(privKeyBytes[:32], randBytes)
	pubKeyBytes := ed25519.MakePublicKey(privKeyBytes)
	pubKey := account.PubKeyEd25519(pubKeyBytes[:])
	return &Key{
		Id:         uuid.NewRandom(),
		Type:       KeyTypeEd25519,
		Address:    pubKey.Address(),
		PrivateKey: privKeyBytes[:],
	}
}

func pubKeySecp256k1(k *Key) ([]byte, error) {
	return secp256k1.GeneratePubKey(k.PrivateKey)
}

func pubKeyEd25519(k *Key) ([]byte, error) {
	priv := k.PrivateKey
	var privKeyBytes [64]byte
	copy(priv, privKeyBytes[:])
	pubKeyBytes := ed25519.MakePublicKey(&privKeyBytes)
	return pubKeyBytes[:], nil
}

func signSecp256k1(k *Key, hash []byte) ([]byte, error) {
	return secp256k1.Sign(hash, k.PrivateKey)
}

func signEd25519(k *Key, hash []byte) ([]byte, error) {
	priv := k.PrivateKey
	var privKeyBytes [64]byte
	copy(priv, privKeyBytes[:])
	privKey := account.PrivKeyEd25519(privKeyBytes[:])
	sig := privKey.Sign(hash)
	sigB := sig.([]byte)
	return sigB, nil
}
