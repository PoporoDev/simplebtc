package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func HexToBytes(s string) ([]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func ElectrumHistory(config Configuration) (*RetHistory, error) {
	wParam := fmt.Sprintf(`-w=%s`, config.WatchWalletPath)
	var cmd *exec.Cmd
	if len(config.Network) > 0 {
		cmd = exec.Command("python3", config.RunElectrumPath, wParam, "history", config.Network)
	} else {
		cmd = exec.Command("python3", config.RunElectrumPath, wParam, "history")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, ElectrumWErr{"Command history fail", err}
	}
	//fmt.Println(string(output))
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	ret := &RetHistory{}
	err = decoder.Decode(ret)
	if err != nil {
		return nil, ElectrumWErr{"decode fail:" + string(output), err}
	}
	return ret, nil
}

func ElectrumGetTransaction(config Configuration, txid string) (*RetGetTransaction, error) {
	wParam := fmt.Sprintf(`-w=%s`, config.WatchWalletPath)
	var cmd *exec.Cmd
	if len(config.Network) > 0 {
		cmd = exec.Command("python3", config.RunElectrumPath, wParam, "gettransaction", txid, config.Network)
	} else {
		cmd = exec.Command("python3", config.RunElectrumPath, wParam, "gettransaction", txid)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, ElectrumWErr{"Command gettransaction fail", err}
	}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	ret := &RetGetTransaction{}
	err = decoder.Decode(ret)
	if err != nil {
		return nil, ElectrumWErr{"decode fail", err}
	}
	//fmt.Println(ret)
	return ret, nil
}

func ElectrumDeserialize(config Configuration, hex string) (*RetDeserTx, error) {
	wParam := fmt.Sprintf(`-w=%s`, config.WatchWalletPath)
	var cmd *exec.Cmd
	if len(config.Network) > 0 {
		cmd = exec.Command("python3", config.RunElectrumPath, wParam, "deserialize", hex, config.Network)
	} else {
		cmd = exec.Command("python3", config.RunElectrumPath, wParam, "deserialize", hex)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, ElectrumWErr{"Command deserialize fail", err}
	}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	ret := &RetDeserTx{}
	err = decoder.Decode(&ret)
	if err != nil {
		return nil, ElectrumWErr{"decode fail", err}
	}
	return ret, nil
}

func ElectrumPayToMany(config Configuration, outputs string) (*RetGetTransaction, error) {
	feerate, err := ElectrumGetfeerate(config)

	if n, err := strconv.Atoi(feerate); err == nil {
		if len(feerate) < 8 {
			feerate = "0." + strings.Repeat("0", 8-len(feerate)) + feerate
			//fmt.Println("lucky feerate!")
		} else {
			feerate = fmt.Sprintf("%f", n/100000000)
		}
	} else {
		feerate = "0.00600000"
	}

	fmt.Println("paytomany:", outputs, "\nfeerate:", feerate)
	wParam := fmt.Sprintf(`-w=%s`, config.PayableWalletPath)
	var cmd *exec.Cmd
	if len(config.Network) > 0 {
		cmd = exec.Command("python3", config.RunElectrumPath, wParam, "paytomany", outputs, "-f="+feerate, config.Network)
	} else {
		cmd = exec.Command("python3", config.RunElectrumPath, wParam, "paytomany", outputs, "-f="+feerate)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, ElectrumWErr{"Command paytomany fail：" + string(output), err}
	}
	//fmt.Println(string(output))
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	ret := &RetGetTransaction{}
	err = decoder.Decode(&ret)
	if err != nil {
		return nil, ElectrumWErr{"decode fail" + string(output), err}
	}
	txid, errBC := ElectrumBroadcast(config, ret.Hex)
	fmt.Println("broadcast return:", txid)
	if errBC != nil {
		return nil, errBC
	}
	ret.Hex = txid
	return ret, nil
}

func ElectrumGetfeerate(config Configuration) (string, error) {
	var cmd *exec.Cmd
	if len(config.Network) > 0 {
		cmd = exec.Command("python3", config.RunElectrumPath, "getfeerate", "--fee_method=mempool", config.Network)
	} else {
		cmd = exec.Command("python3", config.RunElectrumPath, "getfeerate", "--fee_method=mempool")
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "600000", ElectrumWErr{"Command getfeerate fail：" + string(output), err}
	}
	return strings.TrimSpace(string(output)), nil
}

func ElectrumBroadcast(config Configuration, txhex string) (string, error) {
	var cmd *exec.Cmd
	if len(config.Network) > 0 {
		cmd = exec.Command("python3", config.RunElectrumPath, "broadcast", txhex, config.Network)
	} else {
		cmd = exec.Command("python3", config.RunElectrumPath, "broadcast", txhex)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", ElectrumWErr{"Command broadcast fail: " + string(output), err}
	}
	return strings.TrimSpace(string(output)), nil
}

func GetTx(config Configuration, txid string) (*RetDeserTx, error) {
	txhex, err := ElectrumGetTransaction(config, txid)
	if err != nil {
		return nil, err
	}
	tx, err := ElectrumDeserialize(config, txhex.Hex)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func GetBlockTxs(config Configuration, blockHeight int) ([]string, error) {
	history, err := ElectrumHistory(config)
	if err != nil {
		return nil, err
	}
	txs := make([]string, 0)
	indexs := make([]int, 0)
	for i, v := range history.Transactions {
		if v.Height == blockHeight {
			txs = append(txs, v.Txid)
			indexs = append(indexs, i)
		}
	}
	//todo: need sort? did sort by electrum
	sort.Slice(txs, func(i, j int) bool {
		return history.Transactions[indexs[i]].Txpos_in_block < history.Transactions[indexs[j]].Txpos_in_block
	})
	//fmt.Println("txs", txs)
	return txs, nil
}

//---------------------------------electrum block header read

func Reverse(bytes []byte) []byte {
	newbytes := make([]byte, len(bytes))
	for i, j := 0, len(bytes)-1; i <= j; i, j = i+1, j-1 {
		newbytes[i], newbytes[j] = bytes[j], bytes[i]
	}
	return newbytes
}

func DeserializeHeader(s []byte) BlockHeader {
	var header BlockHeader
	header.version = binary.LittleEndian.Uint32(s[0:4])
	header.prev_block_hash = hex.EncodeToString(Reverse(s[4:36]))
	header.merkle_root = hex.EncodeToString(Reverse(s[36:68]))
	header.timestamp = uint64(binary.LittleEndian.Uint32(s[68:72]))
	header.bits = binary.LittleEndian.Uint32(s[72:76])
	header.nonce = binary.LittleEndian.Uint32(s[76:80])
	return header
}

func GetBlockHeader(height int, f *os.File) (*BlockHeader, error) {
	f.Seek(int64((height)*HEADER_SIZE), 0)
	s := make([]byte, HEADER_SIZE)
	n1, err := f.Read(s)
	if err != nil || n1 != HEADER_SIZE {
		return nil, ElectrumWErr{"read block header file error", err}
	}

	kheader := DeserializeHeader(s)
	header := &kheader
	header.block_height = uint32(height)

	// todo calc block hash.
	/*
		dst := make([]byte, hex.DecodedLen(len(s)))
		hex.Decode(dst, s)
		sum0 := sha256.Sum256(dst)
		sum := sha256.Sum256(sum0[:])
		fmt.Println(hex.EncodeToString(Reverse(sum[:])))
	*/
	return header, nil
}

func GetBlockHashs(fromHeight, count int, f *os.File) ([]string, error) {
	hashs := make([]string, 0)
	f.Seek(int64((fromHeight+1)*HEADER_SIZE), 0)
	s := make([]byte, HEADER_SIZE)

	for i := 0; i < count; i++ {
		n1, err := f.Read(s)
		if err != nil || n1 != HEADER_SIZE {
			return nil, ElectrumWErr{"read block header file error", err}
		}
		header := DeserializeHeader(s)
		header.block_height = uint32(i + fromHeight + 1)
		hashs = append(hashs, header.prev_block_hash)
	}

	return hashs, nil
}
