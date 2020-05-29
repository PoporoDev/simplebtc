//go run simple-btc-poporo.go -tfile C:/Users/Scott/Downloads/blockchain_headers -height 622073
//go run simple-btc-poporo.go -httpaddr ":8080"
//test call
//curl -H 'Content-Type: application/json' -X POST http://127.0.0.1:8080 -s -d '{"jsonrpc":"1.0","id":"curltext","method":"getbid","params":[622073,"3CHWBNWzh8RhgHFHFtAo6ZzyDD9aBAABcN",10]}'
//curl -H 'Content-Type: application/json' -X POST http://127.0.0.1:8080 -s -d '{"jsonrpc":"1.0","id":"curltext","method":"bid","params":["3CHWBNWzh8RhgHFHFtAo6ZzyDD9aBAABcN","0.0005", "2016e20fd08125f95021f2c61aa42ea9d56ae9eba133cffbcb2792c0ff2cd288","418015bb9ae982a1975da7d79277c2705727a56894ba0fb246adaabb1f4632e3", "000000000000000000000000000000000075c8326b32645e5ae454ca0fcbbd6c"]}'

package main

import (
	"flag"
	"fmt"

	"io"
	"io/ioutil"
	"log"
	"os"

	//"path"
	//"regexp"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"

	"strconv"
	"strings"

	//"crypto/sha256"
	"net/http"
	//"net/url"
	//"time"
	"errors"
)

//---------------------------------app init , parameters
var (
	h          bool
	configfile string
)

const HEADER_SIZE = 80

func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&configfile, "configfile", "./config-simple-btc.json", "config file path")
}

func usage() {
	fmt.Fprintf(os.Stderr, `app version: app/1.10.0
Usage:  [-h] 

Options:
`)
	flag.PrintDefaults()
}

var config Configuration

const (
	OP_0         = 0x00
	OP_DATA_1    = 0x01 // 1
	OP_PUSHDATA1 = 0x4c //76
	OP_PUSHDATA2 = 0x4d //77
	OP_PUSHDATA4 = 0x4e //78
	OP_1NEGATE   = 0x4f // 79
	OP_1         = 0x51 //81
	OP_RETURN    = 0x6a //106
)

func GetScriptOp(scriptPubKey string, pc uint) (pcRet, opcode uint, vchRet string) {
	pcRet = pc + 2
	if pcRet > uint(len(scriptPubKey)) {
		return
	}
	opcodeRet, err := strconv.ParseUint(scriptPubKey[pc:pcRet], 16, 32)
	if err != nil {
		fmt.Println("GetScriptOp", err)
		return
	}
	opcode = uint(opcodeRet)
	//fmt.Println("opcode", opcode)
	if opcode <= OP_PUSHDATA4 {
		var nSize uint = 0
		if opcode < OP_PUSHDATA1 {
			nSize = opcode
		} //something else
		vchRet = scriptPubKey[pcRet : pcRet+nSize*2]
		pcRet = pcRet + nSize*2
	}
	return
}

