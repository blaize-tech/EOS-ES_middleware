package main

import (
	"errors"
	"encoding/json"
	"github.com/olivere/elastic"
	"context"
	"net/http"
	"bufio"
	"regexp"
	"strings"
	"math"
	"strconv"
	"fmt"
)

const AccountsIndex          string = "accounts"
const BlocksIndex            string = "blocks"
const TransactionsIndex      string = "transactions"
const TransactionTracesIndex string = "transaction_traces"
const ActionTracesIndex      string = "action_traces"


//get index list from ES and parse indices from it
//return a map where every prefix from input array is a key
//and a value is vector of corresponding indices
func getIndices(esUrl string, prefixes []string) map[string][]string {
	result := make(map[string][]string)
	resp, err := http.Get(esUrl + "/_cat/indices?v&s=index")
	if err != nil {
		return result
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	for _, prefix := range prefixes {
		r, _ := regexp.Compile("\\s" + prefix + "-(\\d)*\\s")
		for _, line := range lines {
			match := r.FindString(line)
			if len(match) != 0 {
				result[prefix] = append(result[prefix], strings.TrimSpace(match))
			}
		}
	}
	return result
}


func findActionTrace(txTrace *TransactionTrace, actionSeq interface{}) (*TransactionTraceActionTrace, error) {
	actionTraces := txTrace.ActionTraces
	trace := new(TransactionTraceActionTrace)
	for len(actionTraces) > 0 {
		trace = &actionTraces[0]
		var receipt map[string]*json.RawMessage
		err := json.Unmarshal(trace.Receipt, &receipt)
		if err != nil || receipt["global_sequence"] == nil {
			continue
		}
		var seq string
		err = json.Unmarshal(*receipt["global_sequence"], &seq)
		if err != nil {
			continue
		}
		var actionSeqStr string
		if i, ok := actionSeq.(float64); ok { // yeah, JSON numbers are floats, gotcha!
			num := uint64(i)
			actionSeqStr = strconv.FormatUint(num, 10)
		} else if s, ok := actionSeq.(string); ok {
			actionSeqStr = s
		}
		if seq == actionSeqStr {
			return trace, nil
		}
		actionTraces = append(actionTraces[1:len(actionTraces)], trace.InlineTraces...)
	}
	return nil, errors.New("Action trace not found in transaction trace")
}

func getActionTrace(client *elastic.Client, txId string, actionSeq interface{}, indices map[string][]string) (json.RawMessage, error) {
	multiGet := client.MultiGet()
	for _, index := range indices[TransactionTracesIndexPrefix] {
		multiGet.Add(elastic.NewMultiGetItem().Index(index).Id(txId))
	}
	mgetResult, err := multiGet.Do(context.Background())
	if err != nil || mgetResult == nil || mgetResult.Docs == nil {
		return nil, err
	}
	var getResult *elastic.GetResult
	for _, doc := range mgetResult.Docs {
		if doc == nil || doc.Error != nil || !doc.Found {
			continue
		}
		getResult = doc
	}

	if getResult == nil || !getResult.Found || getResult.Source == nil {
		return nil, errors.New("Action trace not found")
	}
	var txTrace TransactionTrace
	err = json.Unmarshal(*getResult.Source, &txTrace)
	if err != nil {
		return nil, errors.New("Failed to parse ES response")
	}

	trace, err := findActionTrace(&txTrace, actionSeq)
	if err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(trace)
	if err != nil {
		return nil, errors.New("Failed to parse ES response")
	}
	return bytes, nil
}


func getActions(client *elastic.Client, params GetActionsParams, indices map[string][]string) (*GetActionsResult, error) {
	ascOrder := true
	if *params.Pos < 0 {
		ascOrder = false
		*params.Pos = int64(math.Abs(float64(*params.Pos))) - 1
		*params.Offset = -*params.Offset
	}
	pos1 := *params.Pos
	pos2 := *params.Pos + *params.Offset
	start := int64(math.Min(float64(pos1), float64(pos2)))
	if start < 0 {
		*params.Offset = int64(math.Abs(float64(*params.Offset - start)))
		start = 0
	}
	if *params.Offset < 0 {
		*params.Offset = int64(math.Abs(float64(*params.Offset)))
	}
	
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMultiMatchQuery(params.AccountName, "receipt.receiver", "act.authorization.actor"))
	msearch := client.MultiSearch()
	for _, index := range indices[ActionTracesIndexPrefix] {
		sreq := elastic.NewSearchRequest().
			Index(index).Query(query).
			Sort("receipt.global_sequence", ascOrder).
			From(int(start)).Size(int(*params.Offset))
		msearch.Add(sreq)
	}
	msearchResult, err := msearch.Do(context.Background())
	if err != nil || msearchResult == nil || msearchResult.Responses == nil {
		return nil, err
	}

	var searchHits []elastic.SearchHit
	for i, _ := range msearchResult.Responses {
		var resp *elastic.SearchResult
		if ascOrder {
			resp = msearchResult.Responses[i]
		} else {
			resp = msearchResult.Responses[len(msearchResult.Responses)-1-i]
		}
		if resp == nil || resp.Error != nil {
			continue
		}
		for _, hit := range resp.Hits.Hits {
			if hit != nil && len(searchHits) < int(*params.Offset) {
				searchHits = append(searchHits, *hit)
			}
		}
		if len(searchHits) == int(*params.Offset) {
			break
		}
	}
	
	result := new(GetActionsResult)
	result.Actions = make([]Action, 0, len(searchHits))
	for _, hit := range searchHits {
		if hit.Source == nil {
			continue
		}

		var actionTrace ActionTrace
		err = json.Unmarshal(*hit.Source, &actionTrace)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		trace, err := getActionTrace(client, actionTrace.TrxId, actionTrace.Receipt.GlobalSequence, indices)
		if err != nil {
			continue
		}
		var globalActionSeq uint64
		if i, ok := actionTrace.Receipt.GlobalSequence.(float64); ok {
			globalActionSeq = uint64(i)
		} else if s, ok := actionTrace.Receipt.GlobalSequence.(string); ok {
			var err error
			globalActionSeq, err = strconv.ParseUint(s, 10, 64)
			if err != nil {
				return nil, err
			}
		}
		action := Action { GlobalActionSeq: globalActionSeq,
			BlockNum: actionTrace.BlockNum, BlockTime: actionTrace.BlockTime,
			ActionTrace: trace }
		result.Actions = append(result.Actions, action)
	}
	return result, nil
}


