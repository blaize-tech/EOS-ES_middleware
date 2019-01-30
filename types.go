package main

import (
	"encoding/json"
)


type ErrorResult struct {
	Code    uint16 `json:"code"`
	Message string `json:"message"`
}


type GetBlockParams struct {
	BlockNum json.RawMessage `json:"block_num_or_id"`
}

type ChainGetBlockResult struct {
	Transactions []struct {
		Status        json.RawMessage `json:"status"`
		CpuUsageUs    json.RawMessage `json:"cpu_usage_us"`
		NetUsageWords json.RawMessage `json:"net_usage_words"`
		Trx           json.RawMessage `json:"trx"`
	} `json:"transactions"`
}

type TransactionFromBlock struct {
	Id                             string `json:"id,omitempty"`
	Signatures            json.RawMessage `json:"signatures"`
	Compression           json.RawMessage `json:"compression"`
	PackedContextFreeData json.RawMessage `json:"packed_context_free_data"`
	PackedTrx             json.RawMessage `json:"packed_trx"`
}


//get_actions types
type GetActionsParams struct {
	AccountName string `json:"account_name"`
	Pos         *int64 `json:"pos,omitempty"`
	Offset      *int64 `json:"offset,omitempty"`
}

type Action struct {
	GlobalActionSeq json.RawMessage `json:"global_action_seq"`
	BlockNum        json.RawMessage `json:"block_num"`
	BlockTime       json.RawMessage `json:"block_time"`
	ActionTrace     json.RawMessage `json:"action_trace"`
}

type GetActionsResult struct {
	Actions []Action `json:"actions"`
}


//get_transaction types
type GetTransactionParams struct {
	Id           string `json:"id"`
}

type GetTransactionResult struct {
	Id                      string `json:"id"`
	Trx map[string]json.RawMessage `json:"trx"`
	BlockTime      json.RawMessage `json:"block_time"`
	BlockNum       json.RawMessage `json:"block_num"`
	Traces         json.RawMessage `json:"traces"`
}


//get_key_accounts types
type GetKeyAccountsParams struct {
	PublicKey string `json:"public_key"`
}

type GetKeyAccountsResult struct {
	AccountNames []json.RawMessage `json:"account_names"`
}


//get_controlled_accounts types
type GetControlledAccountsParams struct {
	ControllingAccount string `json:"controlling_account"`
}

type GetControlledAccountsResult struct {
	ControlledAccounts []json.RawMessage `json:"controlled_accounts"`
}