func Getbid(params []interface{}) (*RetGetBidData, error) {
	log.Println("call getbid", params)
	//fmt.Printf("%T %T %T\n", params[0], params[1], params[2])
	fblockHeight, ok0 := (params[0]).(float64)
	bidaddress, ok1 := (params[1]).(string)
	fcount, ok2 := (params[2]).(float64)
	if !ok0 || !ok1 || !ok2 {
		return nil, errors.New("parameter error")
	}
	blockHeight := int(fblockHeight)
	count := int(fcount)

	fHeader, _ := os.Open(config.ElectrumHeaderPath)
	defer fHeader.Close()

	var ret RetGetBidData
	hashs, err := GetBlockHashs(blockHeight, count, fHeader)
	if err != nil {
		return nil, err
	}
	ret.Result.Hashes = hashs
	//
	txids, err := GetBlockTxs(config, blockHeight)
	if err != nil {
		return nil, err
	}
	for _, txid := range txids {
		tx, err := GetTx(config, txid)
		//fmt.Println(tx)
		if err != nil {
			return nil, err
		}
		biddata := BidData{}
		for _, v := range tx.Outputs {
			if v.Address == bidaddress { //TODO: check if begin 0x
				biddata.Amount = v.Value
			}
			if v.Type_ == 2 && len(v.ScriptPubKey) == 166 { // TODO: script should convert to []rune
				var pc, opcode uint = 0, 0
				var hexRet string
				pc, opcode, hexRet = GetScriptOp(v.ScriptPubKey, pc)
				if opcode != OP_RETURN {
					return nil, errors.New("first opcode error")
				}
				for opcode != 0 {
					if len(hexRet) > 0 {
						if len(hexRet) < 32 {
							hexRet = hexRet + "0000000000000000000000000000000000"
						}
						//hexRet = string(Reverse([]byte(hexRet)))
						//fmt.Println(hexRet)
						byteHexRet := []byte(hexRet)
						newbytes := make([]byte, len(byteHexRet))
						for i, j := 0, len(byteHexRet)-2; i <= j; i, j = i+2, j-2 {
							newbytes[i], newbytes[i+1], newbytes[j], newbytes[j+1] =
								byteHexRet[j], byteHexRet[j+1], byteHexRet[i], byteHexRet[i+1]
						}
						hexRet = string(newbytes)
						//fmt.Println(hexRet)
						biddata.Hashs = append(biddata.Hashs, hexRet)
					}
					pc, opcode, hexRet = GetScriptOp(v.ScriptPubKey, pc)
				}
			}
		}

		ret.Result.Bids = append(ret.Result.Bids, biddata)
	}

	header, err := GetBlockHeader(blockHeight, fHeader)
	if err != nil {
		return nil, err
	}
	ret.Result.Time = header.timestamp
	return &ret, nil
}

//copy from scriptbuilder.go
func AddData(src, data []byte) []byte {
	dataLen := len(data)

	// When the data consists of a single number that can be represented
	// by one of the "small integer" opcodes, use that opcode instead of
	// a data push opcode followed by the number.
	if dataLen == 0 || dataLen == 1 && data[0] == 0 {
		src = append(src, OP_0)
		return src
	} else if dataLen == 1 && data[0] <= 16 {
		src = append(src, (OP_1-1)+data[0])
		return src
	} else if dataLen == 1 && data[0] == 0x81 {
		src = append(src, byte(OP_1NEGATE))
		return src
	}

	// Use one of the OP_DATA_# opcodes if the length of the data is small
	// enough so the data push instruction is only a single byte.
	// Otherwise, choose the smallest possible OP_PUSHDATA# opcode that
	// can represent the length of the data.
	if dataLen < OP_PUSHDATA1 {
		src = append(src, byte((OP_DATA_1-1)+dataLen))
	} else if dataLen <= 0xff {
		src = append(src, OP_PUSHDATA1, byte(dataLen))
	} else if dataLen <= 0xffff {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(dataLen))
		src = append(src, OP_PUSHDATA2)
		src = append(src, buf...)
	} else {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(dataLen))
		src = append(src, OP_PUSHDATA4)
		src = append(src, buf...)
	}

	// Append the actual data.
	src = append(src, data...)

	return src
}

func Bid(params []interface{}) (ret *RetBidData, err error) {
	if len(params) != 5 {
		err = errors.New("must need 5 parameter ")
		return
	}
	bidaddress, ok0 := (params[0]).(string)
	amount, ok1 := (params[1]).(string)
	if !ok0 || !ok1 {
		err = errors.New(fmt.Sprintf("parameter error %t %t", ok0, ok1))
		return
	}
	fmt.Println("bid", bidaddress, amount)

	rescript := make([]byte, 0, 200)
	rescript = append(rescript, OP_RETURN)
	for i := 2; i < 5; i++ {
		hash, ok := (params[i]).(string)
		if !ok {
			err = errors.New("hash parameter is not a string")
			return
		}
		byteHash, errHex := HexToBytes(hash)
		if errHex != nil {
			err = errHex
			return
		}
		byteHash = Reverse(byteHash)
		//BetCastToBool
		if i == 4 {
			//fmt.Println(byteHash)
			byteHash = byteHash[0:15]
		}
		//fmt.Println("add", rescript, byteHash)
		rescript = AddData(rescript, byteHash)
	}
	outputs := fmt.Sprintf(`[["%s","%s"],["%s","%s"]]`, bidaddress, amount, hex.EncodeToString(rescript), "0")
	//outputs := fmt.Sprintf(`[["%s","%s"],["%s","%s"]]`, bidaddress, amount, "OP_RETURN "+hex.EncodeToString(rescript), "0")
	txraw, errPTM := ElectrumPayToMany(config, outputs)
	if errPTM != nil {
		err = errPTM
		return
	}
	//TODO: broadcast
	ret = &RetBidData{}
	ret.Result = txraw.Hex
	return
}

