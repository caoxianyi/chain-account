package ethereum

import (
	"chain-account/common/util"
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
		Height: 2222222,
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

	resp, err := adaptor.GetBlockHeaderByNumber(&account.BlockHeaderNumberRequest{
		Chain:   ChainName,
		Network: "mainnet",
		Height:  2222222,
	})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)
	//assert.Equal(t, common.ReturnCode_SUCCESS, resp.Code)
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
		Address:         "0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97",
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

	resp, err := adaptor.SendTx(&account.SendTxRequest{})
	if err != nil {
		t.Error(err)
		return
	}
	json := util.ToPrettyJSON(resp)
	fmt.Println(json)
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
