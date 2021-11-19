// httpexecute in Go. Copyright (C) Kost. Distributed under MIT.
// RESTful interface to your operating system shell

package httpexecute

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

// CmdReq holds JSON input request.
type CmdReq struct {
	Cmd    string
	Nojson bool
	Stdin  string
}

// CmdResp holds JSON output request.
type CmdResp struct {
	Cmd    string
	Stdout string
	Stderr string
	Err    string
}

type CmdConfig struct {
	VerboseLevel int
	SilentOutput bool
	Log *log.Logger
}

// real content Handler
func (cc *CmdConfig) ExecuteHandler(w http.ResponseWriter, r *http.Request) {
	var jsonout bool
	var inputjson CmdReq
	var outputjson CmdResp
	var body []byte
	if r.Header.Get("Content-Type") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		jsonout = true
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}
	cmdstr := ""
	urlq, urlErr := url.QueryUnescape(r.URL.RawQuery)
	if urlErr != nil {
		cc.Log.Printf("url query unescape: %v", urlErr)
	}
	if r.Method == "GET" || r.Method == "HEAD" {
		cmdstr = urlq
	}
	if r.Method == "POST" {
		var rerr error
		body, rerr = ioutil.ReadAll(r.Body)
		if rerr != nil {
			cc.Log.Printf("read Body: %v", rerr)
		}
		if closeErr := r.Body.Close(); closeErr != nil {
			cc.Log.Printf("body close: %v", closeErr)

		}
		if cc.VerboseLevel > 2 {
			cc.Log.Printf("Body: %s", body)
		}

		if len(urlq) > 0 {
			cmdstr = urlq
		} else {
			if jsonout {
				jerr := json.Unmarshal(body, &inputjson)
				if jerr != nil {
					// http.Error(w, jerr.Error(), 400)
					return
				}
				cmdstr = inputjson.Cmd
				jsonout = !inputjson.Nojson
			} else {
				cmdstr = string(body)
			}
		}
	}
	if cc.VerboseLevel > 0 {
		log.Printf("Command to execute: %s", cmdstr)
	}

	if len(cmdstr) < 1 {
		return
	}

	parts := strings.Fields(cmdstr)
	head := parts[0]
	parts = parts[1:]

	cmd := exec.Command(head, parts...)

	// Handle stdin if have any
	if len(urlq) > 0 && r.Method == "POST" {
		if cc.VerboseLevel > 2 {
			cc.Log.Printf("Stdin: %s", body)
		}
		cmd.Stdin = bytes.NewReader(body)
	}
	if len(inputjson.Stdin) > 0 {
		if cc.VerboseLevel > 2 {
			cc.Log.Printf("JSON Stdin: %s", inputjson.Stdin)
		}
		cmd.Stdin = strings.NewReader(inputjson.Stdin)
	}

	var err error
	var jStdout bytes.Buffer
	var jStderr bytes.Buffer
	if r.Method == "HEAD" {
		err = cmd.Start()
	} else {
		if jsonout {
			cmd.Stdout = &jStdout
			cmd.Stderr = &jStderr
		} else {
			cmd.Stdout = w
			cmd.Stderr = w
		}
		err = cmd.Run()
	}
	if err != nil {
		if cc.VerboseLevel > 0 {
			cc.Log.Printf("Error executing: %s", err)
		}
		if jsonout {
			outputjson.Err = err.Error()
		} else {
			if !cc.SilentOutput {
				_, writeErr := w.Write([]byte(err.Error()))
				if writeErr != nil {
					cc.Log.Printf("write: %v", writeErr)
				}
			}
		}
	}

	if jsonout {
		outputjson.Stdout = jStdout.String()
		outputjson.Stderr = jStderr.String()
		outputjson.Cmd = cmdstr
		if encodeErr := json.NewEncoder(w).Encode(outputjson); encodeErr != nil {
			cc.Log.Printf("encode: %v", err)
		}
	}
}

