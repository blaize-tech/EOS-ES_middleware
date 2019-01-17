package main

import (
	"fmt"
    "net/http"
    "io/ioutil"
	"encoding/json"
)


func onlyGet(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if (r.Method != http.MethodGet) {
			http.Error(w, "Invalid request method.", 405)
			return
		}
		h(w, r)
	}
}


func getActionsHandler(w http.ResponseWriter, r *http.Request) {
    bytes, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, err.Error(), 500)
		return
    }

	var params GetActionsParams
    err = json.Unmarshal(bytes, &params)
	if err != nil {
		http.Error(w, "Invalid arguments.", 400)
		return
	}

    fmt.Fprintf(w, "Actions requested")
}

func getTransactionHandler(w http.ResponseWriter, r *http.Request) {
    bytes, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, err.Error(), 500)
		return
    }

	var params GetTransactionParams
    err = json.Unmarshal(bytes, &params)
	if err != nil {
		http.Error(w, "Invalid arguments.", 400)
		return
    }
    
    fmt.Fprintf(w, "Transaction requested")
}

func getKeyAccountsHandler(w http.ResponseWriter, r *http.Request) {
    bytes, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, err.Error(), 500)
		return
    }

	var params GetKeyAccountsParams
    err = json.Unmarshal(bytes, &params)
	if err != nil {
		http.Error(w, "Invalid arguments.", 400)
		return
    }
    
    fmt.Fprintf(w, "Key accounts requested")
}

func getControlledAccountsHandler(w http.ResponseWriter, r *http.Request) {
    bytes, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, err.Error(), 500)
		return
    }

	var params GetControlledAccountsParams
    err = json.Unmarshal(bytes, &params)
	if err != nil {
		http.Error(w, "Invalid arguments.", 400)
		return
    }
    
    fmt.Fprintf(w, "Controlled accounts requested")
}