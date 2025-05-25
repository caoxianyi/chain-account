package ethereum

import (
	"chain-account/common/util"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"chain-account/chain"
	"chain-account/config"
	"chain-account/rpc/account"
	"github.com/ethereum/go-ethereum/log"
)

func setup() (chain.IChainAdaptor, error) {
	conf, err := config.NewConfig("../../config.yml")

	if err != nil {
		log.Error("load config failed, error:", err)
		return nil, err
	}
	adaptor, err := NewChainAdaptor(conf)
	if err != nil {
		log.Error("create chain adaptor failed, error:", err)
		return nil, err
	}
	return adaptor, nil
}

const (
	privateKey = "0x4f40b69b64cdc6751e2377578cea8443410d0d54cd0449718a0d8bd964b9656e"
	publicKey  = "0x0396e9b5916d03f2acbab92fd8e20599504e53cb5279c8f7e4f7c9f3f3b782bfef"
	address    = "0x62EccDa8bB2Ae5690E319F3eFde897dEAeD86631"
)

/*
 */
func Test_GetSupportChains(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}
	resp, err := adaptor.GetSupportChains(&account.SupportChainsRequest{
		Chain:   ChainName,
		Network: "mainnet",
	})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}

func Test_ConvertAddress(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	resp, err := adaptor.ConvertAddress(&account.ConvertAddressRequest{
		PublicKey: publicKey,
		Network:   "mainnet",
	})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}

func Test_ValidAddress(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	resp, err := adaptor.ValidAddress(&account.ValidAddressRequest{
		Chain:   ChainName,
		Network: "mainnet",
		Address: address,
	})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}

func Test_GetBlockByNumber(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	resp, err := adaptor.GetBlockByNumber(&account.BlockNumberRequest{
		Chain:  ChainName,
		Height: 8386706,
		ViewTx: true,
	})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)

}

func Test_GetBlockByHash(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	resp, err := adaptor.GetBlockByHash(&account.BlockHashRequest{
		Chain:  ChainName,
		Hash:   "0xc65ac7164218ff90ca8101c360f717c465da10bd63007b071c26dd7dedd41324",
		ViewTx: true,
	})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}

func Test_GetBlockHeaderByNumber(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	_, err = adaptor.GetBlockHeaderByNumber(&account.BlockHeaderNumberRequest{
		Chain:   ChainName,
		Network: "mainnet",
		Height:  8386706,
	})
	if err != nil {
		t.Error(err)
		return
	}
	//jsonp := util.ToPrettyJSON(resp)
	//fmt.Println(jsonp)
	/*
		"height": 8386706,
		"hash": "0x84de0708b134572dae223a6e81b85e48da1762ac3abe74a88966aff5267c9b70",
		"Phash":"0xf232edad1b7d2d1cd06413484651aa74a88e694798d52048f19397b20f42855e"
	*/

	/*
		"height": 8386707,
		"hash": "0x1faa86fa7145deef6f31f85fe9d4ac1da3b90e7cea6265c9d00a994b319be9c4",
		"Phash":"0x84de0708b134572dae223a6e81b85e48da1762ac3abe74a88966aff5267c9b70"
	*/
}

func Test_GetBlockHeaderByHash(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	resp, err := adaptor.GetBlockHeaderByHash(&account.BlockHeaderHashRequest{
		Chain:   ChainName,
		Network: "mainnet",
		Hash:    "0xc65ac7164218ff90ca8101c360f717c465da10bd63007b071c26dd7dedd41324",
	})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}

func Test_GetTxByHash(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	resp, err := adaptor.GetTxByHash(&account.TxHashRequest{
		Chain:   ChainName,
		Network: "mainnet",
		Hash:    "0xa2d3a9cd843eae201d1687fd36437caadd288066a87a32c3f003f8ca4fdb9f4e",
	})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json, resp.Tx.Value)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}

func Test_GetBlockByRange(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	resp, err := adaptor.GetBlockByRange(&account.BlockByRangeRequest{
		Chain:   ChainName,
		Network: "mainnet",
		Start:   "22464166",
		End:     "22464167",
	})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}

func Test_GetFee(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	resp, err := adaptor.GetFee(&account.FeeRequest{})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}

// ---
func Test_GetAccount(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	resp, err := adaptor.GetAccount(&account.AccountRequest{
		Chain:           ChainName,
		Network:         "mainnet",
		Address:         "0xac2d6429172af6086efb423dd9042ae11ccec8af",
		ContractAddress: "0x00",
	})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}

