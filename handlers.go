package main

import (
	"fmt"
	"net/http"
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
  fmt.Fprintf(w, "Actions requested")
}

func getTransactionHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Transaction requested")
}

func getKeyAccountsHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Key accounts requested")
}

func getControlledAccountsHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Controlled accounts requested")
}