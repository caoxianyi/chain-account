package ethereum

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"testing"
)

func Test_CreateEip1559UnSignTx(t *testing.T) {
	toAddress := common.HexToAddress("0xf63948D0c77d161A491CD787403ac4222F4d9E55")
	txData := &types.DynamicFeeTx{
		ChainID:   big.NewInt(1),
		Nonce:     5,
		GasTipCap: big.NewInt(2000000000),
		GasFeeCap: big.NewInt(3000000000),
		Gas:       21000,
		To:        &toAddress,
		Value:     big.NewInt(1000000000000000000),
	}

	rawTx, err := CreateEip1559UnSignTx(txData, big.NewInt(1))
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("resp", rawTx)
	// 0x544d3d35826114d3e7b69b0d31855c503e5876ba2f64b029652e35876bd31542
}

func Test_CreateEip1559SignedTx(t *testing.T) {
	toAddress := common.HexToAddress("0xf63948D0c77d161A491CD787403ac4222F4d9E55")
	txData := &types.DynamicFeeTx{
		ChainID:   big.NewInt(1),
		Nonce:     5,
		GasTipCap: big.NewInt(2000000000),
		GasFeeCap: big.NewInt(3000000000),
		Gas:       21000,
		To:        &toAddress,
		Value:     big.NewInt(1000000000000000000),
	}
	signature := "04833b4a3a69c4bbf7ac77f64f6f5564f55c9ed6d4c78a7158d5bbdb1156abde37f2228df2f2d43f7d0eb439040e6d857d716cfda3ed7f5727a164215fae8c7e00"
	signatureByte, _ := hex.DecodeString(signature)
	fmt.Println("signatureByte", signatureByte)
	// [4 131 59 74 58 105 196 187 247 172 119 246 79 111 85 100 245 92 158 214 212 199 138 113 88 213 187 219 17 86 171 222 55 242 34 141 242 242 212 63 125 14 180 57 4 14 109 133 125 113 108 253 163 237 127 87 39 161 100 33 95 174 140 126 0]
	signer, signedTx, rawTx, txHash, err := CreateEip1559SignedTx(txData, signatureByte, txData.ChainID)
	if err != nil {
		t.Error(err)
		return
	}
	sender, err := types.Sender(signer, signedTx)
	fmt.Println("sender", sender)
	// 0x82565b64e8063674CAea7003979280f4dbC3aAE7

	fmt.Println(rawTx)
	fmt.Println(txHash)
	//json := util.ToPrettyJSON(resp)
	//fmt.Println(json)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}
