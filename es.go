package main

import (
	"errors"
	"encoding/json"
	"github.com/olivere/elastic"
	"context"
)

const AccountsIndex          string = "accounts"
const BlocksIndex            string = "blocks"
const TransactionsIndex      string = "transactions"
const TransactionTracesIndex string = "transaction_traces"
const ActionTracesIndex      string = "action_traces"


func getActions(params GetActionsParams) (*GetActionsResult, error) {
	client, err := elastic.NewClient()
	if err != nil {
		return nil, err
	}
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMultiMatchQuery(params.AccountName, "receipt.receiver", "act.authorization.actor"))
	search := client.Search().
		Index(ActionTracesIndex).
		Query(query).
		Sort("receipt.global_sequence", true). //from old to recent
		From(int(*params.Pos))
	if *params.Offset > 0 {
		search = search.Size(int(*params.Offset))
	}
	searchResult, err := search.Do(context.Background())
	if err != nil {
		return nil, err
	}

	result := new(GetActionsResult)
	for _, hit := range searchResult.Hits.Hits {
		if hit.Source == nil {
			continue
		}

		var source map[string]*json.RawMessage
		err = json.Unmarshal(*hit.Source, &source)
		if err != nil {
			continue
		}
		var receipt map[string]*json.RawMessage
		err = json.Unmarshal(*source["receipt"], &receipt)
		if err != nil {
			continue
		}
		action := Action { GlobalActionSeq: *receipt["global_sequence"],
			BlockNum: *source["block_num"], BlockTime: *source["block_time"],
			ActionTrace: *hit.Source }
		result.Actions = append(result.Actions, action)
	}
	return result, nil
}


func getTransaction(params GetTransactionParams) (*GetTransactionResult, error) {
	client, err := elastic.NewClient()
	if err != nil {
		return nil, err
	}
	getResult, err := client.MultiGet().
		Add(elastic.NewMultiGetItem().Index(TransactionsIndex).Id(params.Id)).
		Add(elastic.NewMultiGetItem().Index(TransactionTracesIndex).Id(params.Id)).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	if getResult == nil || getResult.Docs == nil || len(getResult.Docs) != 2 ||
		getResult.Docs[0].Error != nil || getResult.Docs[1].Error != nil {
		return nil, errors.New("Failed to query ES")
	}
	docTx := getResult.Docs[0]
	docTxTrace := getResult.Docs[1]

	if !(docTx.Found && docTxTrace.Found) {
		return nil, errors.New("Transaction not found")
	}

	var txSource map[string]*json.RawMessage
	err = json.Unmarshal(*docTx.Source, &txSource)
	if err != nil {
		return nil, err
	}
	var txTraceSource map[string]*json.RawMessage
	err = json.Unmarshal(*docTxTrace.Source, &txTraceSource)
	if err != nil {
		return nil, err
	}

	result := new(GetTransactionResult)
	result.Id = params.Id
	result.Trx = make(map[string]json.RawMessage)
	result.BlockTime = *txTraceSource["block_time"]
	result.BlockNum = *txSource["block_num"]
	result.Traces = *txTraceSource["action_traces"]
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
		return nil, err
	}
	result.Trx["trx"] = byteTrx
	result.Trx["receipt"] = *txTraceSource["receipt"] //TODO add packed_trx
	return result, nil
}


func getKeyAccounts(params GetKeyAccountsParams) (*GetKeyAccountsResult, error) {
	client, err := elastic.NewClient()
	if err != nil {
		return nil, err
	}
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("pub_keys.key", params.PublicKey))
	searchResult, err := client.Search().
		Index(AccountsIndex).
		Query(query).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	result := new(GetKeyAccountsResult)
	result.AccountNames = make([]json.RawMessage, 0)
	for _, hit := range searchResult.Hits.Hits {
		if hit.Source == nil {
			continue
		}
		var objmap map[string]*json.RawMessage
		err := json.Unmarshal(*hit.Source, &objmap)
		if err != nil {
			return nil, err
		}
		result.AccountNames = append(result.AccountNames, *objmap["name"])
	}
	return result, nil
}


func getControlledAccounts(params GetControlledAccountsParams) (*GetControlledAccountsResult, error) {
	client, err := elastic.NewClient()
	if err != nil {
		return nil, err
	}
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("name", params.ControllingAccount)) //Is it better to convert name to number and search by id?
	searchResult, err := client.Search().
		Index(AccountsIndex).
		Query(query).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	result := new(GetControlledAccountsResult)
	result.AccountNames = make([]json.RawMessage, 0)
	for _, hit := range searchResult.Hits.Hits {
		if hit.Source == nil {
			continue
		}
		var objmap map[string]*json.RawMessage
		err := json.Unmarshal(*hit.Source, &objmap)
		if err != nil {
			return nil, err
		}
		var accounts []json.RawMessage
		err = json.Unmarshal(*objmap["account_controls"], &accounts)
		if err != nil {
			return nil, err
		}
		result.AccountNames = append(result.AccountNames, accounts...)
	}
	return result, nil
}