func getTransaction(client *elastic.Client, params GetTransactionParams, indices map[string][]string) (*GetTransactionResult, error) {
	mgetTx := client.MultiGet()
	mgetTxTrace := client.MultiGet()
	for _, index := range indices[TransactionsIndexPrefix] {
		mgetTx.Add(elastic.NewMultiGetItem().Index(index).Id(params.Id))
	}
	for _, index := range indices[TransactionTracesIndexPrefix] {
		mgetTxTrace.Add(elastic.NewMultiGetItem().Index(index).Id(params.Id))
	}
	mgetTxResult, err := mgetTx.Do(context.Background())
	if err != nil || mgetTxResult == nil || mgetTxResult.Docs == nil {
		return nil, err
	}
	mgetTxTraceResult, err := mgetTxTrace.Do(context.Background())
	if err != nil || mgetTxTraceResult == nil || mgetTxTraceResult.Docs == nil {
		return nil, err
	}

	var getTxResult *elastic.GetResult
	for _, doc := range mgetTxResult.Docs {
		if doc == nil || doc.Error != nil || !doc.Found {
			continue
		}
		getTxResult = doc
	}
	var getTxTraceResult *elastic.GetResult
	for _, doc := range mgetTxTraceResult.Docs {
		if doc == nil || doc.Error != nil || !doc.Found {
			continue
		}
		getTxTraceResult = doc
	}

	if getTxResult == nil || getTxTraceResult == nil || 
		!getTxResult.Found || !getTxTraceResult.Found {
		return nil, errors.New("Transaction not found")
	}
	
	var transaction Transaction
	err = json.Unmarshal(*getTxResult.Source, &transaction)
	if err != nil {
		return nil, errors.New("Failed to parse ES response")
	}
	var txTrace TransactionTrace
	err = json.Unmarshal(*getTxTraceResult.Source, &txTrace)
	if err != nil {
		return nil, errors.New("Failed to parse ES response")
	}

	result := new(GetTransactionResult)
	result.Id = params.Id
	result.Trx = make(map[string]json.RawMessage)
	result.BlockTime = txTrace.BlockTime
	result.BlockNum = transaction.BlockNum
	result.Traces, err = json.Marshal(txTrace.ActionTraces)
	if err != nil {
		return nil, errors.New("Internal error")
	}
	trx := make(map[string]json.RawMessage)
	trx["expiration"] = transaction.Expiration
	trx["ref_block_num"] = transaction.RefBlockNum
	trx["ref_block_prefix"] = transaction.RefBlockPrefix
	trx["max_net_usage_words"] = transaction.MaxNetUsageWords
	trx["max_cpu_usage_ms"] = transaction.MaxCpuUsageMs
	trx["delay_sec"] = transaction.DelaySec
	trx["context_free_actions"] = transaction.ContextFreeActions
	trx["actions"] = transaction.Actions
	trx["transaction_extensions"] = transaction.TransactionExtensions
	trx["signatures"] = transaction.Signatures
	trx["context_free_data"] = transaction.ContextFreeData
	byteTrx, err := json.Marshal(trx)
	if err != nil {
		return nil, errors.New("Internal error")
	}
	result.Trx["trx"] = byteTrx
	result.Trx["receipt"] = txTrace.Receipt
	return result, nil
}


