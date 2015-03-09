#!/bin/bash
set -e
rm -f /tmp/success # in case its around

cd $GOPATH/src/github.com/eris-ltd/epm-go

# run the go unit tests
cd epm && go test -v ./... -race
cd ../chains && go test -v ./... -race
cd ../utils && go test -v ./... -race
cd ../cmd/epm && go test -v ./... -race # these don't exist yet

# run the base pdx deploy test
cd ../tests && go test -v ./... -race

# install epm
cd $GOPATH/src/github.com/eris-ltd/epm-go/cmd/epm
go install

# test suite of eris-std-lib deploys
cd $GOPATH/src/github.com/eris-ltd/eris-std-lib/DTT/tests
./test.sh

# fig up doesn't return proper error codes, so this is our hack
touch /tmp/success
