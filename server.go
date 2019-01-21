package main

import (
	"os"
	"fmt"
	"encoding/json"
	"net/http"
	"log"
)

const ConfigFilename string = "config.json"
const ApiPath string = "/v1/history/"


func main() {
	var config Config
	file, err := os.Open(ConfigFilename)
	if err != nil {
		fmt.Printf("Failed to open %s\n", ConfigFilename)
		return
	}
	decoder := json.NewDecoder(file) 
	err = decoder.Decode(&config) 
	if err != nil {
		fmt.Printf("Failed decode config\n")
		return
	}
	
	http.HandleFunc(ApiPath + "get_actions", onlyGet(getActionsHandler))
	http.HandleFunc(ApiPath + "get_transaction", onlyGet(getTransactionHandler))
	http.HandleFunc(ApiPath + "get_key_accounts", onlyGet(getKeyAccountsHandler))
	http.HandleFunc(ApiPath + "get_controlled_accounts", onlyGet(getControlledAccountsHandler))
    err = http.ListenAndServe(":" + fmt.Sprint(config.Port), nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}