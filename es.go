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


func getBlock(client *elastic.Client, id string) (*json.RawMessage, error) {
	getResult, err := client.Get().
		Index(BlocksIndex).
		Id(id).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	if !getResult.Found || getResult.Source == nil {
		return nil, errors.New("Block not found")
	}
	return getResult.Source, nil
}


func getPackedTx(block *json.RawMessage, trxId string) json.RawMessage {
	var blockSource map[string]*json.RawMessage
	err := json.Unmarshal(*block, &blockSource)
	if err != nil {
		return json.RawMessage{}
	}

	var transactions []json.RawMessage
	err = json.Unmarshal(*blockSource["transactions"], &transactions)
	if err != nil {
		return json.RawMessage{}
	}
	for _, rawTx := range transactions {
		var tx map[string]*json.RawMessage
		err := json.Unmarshal(rawTx, &tx)
		if err != nil {
			return json.RawMessage{}
		}
		var trx map[string]*json.RawMessage
		err = json.Unmarshal(*tx["trx"], &trx)
		if err != nil {
			return json.RawMessage{}
		}

		var id string
		err = json.Unmarshal(*trx["id"], &id)
		if err != nil {
			return json.RawMessage{}
		}
		if id == trxId {
			data := make(map[string]json.RawMessage)
			data["signatures"] = *trx["signatures"]
			data["compression"] = *trx["compression"]
			data["packed_context_free_data"] = *trx["packed_context_free_data"]
			data["packed_trx"] = *trx["packed_trx"]
			byteTrx, err := json.Marshal(data)
			if err != nil {
				return json.RawMessage{}
			}
			return byteTrx
		}
	}
	return json.RawMessage{}
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

	var blockId string
	err = json.Unmarshal(*txTraceSource["producer_block_id"], &blockId)
	if err != nil {
		return nil, errors.New("Failed to parse ES response")
	}
	block, err := getBlock(client, blockId)
	if err != nil {
		return nil, err
	}
	packed_trx := getPackedTx(block, params.Id)
	

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
	receipt := make(map[string]json.RawMessage)
	err = json.Unmarshal(*txTraceSource["receipt"], &receipt)
	if err != nil {
		return nil, errors.New("Failed to parse ES response")
	}
	receiptTrx := []interface{}{1, packed_trx}
	byteReceiptTrx, err := json.Marshal(receiptTrx)
	if err != nil {
		return nil, errors.New("Failed to parse ES response")
	}
	receipt["trx"] = byteReceiptTrx
	byteReceipt, err := json.Marshal(receipt)
	if err != nil {
		return nil, errors.New("Failed to parse ES response")
	}
	result.Trx["receipt"] = byteReceipt
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
	result.AccountNames = make([]json.RawMessage, 0)
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
	query = query.Filter(elastic.NewMatchQuery("name", params.ControllingAccount)) //Is it better to convert name to number and search by id?
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
			return nil, errors.New("Failed to parse ES response")
		}
		var accounts []json.RawMessage
		err = json.Unmarshal(*objmap["account_controls"], &accounts)
		if err != nil {
			return nil, errors.New("Failed to parse ES response")
		}
		result.AccountNames = append(result.AccountNames, accounts...)
	}
	return result, nil
}