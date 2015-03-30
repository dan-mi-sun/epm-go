package server

import (
	"fmt"
	"github.com/go-martini/martini"
	"log"
	"os"
)

var logger *log.Logger = log.New(os.Stdout, "[Server] ", log.LstdFlags)

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

	// Default 404 message.
	cm.NotFound(this.httpService.handleNotFound)

	// Handle websocket negotiation requests.
	cm.Get("/websocket", this.wsService.handleWs)

	// Simple echo for testing http
	cm.Get("/echo/:message", this.httpService.handleEcho)

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
