package main

import (
	"net/http"
	"log"
)

const ApiPath string = "/v1/history/"


func main() {
	http.HandleFunc(ApiPath + "get_actions", onlyGet(getActionsHandler))
	http.HandleFunc(ApiPath + "get_transaction", onlyGet(getTransactionHandler))
	http.HandleFunc(ApiPath + "get_key_accounts", onlyGet(getKeyAccountsHandler))
	http.HandleFunc(ApiPath + "get_controlled_accounts", onlyGet(getControlledAccountsHandler))
    err := http.ListenAndServe(":9000", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}