//---------------------------------http handler
func root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		io.WriteString(w, "get method\n") //http.ServeFile(w, r, "form.html")
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		//log.Println("r.PostForm", r.PostForm)
		//log.Println("r.Form", r.Form)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//log.Println("r.Body", string(body))

		//values, err := url.ParseQuery(string(body))
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	return
		//}

		//log.Println("indices from body", values.Get("indices"))
		strbody := strings.Replace(string(body), `"id":1`, `"id":"1"`, 1) // to fix 'id bug' for old poporo version
		fmt.Println("bid req:", strbody)

		dec := json.NewDecoder(strings.NewReader(strbody))
		// read open bracket

		var m BtcRpcRequest
		err = dec.Decode(&m)
		if err != nil {
			log.Fatal(err)
		}

		//log.Println(m)
		switch m.Method {
		case "getbid":
			ret, err := Getbid(m.Params)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"result":null,"error":{"code":-32603,"message":"%s"},"id":""}`, err.Error()), http.StatusBadRequest)
				return
			}
			ret.Id = m.Id
			e, err := json.Marshal(&ret)
			if err == nil {
				fmt.Fprintf(w, string(e))
			} else {
				http.Error(w, fmt.Sprintf(`{"result":null,"error":{"code":-32603,"message":"%s"},"id":""}`, err.Error()), http.StatusBadRequest)
			}
		case "bid":
			//amount to string
			var objmap map[string]json.RawMessage
			err = json.Unmarshal([]byte(strbody), &objmap)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			arrParam := strings.Split(string(objmap["params"]), ",")
			am := strings.Trim(arrParam[1], " ")
			if am[0] != '"' {
				arrParam[1] = `"` + am + `"`
				replaceparams := strings.Join(arrParam, ",")
				objmap["params"] = []byte(replaceparams)
				newbody, err := json.Marshal(objmap)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				strbody = string(newbody)
				fmt.Println("strbody after process:", strbody)
				dec = json.NewDecoder(strings.NewReader(strbody))
				err = dec.Decode(&m)
				if err != nil {
					log.Fatal(err)
				}
			}

			ret, err := Bid(m.Params)
			if err != nil {
				fmt.Println("Bid error :", err)
				http.Error(w, fmt.Sprintf(`{"result":null,"error":{"code":-32603,"message":"%s"},"id":""}`, err.Error()), http.StatusBadRequest)
				return
			}
			ret.Id = m.Id
			e, err := json.Marshal(&ret)
			if err == nil {
				fmt.Fprintf(w, string(e))
			} else {
				http.Error(w, fmt.Sprintf(`{"result":null,"error":{"code":-32603,"message":"%s"},"id":""}`, err.Error()), http.StatusBadRequest)
			}
		default:
			http.Error(w, "404 method not found.", http.StatusNotFound)
		}

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

//---------------------------------

func main() {
	flag.Parse()
	if h {
		flag.Usage()
		return
	}

	file, _ := os.Open(configfile)
	//defer file.Close()

	decoder := json.NewDecoder(file)
	err := decoder.Decode(&config)
	file.Close()
	if err != nil {
		fmt.Errorf("error:%v", err)
		return
	}

	if config.HttpListenAddr != "" {
		http.HandleFunc("/", root)
		log.Fatal(http.ListenAndServe(config.HttpListenAddr, nil))
	}
}
