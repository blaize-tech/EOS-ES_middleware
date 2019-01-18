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
		*params.Offset = 0
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
	search := client.Search().
		Index("action_traces").
		Query(query).
		Sort("receipt.global_sequence", true). //from old to recent
		From(int(*params.Pos))
	if *params.Offset > 0 { //TODO: check for ES max records
		search = search.Size(int(*params.Offset))
	}
	searchResult, err := search.Do(context.Background())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	fmt.Printf("Found a total of %d records\n", searchResult.Hits.TotalHits)

	result := GetActionsResult { Actions: []Action{} }
	for _, hit := range searchResult.Hits.Hits {
		if hit.Source == nil {
			continue
		}

		var source map[string]*json.RawMessage
		err = json.Unmarshal(*hit.Source, &source)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		var receipt map[string]*json.RawMessage
		err = json.Unmarshal(*source["receipt"], &receipt)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		action := Action { GlobalActionSeq: *receipt["global_sequence"],
			BlockNum: *source["block_num"], BlockTime: *source["block_time"],
			ActionTrace: *hit.Source }
		result.Actions = append(result.Actions, action)
	}
	b, err := json.Marshal(result)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    fmt.Fprintf(w, string(b))
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
	
	client, err := elastic.NewClient()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	getResult, err := client.MultiGet().
		Add(elastic.NewMultiGetItem().Index("transactions").Id(params.Id)).
		Add(elastic.NewMultiGetItem().Index("transaction_traces").Id(params.Id)).
		Do(context.Background())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if getResult == nil || getResult.Docs == nil || len(getResult.Docs) != 2 ||
		getResult.Docs[0].Error != nil || getResult.Docs[1].Error != nil {
		http.Error(w, "Failed to query ES", 500)
		return
	}
	docTx := getResult.Docs[0]
	docTxTrace := getResult.Docs[1]

	var txSource map[string]*json.RawMessage
	err = json.Unmarshal(*docTx.Source, &txSource)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var txTraceSource map[string]*json.RawMessage
	err = json.Unmarshal(*docTxTrace.Source, &txTraceSource)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	result := GetTransactionResult { Id: params.Id, Trx: make(map[string]json.RawMessage),
		BlockTime: *txTraceSource["block_time"], BlockNum: *txSource["block_num"],
		Traces: *txTraceSource["action_traces"] }
	trx := make(map[string]json.RawMessage)
	trx["expiration"] = *txSource["expiration"]
	trx["ref_block_num"] = *txSource["ref_block_num"]
	trx["ref_block_prefix"] = *txSource["ref_block_prefix"]
	trx["max_net_usage_words"] = *txSource["max_net_usage_words"]
	trx["max_cpu_usage_ms"] = *txSource["max_cpu_usage_ms"]
	trx["delay_sec"] = *txSource["delay_sec"]
	trx["context_free_actions"] = *txSource["context_free_actions"]
	trx["actions"] = *txSource["actions"]
	trx["transaction_extensions"] = *txSource["transaction_extensions"]
	trx["signatures"] = *txSource["signatures"]
	trx["context_free_data"] = *txSource["context_free_data"]
	byteTrx, err := json.Marshal(trx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	result.Trx["trx"] = byteTrx
	result.Trx["receipt"] = *txTraceSource["receipt"]
	b, err := json.Marshal(result)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    fmt.Fprintf(w, string(b))
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
		Do(context.Background())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

    fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	fmt.Printf("Found a total of %d records\n", searchResult.Hits.TotalHits)

	result := GetKeyAccountsResult { AccountNames: []json.RawMessage{} }
	for _, hit := range searchResult.Hits.Hits {
		if hit.Source == nil {
			continue
		}
		var objmap map[string]*json.RawMessage
		err := json.Unmarshal(*hit.Source, &objmap)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		result.AccountNames = append(result.AccountNames, *objmap["name"])
	}
	b, err := json.Marshal(result)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    fmt.Fprintf(w, string(b))
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

	client, err := elastic.NewClient()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("name", params.ControllingAccount)) //Is it better to convert name to number and search by id?
	searchResult, err := client.Search().
		Index("accounts").
		Query(query).
		Do(context.Background())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
    
    fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	fmt.Printf("Found a total of %d records\n", searchResult.Hits.TotalHits)

	result := GetControlledAccountsResult { AccountNames: []json.RawMessage{} }
	for _, hit := range searchResult.Hits.Hits {
		if hit.Source == nil {
			continue
		}
		var objmap map[string]*json.RawMessage
		err := json.Unmarshal(*hit.Source, &objmap)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		var accounts []json.RawMessage
		err = json.Unmarshal(*objmap["account_controls"], &accounts)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		result.AccountNames = append(result.AccountNames, accounts...)
	}
	b, err := json.Marshal(result)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    fmt.Fprintf(w, string(b))
}