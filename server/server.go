package server

import (
	"fmt"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/go-martini/martini"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/thelonious/monklog"
	// "log"
	// "os"
)

var logger *monklog.Logger = monklog.NewLogger("EPM-CLI")

// The server object.
type Server struct {
	// The maximum number of active connections that the server allows.
	maxConnections uint32
	// The host.
	host string
	// The port.
	port uint16
	// The root, or serving directory.
	rootDir string
	// The classic martini instance.
	cMartini *martini.ClassicMartini
	// The http service.
	httpService *HttpService
	// The websocket service.
	wsService *WsService
}

// Create a new server.
func NewServer(host string, port uint16, maxConnections uint32, rootDir string) *Server {

	cMartini := martini.Classic()

	// TODO remember to change to martini.Prod
	// better to not set this here and just run from
	// env vars per
	// https://github.com/go-martini/martini/blob/master/env.go#L14
	martini.Env = martini.Dev

	httpService := NewHttpService()
	wsService := NewWsService(maxConnections)

	return &Server{
		maxConnections,
		host,
		port,
		rootDir,
		cMartini,
		httpService,
		wsService,
	}
}

// Start running the server.
func (this *Server) Start() error {

	cm := this.cMartini

	// Static.
	cm.Use(martini.Static(this.rootDir))

	// Simple echo for testing http
	cm.Get("/echo/:message", this.httpService.handleEcho)

	// Informational commands
	cm.Get("/eris/plop/:chainName/:toPlop", this.httpService.handlePlop)
	cm.Get("/eris/refs/ls", this.httpService.handleLsRefs)
	cm.Post("/eris/refs/add/:chainName/:chainType/:chainType", this.httpService.handleAddRefs)
	cm.Post("/eris/refs/rm/:chainName", this.httpService.handleRmRefs)

	// Chain management commands
	cm.Post("/eris/checkout/:chainName", this.httpService.handleCheckout)
	cm.Post("/eris/clean/:chainName", this.httpService.handleClean)
	cm.Post("/eris/fetch/:chainName/:fetchIP/:fetchPort", this.httpService.handleFetch)
	cm.Post("/eris/new/:chainName", this.httpService.handleNewChain)
	cm.Post("/eris/config/:chainName", this.httpService.handleConfig)
	cm.Post("/eris/rawconfig/:chainName", this.httpService.handleRawConfig)

	// Blockchain client commands
	cm.Post("/eris/start", this.httpService.handleStartChain)
	cm.Post("/eris/stop", this.httpService.handleStopChain)
	cm.Post("/eris/restart", this.httpService.handleRestartChain)

	// Key import
	cm.Post("/eris/importkey/:keyName", this.httpService.handleKeyImport)

	// Handle websocket negotiation requests.
	cm.Get("/ws", this.wsService.handleWs)

	// Default 404 message.
	cm.NotFound(this.httpService.handleNotFound)

	cm.RunOnAddr(this.host + ":" + fmt.Sprintf("%d", this.port))

	return nil
}

// Get the maximum number of active connections/sessions that the server allows.
func (this *Server) MaxConnections() uint32 {
	return this.maxConnections
}

// Get the root, or served directory.
func (this *Server) RootDir() string {
	return this.rootDir
}

// Get the http service object.
func (this *Server) HttpService() *HttpService {
	return this.httpService
}

// Get the websocket service object.
func (this *Server) WsService() *WsService {
	return this.wsService
}
