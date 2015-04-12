#!/bin/bash
set -e
rm -f /tmp/success # in case its around

# install epm
# cd $GOPATH/src/github.com/eris-ltd/epm-go/cmd/epm
# go install
# epm init

cd $GOPATH/src/github.com/eris-ltd/epm-go

# XXX: set develop lllc server
export DECERVER=/home/eris/.decerver
echo `jq '.lll.url |= "http://ps.erisindustries.com:8090/compile"' $DECERVER/languages/config.json` > $DECERVER/languages/config.json
echo `jq '.se.url |= "http://ps.erisindustries.com:8090/compile"' $DECERVER/languages/config.json` > $DECERVER/languages/config.json
echo `jq '.sol.url |= "http://ps.erisindustries.com:8090/compile"' $DECERVER/languages/config.json` > $DECERVER/languages/config.json

# run the go unit tests
cd epm && go test -v ./... -race
cd ../chains && go test -v ./... -race
cd ../utils && go test -v ./... -race
cd ../cmd/epm && go test -v ./... -race # these don't exist yet

# run the base pdx deploy test
cd ../tests && go test -v ./... -race

# test suite of eris-std-lib deploys
cd $GOPATH/src/github.com/eris-ltd/eris-std-lib/DTT/tests
./test.sh

# grab a new eth chain and test serpent and solidity
epm --log 5 new -type eth -checkout
epm keys import $GOPATH/src/github.com/eris-ltd/epm-go/cmd/tests/tester-c5ac1950c7fa7f8f0abd54e83495425fc0c5fc2e
cd $GOPATH/src/github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/lllc-server/tests
epm --log 5 deploy test_serpent.pdx
epm --log 5 deploy test_solidity.pdx

# fig up doesn't return proper error codes, so this is our hack
touch /opt/success

