package main

import (
	"encoding/json"
)


type ActionTrace struct {
	Receipt struct {
		Receiver     json.RawMessage `json:"receiver"`
		ActDigest    json.RawMessage `json:"act_digest"`
		GlobalSequence        uint64 `json:"global_sequence"`
		RecvSequence json.RawMessage `json:"recv_sequence"`
		AuthSequence json.RawMessage `json:"auth_sequence"`
		CodeSequence json.RawMessage `json:"code_sequence"`
		AbiSequence  json.RawMessage `json:"abi_sequence"`
	} `json:"receipt"`
	Act struct {
		Account       json.RawMessage `json:"account"`
		Name          json.RawMessage `json:"name"`
		Authorization json.RawMessage `json:"authorization"`
		Data          json.RawMessage `json:"data"`
		HexData       json.RawMessage `json:"hex_data"`
	} `json:"act"`
	ContextFree      json.RawMessage `json:"context_free"`
	Elapsed          json.RawMessage `json:"elapsed"`
	Console          json.RawMessage `json:"console"`
	TrxId                     string `json:"trx_id"`
	BlockNum         json.RawMessage `json:"block_num"`
	BlockTime        json.RawMessage `json:"block_time"`
	ProducerBlockId  json.RawMessage `json:"producer_block_id"`
	AccountRamDeltas json.RawMessage `json:"account_ram_deltas"`
	Except           json.RawMessage `json:"except"`
}