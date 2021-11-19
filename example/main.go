package main

import (
	"net/http"
	"github.com/kost/httpexecute"
	"log"
)

func main() {
	// initialize handler & logging
	he := &httpexecute.CmdConfig{Log: log.Default(), VerboseLevel: 1}

	// use it as standard handler for http
	http.HandleFunc("/execute", he.ExecuteHandler)
	http.ListenAndServe(":4455", nil)
}

