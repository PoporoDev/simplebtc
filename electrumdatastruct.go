// myconfig
package main

import (
	"fmt"
	"time"
)

//history
type HistorySummary struct {
	End_balance   string `json:"end_balance,omitempty"`
	End_date      string `json:"end_date,omitempty"`
	From_height   int    `json:"from_height,omitempty"`
	Incoming      string `json:"incoming,omitempty"`
	Outgoing      string `json:"outgoing,omitempty"`
	Start_balance string `json:"start_balance,omitempty"`
	Start_date    string `json:"start_date,omitempty"`
	To_height     int    `json:"to_height,omitempty"`
}
type HistoryTxItem struct {
	Balance        string `json:"balance,omitempty"`
	Confirmations  int    `json:"confirmations,omitempty"`
	Date           string `json:"date,omitempty"`
	Height         int    `json:"height,omitempty"`
	Incoming       bool   `json:"incoming,omitempty"`
	Label          string `json:"label,omitempty"`
	Timestamp      uint   `json:"timestamp,omitempty"`
	Txid           string `json:"txid,omitempty"`
	Txpos_in_block int    `json:"txpos_in_block,omitempty"`
	Value          string `json:"value,omitempty"`
}
type RetHistory struct {
	Summary      HistorySummary  `json:"summary,omitempty"`
	Transactions []HistoryTxItem `json:"transactions,omitempty"`
}

//gettransaction
type RetGetTransaction struct {
	Complete bool   `json:"complete,omitempty"`
	Final    bool   `json:"final,omitempty"`
	Hex      string `json:"hex,omitempty"`
}

//Deserialize tx
type TxInput struct {
	Address      string   `json:"address,omitempty"`
	Num_sig      int      `json:"num_sig,omitempty"`
	Prevout_hash string   `json:"prevout_hash,omitempty"`
	Prevout_n    int      `json:"prevout_n,omitempty"`
	Pubkeys      []string `json:"pubkeys,omitempty"`
	ScriptSig    string   `json:"scriptSig,omitempty"`
	Sequence     uint     `json:"sequence,omitempty"`
	//Signatures string   `json:"signatures,omitempty"`
	Type_ string `json:"type,omitempty"`
	//X_pubkeys []string `json:"x_pubkeys,omitempty"`
}

type TxOutput struct {
	Address      string `json:"address, omitempty"`
	Prevout_n    int    `json:"prevout_n, omitempty"`
	ScriptPubKey string `json:"scriptPubKey, omitempty"`
	Type_        int    `json:"type, omitempty"`
	Value        uint64 `json:"value, omitempty"`
}

type RetDeserTx struct {
	Inputs     []TxInput  `json:"inputs, omitempty"`
	LockTime   int        `json:"lockTime, omitempty"`
	Outputs    []TxOutput `json:"outputs, omitempty"`
	Partial    bool       `json:"partial, omitempty"`
	Segwit_ser bool       `json:"segwit_ser, omitempty"`
	Version    int        `json:"version, omitempty"`
}

// file path that caused it.
type ElectrumWErr struct {
	Desc string
	Err  error // Returned by the system call.
}

func (e ElectrumWErr) Error() string {
	return e.Desc + ": " + e.Err.Error()
}

//block header
type BlockHeader struct {
	version         uint32
	prev_block_hash string
	merkle_root     string
	timestamp       uint64
	bits            uint32
	nonce           uint32
	block_height    uint32
}

func (b BlockHeader) String() string {
	return fmt.Sprintf(
		`{
  "version" : %v %x,
  "prev_block_hash" : "%v",
  "merkle_root" : "%v",
  "timestamp" : %v %v,
  "bits" : %v %x,
  "nonce" : %v,
  "block_height" : %v
}`, b.version, b.version,
		b.prev_block_hash,
		b.merkle_root,
		b.timestamp, time.Unix(int64(b.timestamp), 0),
		b.bits, b.bits,
		b.nonce,
		b.block_height)
}

//---------------------------------btc rpc interface
//{"jsonrpc":"1.0","id":"curltext","method":"getblockchaininfo","params":[]}
//
type BtcRpcRequest struct {
	Jsonrpc string        `json:"jsonrpc,omitempty"`
	Id      string        `json:"id,omitempty"`
	Method  string        `json:"method,omitempty"`
	Params  []interface{} `json:"params,omitempty"`
}

func (o BtcRpcRequest) String() string {
	return fmt.Sprintf(`{"jsonrpc":"%v","id":"%v","method":"%v","params":%v}`,
		o.Jsonrpc,
		o.Id,
		o.Method,
		o.Params)
}

//
type BtcError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
type BidData struct {
	Amount uint64   `json:"amount,omitempty"`
	Hashs  []string `json:"hashs,omitempty"`
}

type BidResult struct {
	Hashes []string  `json:"hashes,omitempty"`
	Bids   []BidData `json:"bids,omitempty"`
	Time   uint64    `json:"time,omitempty"`
}

type RetGetBidData struct {
	Result BidResult `json:"result,omitempty"`
	Error  *BtcError `json:"error"`
	Id     string    `json:"id,omitempty"`
}

type RetBidData struct {
	Result string    `json:"result,omitempty"`
	Error  *BtcError `json:"error"`
	Id     string    `json:"id,omitempty"`
}
