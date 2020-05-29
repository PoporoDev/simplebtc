// To run test just : go test
package main

import (
	"sort"
	"testing"
)

func TestSortTest(t *testing.T) {
	history := RetHistory{}
	item0 := HistoryTxItem{}
	item0.Txid = "0"
	item0.Txpos_in_block = 99
	history.Transactions = append(history.Transactions, item0)
	item1 := HistoryTxItem{}
	item1.Txid = "1"
	item1.Txpos_in_block = 90
	history.Transactions = append(history.Transactions, item1)
	item2 := HistoryTxItem{}
	item2.Txid = "2"
	item2.Txpos_in_block = 105
	history.Transactions = append(history.Transactions, item2)

	txs := make([]string, 0)
	indexs := make([]int, 0)
	blockHeight := 0
	for i, v := range history.Transactions {
		if v.Height == blockHeight {
			txs = append(txs, v.Txid)
			indexs = append(indexs, i)
		}
	}
	sort.Slice(txs, func(i, j int) bool {
		return history.Transactions[indexs[i]].Txpos_in_block < history.Transactions[indexs[j]].Txpos_in_block
	})
	if txs[0] != item1.Txid {
		t.Error("sort fail")
	}
	if txs[1] != item0.Txid {
		t.Error("sort fail")
	}
	if txs[2] != item2.Txid {
		t.Error("sort fail")
	}
}

func TestGetBid(t *testing.T) {
	// get from bitcoin: curl -H 'Content-Type: application/json' -X POST http://user:pwd@127.0.0.1:8332 -s -d '{"jsonrpc":"1.0","id":"","method":"getbid","params":[622073,"3CHWBNWzh8RhgHFHFtAo6ZzyDD9aBAABcN",10]}'
	target := `{"result":{"hashes":["00000000000000000004f1162c6f27ade3403ff0210a8f91d15603f6e0e4c0b9","0000000000000000000f999b30b63e08cdc4ffe7a6305957b1785bf3a0e3c2b9","0000000000000000000e765dae1746f62123569af07345f504648c60e4c5a7fb","0000000000000000000a7d6dcde5ebc731d70e2c9a4333579f488941e819e015","0000000000000000000b994236840452e5a40aeb9971944f1beb9224d1c4b5de","0000000000000000000eee65c404d8176364a3df657f9d803a50c1713a21a5ab","0000000000000000000e489cd65e2dd31e942ba728fe1bd10086f68e5c84a352","00000000000000000002ecf54d4826d350803ee52566ad57caeba9eccdb0a974","00000000000000000009d1accfb98ef44b034d3901b0d51b9c58c860fe82e9da","0000000000000000000cf8ef6daa3f3ff1775eb836d40f765ec34eccf702161d"],"bids":[{"amount":51000,"hashs":["cc35ca532a9a4ab37cc087aa5fd7898a116636186b2d33e1738fc38294114e53","42ce5933fdb3d333282d02113b970f87124a8d522e1edaca278c54eb3dd6c240","00000000000000000000000000000000001e7d5eba24b77143d87c427db87bd2"]}],"time":1584531610},"error":null,"id":""}`
	// get from simple-btc-poporo: curl -H 'Content-Type: application/json' -X POST http://127.0.0.1:8080 -s -d '{"jsonrpc":"1.0","id":"curltext","method":"getbid","params":[622073,"3CHWBNWzh8RhgHFHFtAo6ZzyDD9aBAABcN",10]}'
	getbid := `{"result":{"hashes":["00000000000000000004f1162c6f27ade3403ff0210a8f91d15603f6e0e4c0b9","0000000000000000000f999b30b63e08cdc4ffe7a6305957b1785bf3a0e3c2b9","0000000000000000000e765dae1746f62123569af07345f504648c60e4c5a7fb","0000000000000000000a7d6dcde5ebc731d70e2c9a4333579f488941e819e015","0000000000000000000b994236840452e5a40aeb9971944f1beb9224d1c4b5de","0000000000000000000eee65c404d8176364a3df657f9d803a50c1713a21a5ab","0000000000000000000e489cd65e2dd31e942ba728fe1bd10086f68e5c84a352","00000000000000000002ecf54d4826d350803ee52566ad57caeba9eccdb0a974","00000000000000000009d1accfb98ef44b034d3901b0d51b9c58c860fe82e9da","0000000000000000000cf8ef6daa3f3ff1775eb836d40f765ec34eccf702161d"],"bids":[{"amount":51000,"hashs":["cc35ca532a9a4ab37cc087aa5fd7898a116636186b2d33e1738fc38294114e53","42ce5933fdb3d333282d02113b970f87124a8d522e1edaca278c54eb3dd6c240","00000000000000000000000000000000001e7d5eba24b77143d87c427db87bd2"]}],"time":1584531610},"error":null,"id":""}`
}