func getKeyAccounts(client *elastic.Client, params GetKeyAccountsParams, indices map[string][]string) (*GetKeyAccountsResult, error) {
	query := elastic.NewBoolQuery()
	query = query.Filter(elastic.NewMatchQuery("pub_keys.key", params.PublicKey))
	msearch := client.MultiSearch()
	for _, index := range indices[AccountsIndexPrefix] {
		msearch.Add(elastic.NewSearchRequest().Index(index).Query(query))
	}
	msearchResult, err := msearch.Do(context.Background())
	if err != nil || msearchResult == nil || msearchResult.Responses == nil {
		return nil, err
	}
	var searchHits []elastic.SearchHit
	for _, resp := range msearchResult.Responses {
		if resp == nil || resp.Error != nil {
			continue
		}
		for _, hit := range resp.Hits.Hits {
			if hit != nil {
				searchHits = append(searchHits, *hit)
			}
		}
	}

	result := new(GetKeyAccountsResult)
	result.AccountNames = make([]json.RawMessage, 0, len(searchHits))
	for _, hit := range searchHits {
		if hit.Source == nil {
			continue
		}
		var account Account
		err := json.Unmarshal(*hit.Source, &account)
		if err != nil {
			return nil, errors.New("Failed to parse ES response")
		}
		result.AccountNames = append(result.AccountNames, account.Name)
	}
	return result, nil
}


func getControlledAccounts(client *elastic.Client, params GetControlledAccountsParams, indices map[string][]string) (*GetControlledAccountsResult, error) {
	query := elastic.NewBoolQuery()
	query = query.Filter(elastic.NewMatchQuery("name.keyword", params.ControllingAccount)) //Is it better to convert name to number and search by id?
	msearch := client.MultiSearch()
	for _, index := range indices[AccountsIndexPrefix] {
		msearch.Add(elastic.NewSearchRequest().Index(index).Query(query))
	}
	msearchResult, err := msearch.Do(context.Background())
	if err != nil || msearchResult == nil || msearchResult.Responses == nil {
		return nil, err
	}
	var searchHits []elastic.SearchHit
	for _, resp := range msearchResult.Responses {
		if resp == nil || resp.Error != nil {
			continue
		}
		for _, hit := range resp.Hits.Hits {
			if hit != nil {
				searchHits = append(searchHits, *hit)
			}
		}
	}

	result := new(GetControlledAccountsResult)
	result.ControlledAccounts = make([]json.RawMessage, 0, len(searchHits))
	for _, hit := range searchHits {
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