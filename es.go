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


func getActionTrace(client *elastic.Client, txId string, actionSeq uint64) (json.RawMessage, error) {
	getResult, err := client.Get().
		Index(TransactionTracesIndex).
		Id(txId).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	if !getResult.Found || getResult.Source == nil {
		return nil, errors.New("Action trace not found")
	}
	var txTraceSource map[string]*json.RawMessage
	err = json.Unmarshal(*getResult.Source, &txTraceSource)
	if err != nil || txTraceSource["action_traces"] == nil {
		return nil, errors.New("Failed to parse ES response")
	}
	var actionTraces []map[string]*json.RawMessage
	err = json.Unmarshal(*txTraceSource["action_traces"], &actionTraces)
	if err != nil {
		return nil, errors.New("Failed to parse ES response")
	}
	for _, trace := range actionTraces {
		if trace["receipt"] == nil {
			continue
		}
		var receipt map[string]*json.RawMessage
		err = json.Unmarshal(*trace["receipt"], &receipt)
		if err != nil || receipt["global_sequence"] == nil {
			continue
		}
		var n uint64
		err = json.Unmarshal(*receipt["global_sequence"], &n)
		if err != nil {
			continue
		}
		if n == actionSeq {
			bytes, err := json.Marshal(trace)
			if err != nil {
				return nil, errors.New("Failed to parse ES response")
			}
			return bytes, nil
		}
	}
	return nil, errors.New("Action trace not found")
}


func getActions(client *elastic.Client, params GetActionsParams) (*GetActionsResult, error) {
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
	result.Actions = make([]Action, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		if hit.Source == nil {
			continue
		}

		var actionTrace ActionTrace
		err = json.Unmarshal(*hit.Source, &actionTrace)
		if err != nil {
			continue
		}
		trace, err := getActionTrace(client, actionTrace.TrxId, actionTrace.Receipt.GlobalSequence)
		if err != nil {
			continue
		}
		action := Action { GlobalActionSeq: actionTrace.Receipt.GlobalSequence,
			BlockNum: actionTrace.BlockNum, BlockTime: actionTrace.BlockTime,
			ActionTrace: trace }
		result.Actions = append(result.Actions, action)
	}
	return result, nil
}


func getTransaction(client *elastic.Client, params GetTransactionParams) (*GetTransactionResult, error) {
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
		return nil, errors.New("Failed to parse ES response")
	}
	var txTraceSource map[string]*json.RawMessage
	err = json.Unmarshal(*docTxTrace.Source, &txTraceSource)
	if err != nil {
		return nil, errors.New("Failed to parse ES response")
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
		return nil, errors.New("Failed to parse ES response")
	}
	result.Trx["trx"] = byteTrx
	result.Trx["receipt"] = *txTraceSource["receipt"]
	return result, nil
}


func getKeyAccounts(client *elastic.Client, params GetKeyAccountsParams) (*GetKeyAccountsResult, error) {
	query := elastic.NewBoolQuery()
	query = query.Filter(elastic.NewMatchQuery("pub_keys.key", params.PublicKey))
	searchResult, err := client.Search().
		Index(AccountsIndex).
		Query(query).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	result := new(GetKeyAccountsResult)
	result.AccountNames = make([]json.RawMessage, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		if hit.Source == nil {
			continue
		}
		var objmap map[string]*json.RawMessage
		err := json.Unmarshal(*hit.Source, &objmap)
		if err != nil {
			return nil, errors.New("Failed to parse ES response")
		}
		result.AccountNames = append(result.AccountNames, *objmap["name"])
	}
	return result, nil
}


func getControlledAccounts(client *elastic.Client, params GetControlledAccountsParams) (*GetControlledAccountsResult, error) {
	query := elastic.NewBoolQuery()
	query = query.Filter(elastic.NewMatchQuery("name.keyword", params.ControllingAccount)) //Is it better to convert name to number and search by id?
	searchResult, err := client.Search().
		Index(AccountsIndex).
		Query(query).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	result := new(GetControlledAccountsResult)
	result.ControlledAccounts = make([]json.RawMessage, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		if hit.Source == nil {
			continue
		}
		var account Account
		err := json.Unmarshal(*hit.Source, &account)
		if err != nil {
			return nil, errors.New("Failed to parse ES response")
		}
		for _, acc := range account.AccountControls {
			result.ControlledAccounts = append(result.ControlledAccounts, acc.Name)
		}
	}
	return result, nil
}