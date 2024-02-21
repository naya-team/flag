package server

import (
	"encoding/json"
	"net/http"

	"github.com/naya-team/flag"
)

type flagHandler struct {
	flag flag.Flagger
}

func (h *flagHandler) StoreFlag(w http.ResponseWriter, r *http.Request) {

	var payload CreateFlag

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	flag, err := h.flag.StoreFlag(payload.Flag, payload.IsEnable)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flag)
	w.WriteHeader(http.StatusCreated)
}

func (h *flagHandler) GetFlags(w http.ResponseWriter, r *http.Request) {

	// get flag from query
	flag := r.URL.Query().Get("flag")

	flags, err := h.flag.GetFlags(flag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flags)
	w.WriteHeader(http.StatusOK)
}

func (h *flagHandler) DisableEnableFlag(w http.ResponseWriter, r *http.Request) {

	_flag := r.PathValue("flag")
	action := r.PathValue("action")

	var (
		flag flag.ReleaseFlagModel
		err  error
	)

	if action == "disable" {
		flag, err = h.flag.DisableFlag(_flag)

	} else if action == "enable" {
		flag, err = h.flag.EnableFlag(_flag)
	} else {
		http.Error(w, "action not found", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flag)
	w.WriteHeader(http.StatusOK)
}
