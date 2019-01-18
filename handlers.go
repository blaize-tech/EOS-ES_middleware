package main

import (
	"fmt"
    "net/http"
    "io/ioutil"
	"encoding/json"
	"github.com/olivere/elastic"
	"context"
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
	if params.Pos == nil {
		params.Pos = new(int32)
		*params.Pos = 0
	}
	if params.Offset == nil {
		params.Offset = new(int32)
		*params.Offset = -1
	}
	fmt.Printf("Pos %d\n", int(*params.Pos))
	fmt.Printf("Offset %d\n", int(*params.Offset))

	client, err := elastic.NewClient()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMultiMatchQuery(params.AccountName, "receipt.receiver", "act.authorization.actor"))
	searchResult, err := client.Search().
		Index("action_traces").
		Query(query).
		Sort("receipt.global_sequence", true). //from old to recent
		From(int(*params.Pos)).Size(int(*params.Offset)).
		Pretty(true).
		Do(context.Background())

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	fmt.Printf("Found a total of %d records\n", searchResult.Hits.TotalHits)

	b, err := json.Marshal(searchResult.Hits.Hits)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Fprintf(w, "{ \"actions\": \n" + string(b) + "}")
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
	
	client, err := elastic.NewClient()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("pub_keys.key", params.PublicKey))
	searchResult, err := client.Search().
		Index("accounts").
		Query(query).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

    fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	fmt.Printf("Found a total of %d records\n", searchResult.Hits.TotalHits)

	b, err := json.Marshal(searchResult.Hits.Hits)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Fprintf(w, "account_names: " + string(b))
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