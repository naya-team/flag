package server

import (
	"net/http"

	"github.com/naya-team/flag"
)

func New(mux *http.ServeMux, _flag flag.Flagger) {

	handler := &flagHandler{flag: _flag}

	mux.HandleFunc("GET /flags", handler.GetFlags)
	mux.HandleFunc("POST /flag", handler.StoreFlag)
	mux.HandleFunc("PUT /flag/:flag/:action", handler.DisableEnableFlag)
}
