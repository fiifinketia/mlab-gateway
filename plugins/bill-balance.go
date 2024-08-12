// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ClientRegisterer is the symbol the plugin loader will try to load. It must implement the RegisterClient interface
var ClientRegisterer = registerer("bill-balance")

type registerer string

var logger Logger = nil

func (registerer) RegisterLogger(v interface{}) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", ClientRegisterer))
}

func (r registerer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r registerer) registerClients(_ context.Context, extra map[string]interface{}) (http.Handler, error) {
	// check the passed configuration and initialize the plugin
	name, ok := extra["name"].(string)
	if !ok {
		return nil, errors.New("wrong config")
	}
	if name != string(r) {
		return nil, fmt.Errorf("unknown register %s", name)
	}

	// check the cfg. If the modifier requires some configuration,
	// it should be under the name of the plugin. E.g.:
	/*
	   "extra_config":{
	       "plugin/http-client":{
	           "name":"bill-balance",
	           "bill-balance":{
	               "path": "/some-path"
	           }
	       }
	   }
	*/

	// The config variable contains all the keys you hace defined in the configuration:
	config, _ := extra["bill-balance"].(map[string]interface{})

	// The plugin will look for this path:
	path, _ := config["path"].(string)
	logger.Debug(fmt.Sprintf("The plugin is now hijacking the path %s", path))

	// return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http handler
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// The path matches, it has to be hijacked and no call to the backend happens.
		// The path is the the call to the backend, not the original request by the user.
		if req.URL.Path == path {
			w.Header().Add("Content-Type", "application/json")
			// Return a custom JSON object:
			// res := map[string]string{"message": html.EscapeString(req.URL.Path)}
			// b, _ := json.Marshal(res)
			// w.Write(b)
			// logger.Debug("request:", html.EscapeString(req.URL.Path))
			
			// Send request to /billings/balance if response is true continue
			action, _:= GetRequestAction(fmt.Sprintf("%s:%s", req.Method, req.URL.Path))
			inRequestBody, err := json.Marshal(req.Body)
			balanceReqBody := []byte(fmt.Sprintf(`{"action":"%s", "data": %s}`, action, inRequestBody))
			balanceReqReader := bytes.NewReader(balanceReqBody)

			balanceReq, err := http.NewRequest(http.MethodPost, "http://localhost:8090/api/billings/balance", balanceReqReader)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
			balanceReq.Header.Set("Content-Type", "application/json")
			balanceReq.Header.Set("Authorization", req.Header.Get("Authorization"))

			client := http.Client{
				Timeout: 30 * time.Second,
			}

			res, err := client.Do(balanceReq)
			if err!= nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                os.Exit(1)
            }

			bodyBytes, _ := io.ReadAll(res.Body)

			logger.Critical(bodyBytes)
			// defer res.Body.Close()

			
		}

		// If the requested path is not what we defined, continue.
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Copy headers, status codes, and body from the backend to the response writer
		for k, hs := range resp.Header {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}
		w.WriteHeader(resp.StatusCode)
		if resp.Body == nil {
			return
		}
		io.Copy(w, resp.Body)
		resp.Body.Close()

	}), nil
}

func main() {}

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}
