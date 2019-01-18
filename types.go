package main


type GetActionsParams struct {
	AccountName string `json:"account_name"`
	Pos *int32 `json:"pos,omitempty"`
	Offset *int32 `json:"offset,omitempty"`
}


type GetTransactionParams struct {
	Id string `json:"id"`
	BlockNumHint *int32 `json:"block_num_hint,omitempty"`
}


type GetKeyAccountsParams struct {
	PublicKey string `json:"public_key"`
}


type GetControlledAccountsParams struct {
	ControllingAccount string `json:"controlling_account"`
}