package main

import (
	"github.com/eris-ltd/decerver/server"
)

func main() {
	srvr := server.NewServer("localhost", 3000, 1000, ".")
	srvr.Start()
}
