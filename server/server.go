package server

import (
	"fmt"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/go-martini/martini"
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/eris-ltd/thelonious/monklog"
	"net/http"
	"time"
	"log"
	"os"
)

var logger *monklog.Logger = monklog.NewLogger("EPM_SERVER")

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

func epmClassic() *martini.ClassicMartini {
	r := martini.NewRouter()
	m := martini.New()
	m.Use(ServeLogger())
	m.Use(martini.Recovery())
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return &martini.ClassicMartini{m, r}
}

// Create a new server.
func NewServer(host string, port uint16, maxConnections uint32, rootDir string) *Server {

	cMartini := epmClassic()
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

	// Informational handlers
	cm.Get("/eris/plop/:chainName/:toPlop", this.httpService.handlePlop)
	cm.Get("/eris/refs/ls", this.httpService.handleLsRefs)
	cm.Post("/eris/refs/add/:chainName/:chainType/:chainType", this.httpService.handleAddRefs)
	cm.Post("/eris/refs/rm/:chainName", this.httpService.handleRmRefs)

	// Chain management handlers
	cm.Post("/eris/config/:chainName", this.httpService.handleConfig)
	cm.Post("/eris/rawconfig/:chainName", this.httpService.handleRawConfig)
	cm.Post("/eris/checkout/:chainName", this.httpService.handleCheckout)
	cm.Post("/eris/clean/:chainName", this.httpService.handleClean)

	// Blockchain admin handlers
	cm.Post("/eris/fetch/:chainName/:fetchIP/:fetchPort", this.httpService.handleFetchChain)
	cm.Post("/eris/new/:chainName", this.httpService.handleNewChain)
	cm.Post("/eris/start", this.httpService.handleStartChain)
	cm.Post("/eris/stop", this.httpService.handleStopChain)
	cm.Post("/eris/restart", this.httpService.handleRestartChain)
	cm.Get("/eris/status", this.httpService.handleChainStatus)

	// Keys handlers
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

func ServeLogger() martini.Handler {
	out := log.New(os.Stdout, "", 0)
	return func(res http.ResponseWriter, req *http.Request, c martini.Context, log *log.Logger) {
		start := time.Now()

		addr := req.Header.Get("X-Real-IP")
		if addr == "" {
			addr = req.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = req.RemoteAddr
			}
		}

		out.Printf("%d/%02d/%02d %02d:%02d:%02d [EPM_SERVER] Started %s %s for %s\n", start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute(), start.Second(), req.Method, req.URL.Path, addr)

		rw := res.(martini.ResponseWriter)
		c.Next()

		stop := time.Now()
		out.Printf("%d/%02d/%02d %02d:%02d:%02d [EPM_SERVER] Completed %v %s in %v\n", stop.Year(), stop.Month(), stop.Day(), start.Hour(), start.Minute(), start.Second(), rw.Status(), http.StatusText(rw.Status()), time.Since(start))
	}
}