// ---
func Test_SendTx(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	req := &account.SendTxRequest{
		RawTx: "0x02f87283aa36a780832dc6c6836bf71b82ea609462eccda8bb2ae5690e319f3efde897deaed86631872386f26fc1000080c080a02d567a1a35bb4eb0e56c0c911aab9635d937da38b71671c35fa13df8cdcbad89a029f7d503330d7eb448cf813d82ac5ece94066e009c2d11c4a9af7078936a8826",
	}
	resp, err := adaptor.SendTx(req)
	if err != nil {
		t.Error(err)
		return
	}
	jsonp := util.ToPrettyJSON(resp)
	fmt.Println(jsonp)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}

// ---
func Test_GetTxByAddress(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	resp, err := adaptor.GetTxByAddress(&account.TxAddressRequest{
		Chain:    ChainName,
		Network:  "mainnet",
		Address:  "0xf4f341E1D3f58702a396b32eD41065465ac3F6Df",
		Page:     1,
		Pagesize: 10,
	})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
}

func Test_BuildUnSignTransaction(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	txParams := Eip1559DynamicFeeTx{
		ChainId:     "1",
		Nonce:       5,
		FromAddress: "0x35096AD62E57e86032a3Bb35aDaCF2240d55421D",
		ToAddress:   "0xf63948D0c77d161A491CD787403ac4222F4d9E55",
		GasLimit:    21000,
		Gas:         12000,

		MaxFeePerGas:         "3000000000",
		MaxPriorityFeePerGas: "2000000000",

		Amount: "1000000000000000000", // 1 ETH

	}
	jsonData, _ := json.Marshal(txParams)
	base64Tx := base64.StdEncoding.EncodeToString(jsonData)

	item := &account.UnSignTransactionRequest{
		Chain:    "1",
		Network:  "mainnet",
		Base64Tx: base64Tx,
	}
	resp, err := adaptor.BuildUnSignTransaction(item)
	jsonp := util.ToPrettyJSON(resp)
	fmt.Println(jsonp)
	//	0x544d3d35826114d3e7b69b0d31855c503e5876ba2f64b029652e35876bd31542
}

func Test_BuildSignedTransaction(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		return
	}

	txParams := Eip1559DynamicFeeTx{
		ChainId:     "1",
		Nonce:       5,
		FromAddress: "0x35096AD62E57e86032a3Bb35aDaCF2240d55421D",
		ToAddress:   "0xf63948D0c77d161A491CD787403ac4222F4d9E55",
		GasLimit:    21000,
		Gas:         12000,

		MaxFeePerGas:         "3000000000",
		MaxPriorityFeePerGas: "2000000000",

		Amount: "1000000000000000000", // 1 ETH

	}
	jsonData, _ := json.Marshal(txParams)
	base64Tx := base64.StdEncoding.EncodeToString(jsonData)

	req := &account.SignedTransactionRequest{
		Chain:     "1",
		Network:   "mainnet",
		Base64Tx:  base64Tx,
		Signature: "5eb0c7574c660c22a7b18a8d112b01cf173c2a4e0928435f1d0092893cca65307243e73d045421eb4caa9217495d879506efbfdf2cda3cacc68d5f553774653001",
		PublicKey: "0x03fa3af3ad7a5c97e3b6bd8bc6dd5751c4b8ced139a2b3a62716f70560cf4a211e",
	}

	resp, err := adaptor.BuildSignedTransaction(req)

	jsonp := util.ToPrettyJSON(resp)
	fmt.Println(jsonp)
	//	"msg": "0x8cb10ae9780d2962e5abd36507d3ef41ff7cacbb12933bf58ad7aa058deb264d",
	//	"signed_tx": "0x02f8720105847735940084b2d05e0082520894f63948d0c77d161a491cd787403ac4222f4d9e55880de0b6b3a764000080c001a05eb0c7574c660c22a7b18a8d112b01cf173c2a4e0928435f1d0092893cca6530a07243e73d045421eb4caa9217495d879506efbfdf2cda3cacc68d5f5537746530"
}

// 发送人
//"privateKey":"0xfe13c8e55444107c32f50cca04965f9772fe4fd720ffedd30b347e541fe7a97c",
//"publicKey":"0x03fa3af3ad7a5c97e3b6bd8bc6dd5751c4b8ced139a2b3a62716f70560cf4a211e",
//"address":"0x35096AD62E57e86032a3Bb35aDaCF2240d55421D"

// 接收人
// PrivateKey：0ec166cd28eb7ba3ad9aa0692735fb0b65b461ad257fbbcf6bd733459336e1d0
// PublicKey： 043596f1aff5b9c8ffb6e2a0ab693a304c75b2cf18dcc178807680fc1cd968065d9812fc25e7d0be5ba41dd8a2118a93ed763eb658e1ab081a25275640d66d1822
// Address: 0xf63948D0c77d161A491CD787403ac4222F4d9E55
