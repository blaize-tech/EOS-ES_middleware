package main

import (
	"fmt"
    "net/http"
    "io/ioutil"
	"encoding/json"
)


//onlyGet take function (http handler) as an argument
//and returns function that takes http.ResponseWriter and *http.Request
//this function will call given handler only if http method of the request is GET
//otherwise it will respond with 405 error code
func onlyGet(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if (r.Method != http.MethodGet) {
			http.Error(w, "Invalid request method.", 405)
			return
		}
		h(w, r)
	}
}


//getActionsHandler is http handler that takes
//http.ResponseWriter and *http.Request as arguments
//it tries to parse parameters from request body
//and passes them to getActions()
//The result of getActions() is encoded and sent as a response
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
	if params.Pos == nil {
		params.Pos = new(int64)
		*params.Pos = 0
	}
	if params.Offset == nil {
		params.Offset = new(int64)
		*params.Offset = 0
	}

	result, err := getActions(params)
	if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
	b, err := json.Marshal(result)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    fmt.Fprintf(w, string(b))
}

//getTransactionHandler is http handler that takes
//http.ResponseWriter and *http.Request as arguments
//it tries to parse parameters from request body
//and passes them to getTransaction()
//The result of getTransaction() is encoded and sent as a response
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

	result, err := getTransaction(params)
	if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
	b, err := json.Marshal(result)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    fmt.Fprintf(w, string(b))
}

//getKeyAccountsHandler is http handler that takes
//http.ResponseWriter and *http.Request as arguments
//it tries to parse parameters from request body
//and passes them to getKeyAccounts()
//The result of getKeyAccounts() is encoded and sent as a response
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
	
	result, err := getKeyAccounts(params)
	if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
	b, err := json.Marshal(result)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    fmt.Fprintf(w, string(b))
}

//getControlledAccountsHandler is http handler that takes
//http.ResponseWriter and *http.Request as arguments
//it tries to parse parameters from request body
//and passes them to getControlledAccounts()
//The result of getControlledAccounts() is encoded and sent as a response
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

	result, err := getControlledAccounts(params)
	b, err := json.Marshal(result)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    fmt.Fprintf(w, string(b))
}