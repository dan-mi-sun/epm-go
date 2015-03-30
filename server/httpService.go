package server

import (
	"fmt"
	"github.com/go-martini/martini"
	"net/http"
)

const EPM_HELP = "Wrong command, asshole."

type HttpService struct {
	// Maybe keep track of some statistics if this is used to create chains
	// via some Eris web service later, like it works with the compilers.
}

// Create a new http service
func NewHttpService() *HttpService {
	return &HttpService{}
}

// Handler for not found.
func (this *HttpService) handleNotFound(w http.ResponseWriter, r *http.Request) {
	this.writeMsg(w, 404, EPM_HELP)
}

// Handler for echo.
func (this *HttpService) handleEcho(params martini.Params, w http.ResponseWriter, r *http.Request) {
	this.writeMsg(w, 200, params["message"])
}

// Utility method for responding with an error.
func (this *HttpService) writeMsg(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, msg)
}

// ***************************** Add more handlers *********************************
