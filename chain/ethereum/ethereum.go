package ethereum

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	account2 "github.com/dapplink-labs/chain-explorer-api/common/account"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/shopspring/decimal"
	"github.com/status-im/keycard-go/hexutils"

	"chain-account/chain"
	"chain-account/common/global_const"
	"chain-account/common/util"
	"chain-account/config"
	"chain-account/rpc/account"
)

const ChainName = "Ethereum"

type ChainAdaptor struct {
	EthClient IEth
	EthData   *EthData
}

func NewChainAdaptor(con *config.Config) (chain.IChainAdaptor, error) {
	rpcUrl := con.WalletNode.Eth.RpcUrl
	dataApiUrl := con.WalletNode.Eth.DataApiUrl
	dataApiKey := con.WalletNode.Eth.DataApiKey

	ethClient, err := NewEthClient(context.Background(), rpcUrl)
	if err != nil {
		return nil, err
	}

	ethData, err2 := NewEthData(dataApiUrl, dataApiKey, time.Second*35)
	if err2 != nil {
		return nil, err2
	}
	return &ChainAdaptor{
		EthClient: ethClient,
		EthData:   ethData,
	}, nil
}

// 验证 是否满足当前节点
func (c *ChainAdaptor) GetSupportChains(req *account.SupportChainsRequest) (*account.SupportChainsResponse, error) {
	fmt.Println("进这里 2")
	return &account.SupportChainsResponse{
		Code:    global_const.ReturnCode_SUCCESS,
		Msg:     "Support this chain",
		Support: true,
	}, nil
}

// 传入公钥 转换成地址
func (c *ChainAdaptor) ConvertAddress(req *account.ConvertAddressRequest) (*account.ConvertAddressResponse, error) {
	// 1. 处理 0x 前缀
	publicKeyStr := strings.TrimPrefix(req.PublicKey, "0x")

	publicKeyBytes, err := hex.DecodeString(publicKeyStr)
	if err != nil {
		log.Error("decode public key failed:", err)
		return &account.ConvertAddressResponse{
			Code:    global_const.ReturnCode_ERROR,
			Msg:     "convert address fail",
			Address: common.Address{}.String(),
		}, nil
	}

	// 2. 解压公钥
	pubKey, err := decompressPublicKey(publicKeyBytes)
	if err != nil {
		return &account.ConvertAddressResponse{
			Code:    global_const.ReturnCode_ERROR,
			Msg:     "decompress publicKey fail",
			Address: common.Address{}.String(),
		}, nil
	}
	// 3. 计算地址
	fullPubKey := pubKey.SerializeUncompressed()
	addressCommon := common.BytesToAddress(crypto.Keccak256(fullPubKey[1:])[12:]).String()
	return &account.ConvertAddressResponse{
		Code:    global_const.ReturnCode_SUCCESS,
		Msg:     "convert address success",
		Address: addressCommon,
	}, nil
}

// 地址格式验证
func (c *ChainAdaptor) ValidAddress(req *account.ValidAddressRequest) (*account.ValidAddressResponse, error) {
	if len(req.Address) != 42 || !strings.HasPrefix(req.Address, "0x") {
		return &account.ValidAddressResponse{
			Code:  global_const.ReturnCode_SUCCESS,
			Msg:   "invalid address",
			Valid: false,
		}, nil
	}
	ok := regexp.MustCompile("^[0-9a-fA-F]{40}$").MatchString(req.Address[2:])
	if ok {
		return &account.ValidAddressResponse{
			Code:  global_const.ReturnCode_SUCCESS,
			Msg:   "valid address",
			Valid: true,
		}, nil
	} else {
		return &account.ValidAddressResponse{
			Code:  global_const.ReturnCode_SUCCESS,
			Msg:   "invalid address",
			Valid: false,
		}, nil
	}
}

// 通过区块号获取区块数据
func (c *ChainAdaptor) GetBlockByNumber(req *account.BlockNumberRequest) (*account.BlockResponse, error) {
	block, err := c.EthClient.BlockByNumber(big.NewInt(req.Height))
	if err != nil {
		log.Error("get block by number failed:", err)
		return &account.BlockResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get block by number fail",
		}, nil
	}

	height, _ := block.NumberUint64()
	var blockTxList []*account.BlockInfoTransactionList
	for _, tx := range block.Transactions {
		itemTx := &account.BlockInfoTransactionList{
			From:   tx.From,
			To:     tx.To,
			Hash:   tx.Hash,
			Amount: tx.Value,
			Height: height,
		}
		blockTxList = append(blockTxList, itemTx)
	}
	return &account.BlockResponse{
		Code:         global_const.ReturnCode_SUCCESS,
		Msg:          "get block by number success",
		Height:       req.Height,
		Hash:         block.Hash.String(),
		BaseFee:      block.BaseFee,
		Transactions: blockTxList,
	}, nil
}

// 通过区块Hash获取区块数据
func (c *ChainAdaptor) GetBlockByHash(req *account.BlockHashRequest) (*account.BlockResponse, error) {
	block, err := c.EthClient.BlockByHash(common.HexToHash(req.Hash))
	if err != nil {
		log.Error("get block by hash failed:", err)
		return &account.BlockResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get block by hash fail",
		}, nil
	}

	height, _ := block.NumberUint64()
	var blockTxList []*account.BlockInfoTransactionList
	for _, tx := range block.Transactions {
		itemTx := &account.BlockInfoTransactionList{
			From:   tx.From,
			To:     tx.To,
			Hash:   tx.Hash,
			Amount: tx.Value,
			Height: height,
		}
		blockTxList = append(blockTxList, itemTx)
	}
	return &account.BlockResponse{
		Code:         global_const.ReturnCode_SUCCESS,
		Msg:          "get block by hash success",
		Height:       int64(height),
		Hash:         block.Hash.String(),
		BaseFee:      block.BaseFee,
		Transactions: blockTxList,
	}, nil

}

// 通过区块号获取区块头信息
func (c *ChainAdaptor) GetBlockHeaderByNumber(req *account.BlockHeaderNumberRequest) (*account.BlockHeaderResponse, error) {
	var blockNumber *big.Int
	if req.Height == 0 {
		blockNumber = nil
	} else {
		blockNumber = big.NewInt(req.Height)
	}
	blockInfo, err := c.EthClient.BlockHeaderByNumber(blockNumber)
	if err != nil {
		log.Error("get block header by number failed:", err)
		return &account.BlockHeaderResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get block header by number hash fail",
		}, nil
	}
	blockHead := &account.BlockHeader{
		Hash:             blockInfo.Hash().Hex(),
		ParentHash:       blockInfo.ParentHash.String(),
		UncleHash:        blockInfo.UncleHash.String(),
		CoinBase:         blockInfo.Coinbase.String(),
		Root:             blockInfo.Root.String(),
		TxHash:           blockInfo.TxHash.String(),
		ReceiptHash:      blockInfo.ReceiptHash.String(),
		ParentBeaconRoot: common.Hash{}.String(),
		Difficulty:       blockInfo.Difficulty.String(),
		Number:           blockInfo.Number.String(),
		GasLimit:         blockInfo.GasLimit,
		GasUsed:          blockInfo.GasUsed,
		Time:             blockInfo.Time,
		Extra:            hex.EncodeToString(blockInfo.Extra),
		MixDigest:        blockInfo.MixDigest.String(),
		Nonce:            strconv.FormatUint(blockInfo.Nonce.Uint64(), 10),
		BaseFee:          blockInfo.BaseFee.String(),
		WithdrawalsHash:  common.Hash{}.String(),
		BlobGasUsed:      0,
		ExcessBlobGas:    0,
	}
	return &account.BlockHeaderResponse{
		Code:        global_const.ReturnCode_SUCCESS,
		Msg:         "get block header by number success",
		BlockHeader: blockHead,
	}, nil
}

// 通过区块Hash获取区块头信息
func (c *ChainAdaptor) GetBlockHeaderByHash(req *account.BlockHeaderHashRequest) (*account.BlockHeaderResponse, error) {
	blockInfo, err := c.EthClient.BlockHeaderByHash(common.HexToHash(req.Hash))
	if err != nil {
		log.Error("get block header by hash failed:", err)
		return &account.BlockHeaderResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get block header by hash hash fail",
		}, nil
	}
	blockHead := &account.BlockHeader{
		Hash:             blockInfo.Hash().String(),
		ParentHash:       blockInfo.ParentHash.String(),
		UncleHash:        blockInfo.UncleHash.String(),
		CoinBase:         blockInfo.Coinbase.String(),
		Root:             blockInfo.Root.String(),
		TxHash:           blockInfo.TxHash.String(),
		ReceiptHash:      blockInfo.ReceiptHash.String(),
		ParentBeaconRoot: common.Hash{}.String(),
		Difficulty:       blockInfo.Difficulty.String(),
		Number:           blockInfo.Number.String(),
		GasLimit:         blockInfo.GasLimit,
		GasUsed:          blockInfo.GasUsed,
		Time:             blockInfo.Time,
		Extra:            hex.EncodeToString(blockInfo.Extra),
		MixDigest:        blockInfo.MixDigest.String(),
		Nonce:            strconv.FormatUint(blockInfo.Nonce.Uint64(), 10),
		BaseFee:          blockInfo.BaseFee.String(),
		WithdrawalsHash:  common.Hash{}.String(),
		BlobGasUsed:      0,
		ExcessBlobGas:    0,
	}
	return &account.BlockHeaderResponse{
		Code:        global_const.ReturnCode_SUCCESS,
		Msg:         "get block header by hash success",
		BlockHeader: blockHead,
	}, nil
}

// 获取当前账户的信息
func (c *ChainAdaptor) GetAccount(req *account.AccountRequest) (*account.AccountResponse, error) {
	// 获取交易笔数 nonce
	nonce, err := c.EthClient.TxCountByAddress(common.HexToAddress(req.Address))
	if err != nil {
		log.Error("get nonce by address fail:", err)
		return &account.AccountResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get nonce by address fail",
		}, nil
	}
	//通过合约地址和用户地址查询账户代币余额
	balanceResult, err := c.EthData.GetBalanceByAddress(req.ContractAddress, req.Address)
	if err != nil {
		return &account.AccountResponse{
			Code:    global_const.ReturnCode_ERROR,
			Msg:     "get token balance fail",
			Balance: "0",
		}, err
	}
	log.Info("balance result", "balance：", balanceResult.Balance, "balanceStr：", balanceResult.BalanceStr)

	balanceStr := "0"
	if balanceResult.Balance != nil && balanceResult.Balance.Int() != nil {
		balanceStr = balanceResult.Balance.Int().String()
	}
	sequence := strconv.FormatUint(uint64(nonce), 10)

	return &account.AccountResponse{
		Code:          global_const.ReturnCode_SUCCESS,
		Msg:           "get account response success",
		AccountNumber: "0",
		Sequence:      sequence,
		Balance:       balanceStr,
	}, nil
}

// 获取fee
func (c *ChainAdaptor) GetFee(req *account.FeeRequest) (*account.FeeResponse, error) {
	gasPrice, err := c.EthClient.SuggestGasPrice()
	if err != nil {
		log.Error("get gas price fail:", err)
		return &account.FeeResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get gas price fail",
		}, nil
	}
	gasTipCap, err := c.EthClient.SuggestGasTipCap()
	if err != nil {
		log.Error("get gas price fail:", err)
		return &account.FeeResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get gas price fail",
		}, nil
	}
	return &account.FeeResponse{
		Code:      global_const.ReturnCode_SUCCESS,
		Msg:       "get gas price success",
		SlowFee:   gasPrice.String() + "|" + gasTipCap.String(),
		NormalFee: gasPrice.String() + "|" + gasTipCap.String() + "|" + "*2",
		FastFee:   gasPrice.String() + "|" + gasTipCap.String() + "|" + "*3",
	}, nil
}

// 广播交易
func (c *ChainAdaptor) SendTx(req *account.SendTxRequest) (*account.SendTxResponse, error) {
	transaction, err := c.EthClient.SendRawTransaction(req.RawTx)
	if err != nil {
		log.Error("send tx fail:", err)
		return &account.SendTxResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "send tx fail",
		}, nil
	}
	return &account.SendTxResponse{
		Code:   global_const.ReturnCode_SUCCESS,
		Msg:    "send tx success",
		TxHash: transaction.String(),
	}, nil
}

// 通过地址获取交易记录
func (c *ChainAdaptor) GetTxByAddress(req *account.TxAddressRequest) (*account.TxAddressResponse, error) {
	var resp *account2.TransactionResponse[account2.AccountTxResponse]
	var err error
	if req.ContractAddress != "0x00" && req.ContractAddress != "" {
		resp, err = c.EthData.GetTxByAddress(uint64(req.Page), uint64(req.Pagesize), req.Address, "tokentx")
	} else {
		resp, err = c.EthData.GetTxByAddress(uint64(req.Page), uint64(req.Pagesize), req.Address, "txlist")
	}
	if err != nil {
		log.Error("get GetTxByAddress error", "err", err)
		return &account.TxAddressResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get tx list fail",
			Tx:   nil,
		}, err
	}
	txs := resp.TransactionList
	var txsList []*account.TxMessage
	for i := 0; i < len(txs); i++ {
		txsList = append(txsList, &account.TxMessage{
			Hash:   txs[i].TxId,
			To:     txs[i].To,
			From:   txs[i].From,
			Fee:    txs[i].TxId,
			Status: account.TxStatus_Success,
			Value:  txs[i].Amount,
			Type:   1,
			Height: txs[i].Height,
		})
	}
	return &account.TxAddressResponse{
		Code: global_const.ReturnCode_SUCCESS,
		Msg:  "get tx list success",
		Tx:   txsList,
	}, nil
}

// 按Hash获取交易详情
func (c *ChainAdaptor) GetTxByHash(req *account.TxHashRequest) (*account.TxHashResponse, error) {
	// 按Hash 获取交易详情
	transaction, err := c.EthClient.TxByHash(common.HexToHash(req.Hash))
	if err != nil {
		log.Error("get tx by hash fail:", err)
		return &account.TxHashResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get tx by hash fail",
		}, nil
	}
	if errors.Is(err, ethereum.NotFound) {
		log.Error("Ethereum Tx NotFound:", err)
		return &account.TxHashResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "Ethereum Tx NotFound",
		}, nil
	}

	// 按Hash 获取交易收据
	receipt, err := c.EthClient.TxReceiptByHash(common.HexToHash(req.Hash))
	if err != nil {
		log.Error("get tx receipt by hash fail:", err)
		return &account.TxHashResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get tx receipt by hash fail",
		}, nil
	}

	// 按地址获取合约字节码
	var beforeToAddress string
	var beforeTokenAddress string
	var beforeValue *big.Int
	code, err := c.EthClient.EthGetCode(common.HexToAddress(transaction.To().String()))
	if err != nil {
		log.Error("eth get code fail:", err)
		return &account.TxHashResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "eth get code fail",
		}, nil
	}
	if code == "contract" {
		inputData := hexutil.Encode(transaction.Data()[:])
		/*
			34-74 字符：目标地址（20字节）
			74-138 字符：转账金额（32字节）
		*/
		if len(inputData) >= 138 && inputData[:10] == "0xa9059cbb" {
			beforeToAddress = "0x" + inputData[34:74]
			trimHex := strings.TrimLeft(inputData[74:138], "0")
			rawValue, _ := hexutil.DecodeBig("0x" + trimHex)
			beforeTokenAddress = transaction.To().String()
			beforeValue = decimal.NewFromBigInt(rawValue, 0).BigInt()
		}
	} else {
		beforeToAddress = transaction.To().String()
		beforeTokenAddress = common.Address{}.String()
		beforeValue = transaction.Value()
	}

	var txStatus account.TxStatus
	if receipt.Status == 1 {
		txStatus = account.TxStatus_Success
	} else {
		txStatus = account.TxStatus_Failed
	}
	return &account.TxHashResponse{
		Code: global_const.ReturnCode_SUCCESS,
		Msg:  "get transaction success",
		Tx: &account.TxMessage{
			Hash:            transaction.Hash().Hex(),
			Index:           uint32(receipt.TransactionIndex),
			From:            beforeTokenAddress, // 代币合约地址
			To:              beforeToAddress,    // 实际接收地址
			Value:           beforeValue.String(),
			Fee:             transaction.GasFeeCap().String(),
			Status:          txStatus,
			Type:            0,
			Height:          receipt.BlockNumber.String(),
			ContractAddress: beforeTokenAddress,
			Data:            hexutils.BytesToHex(transaction.Data()),
		},
	}, nil
}

// 批量获取区块头信息
func (c *ChainAdaptor) GetBlockByRange(req *account.BlockByRangeRequest) (*account.BlockByRangeResponse, error) {
	startBlock := new(big.Int)
	endBlock := new(big.Int)
	startBlock.SetString(req.Start, 10)
	endBlock.SetString(req.End, 10)
	blockRange, err := c.EthClient.BlockHeadersByRange(startBlock, endBlock, 1)
	if err != nil {
		log.Error("get block range fail", "err", err)
		return &account.BlockByRangeResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get block range fail",
		}, err
	}
	var headerList []*account.BlockHeader
	for _, block := range blockRange {
		blockItem := &account.BlockHeader{
			ParentHash:       block.ParentHash.String(),
			UncleHash:        block.UncleHash.String(),
			CoinBase:         block.Coinbase.String(),
			Root:             block.Root.String(),
			TxHash:           block.TxHash.String(),
			ReceiptHash:      block.ReceiptHash.String(),
			ParentBeaconRoot: block.ParentBeaconRoot.String(),
			Difficulty:       block.Difficulty.String(),
			Number:           block.Number.String(),
			GasLimit:         block.GasLimit,
			GasUsed:          block.GasUsed,
			Time:             block.Time,
			Extra:            string(block.Extra),
			MixDigest:        block.MixDigest.String(),
			Nonce:            strconv.FormatUint(block.Nonce.Uint64(), 10),
			BaseFee:          block.BaseFee.String(),
			WithdrawalsHash:  block.WithdrawalsHash.String(),
			BlobGasUsed:      *block.BlobGasUsed,
			ExcessBlobGas:    *block.ExcessBlobGas,
		}
		headerList = append(headerList, blockItem)
	}
	return &account.BlockByRangeResponse{
		Code:        global_const.ReturnCode_SUCCESS,
		Msg:         "get block range success",
		BlockHeader: headerList,
	}, nil
}

// 构建符合 EIP-1559 标准的未签名的交易
func (c *ChainAdaptor) BuildUnSignTransaction(req *account.UnSignTransactionRequest) (*account.UnSignTransactionResponse, error) {
	dFeeTx, _, err := buildDynamicFeeTx(req.Base64Tx)
	if err != nil {
		log.Error("build dynamic fee tx fail", "err", err)
		return &account.UnSignTransactionResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "build dynamic fee tx fail",
		}, err
	}
	log.Info("ethereum BuildUnSignTransaction", "dFeeTx", util.ToJSONString(dFeeTx))

	rawTx, err := CreateEip1559UnSignTx(dFeeTx, dFeeTx.ChainID)
	if err != nil {
		log.Error("create eip1559 unSign tx fail", "err", err)
		return &account.UnSignTransactionResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "create eip1559 unSign tx fail",
		}, err
	}

	return &account.UnSignTransactionResponse{
		Code:     global_const.ReturnCode_SUCCESS,
		Msg:      "build dynamic fee tx success",
		UnSignTx: rawTx,
	}, nil
}

// 构建并验证 EIP-1559 标准签名交易,将签名后的交易数据序列化，并验证签名地址的合法性
func (c *ChainAdaptor) BuildSignedTransaction(req *account.SignedTransactionRequest) (*account.SignedTransactionResponse, error) {
	dFeeTx, dynamicFeeTx, err := buildDynamicFeeTx(req.Base64Tx)
	if err != nil {
		log.Error("build dynamic fee tx fail", "err", err)
		return &account.SignedTransactionResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "build dynamic fee tx fail",
		}, err
	}
	log.Info("ethereum BuildSignedTransaction", "dFeeTx", util.ToJSONString(dFeeTx))
	log.Info("ethereum BuildSignedTransaction", "dynamicFeeTx", util.ToJSONString(dynamicFeeTx))
	log.Info("ethereum BuildSignedTransaction", "req.Signature", req.Signature)

	inputSignatureByteList, err := hex.DecodeString(req.Signature)
	if err != nil {
		log.Error("decode signature fail", "err", err)
		return &account.SignedTransactionResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "decode signature fail",
		}, err
	}

	signer, signedTx, rawTx, txHash, err := CreateEip1559SignedTx(dFeeTx, inputSignatureByteList, dFeeTx.ChainID)
	if err != nil {
		log.Error("create eip1559 signed tx fail", "err", err)
		return &account.SignedTransactionResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "create eip1559 signed tx fail",
		}, err
	}
	log.Info("ethereum BuildSignedTransaction", "rawTx", rawTx)

	//	验证签名地址
	sender, err := types.Sender(signer, signedTx)
	fmt.Println("sender", sender)
	if err != nil {
		log.Error("get sender fail", "err", err)
		return &account.SignedTransactionResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get sender fail",
		}, err
	}
	if sender.Hex() != dynamicFeeTx.FromAddress {
		log.Error("sender mismatch",
			"expected", dynamicFeeTx.FromAddress,
			"got", sender.Hex(),
		)
		return nil, fmt.Errorf("sender address mismatch: expected %s, got %s",
			dynamicFeeTx.FromAddress,
			sender.Hex(),
		)
	}
	log.Info("ethereum BuildSignedTransaction", "sender", sender.Hex())

	return &account.SignedTransactionResponse{
		Code:     global_const.ReturnCode_SUCCESS,
		Msg:      txHash,
		SignedTx: rawTx,
	}, nil

}

func (c *ChainAdaptor) DecodeTransaction(req *account.DecodeTransactionRequest) (*account.DecodeTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ChainAdaptor) VerifySignedTransaction(req *account.VerifyTransactionRequest) (*account.VerifyTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ChainAdaptor) GetExtraData(req *account.ExtraDataRequest) (*account.ExtraDataResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ChainAdaptor) GetNftListByAddress(req *account.NftAddressRequest) (*account.NftAddressResponse, error) {
	//TODO implement me
	panic("implement me")
}

// 将Base64编码的EIP-1559交易请求的Eip1559DynamicFeeTx结构体 转换为以太坊动态费用交易结构（DynamicFeeTx），支持ETH原生转账和ERC20转账。
func buildDynamicFeeTx(base64Tx string) (*types.DynamicFeeTx, *Eip1559DynamicFeeTx, error) {
	// 1. 将 Base64 字符串还原为原始 JSON 数据。
	txReqJsonByte, err := base64.StdEncoding.DecodeString(base64Tx)

	if err != nil {
		log.Error("decode string fail", "err", err)
		return nil, nil, err
	}

	// 2.  JSON 解析 ，映射到 struct
	var dynamicFeeTx Eip1559DynamicFeeTx
	if err := json.Unmarshal(txReqJsonByte, &dynamicFeeTx); err != nil {
		log.Error("parse json fail", "err", err)
		return nil, nil, err
	}
	fmt.Println("原始 JSON 数据", util.ToPrettyJSON(dynamicFeeTx))

	// 3. 数值转换 将字符串形式的数值转换为 big.Int（以太坊数值类型）。
	chainID := new(big.Int)
	maxPriorityFeePerGas := new(big.Int)
	maxFeePerGas := new(big.Int)
	amount := new(big.Int)

	if _, ok := chainID.SetString(dynamicFeeTx.ChainId, 10); !ok {
		return nil, nil, fmt.Errorf("invalid chain ID: %s", dynamicFeeTx.ChainId)
	}
	if _, ok := maxPriorityFeePerGas.SetString(dynamicFeeTx.MaxPriorityFeePerGas, 10); !ok {
		return nil, nil, fmt.Errorf("invalid max priority fee: %s", dynamicFeeTx.MaxPriorityFeePerGas)
	}
	if _, ok := maxFeePerGas.SetString(dynamicFeeTx.MaxFeePerGas, 10); !ok {
		return nil, nil, fmt.Errorf("invalid max fee: %s", dynamicFeeTx.MaxFeePerGas)
	}
	if _, ok := amount.SetString(dynamicFeeTx.Amount, 10); !ok {
		return nil, nil, fmt.Errorf("invalid amount: %s", dynamicFeeTx.Amount)
	}

	// 4. 交易类型判断
	toAddress := common.HexToAddress(dynamicFeeTx.ToAddress)
	var finalToAddress common.Address
	var finalAmount *big.Int
	var buildData []byte
	log.Info("contract address check",
		"contractAddress", dynamicFeeTx.ContractAddress,
		"isEthTransfer", isEthTransfer(&dynamicFeeTx),
	)

	if isEthTransfer(&dynamicFeeTx) {
		// ETH 转账
		finalToAddress = toAddress
		finalAmount = amount
	} else {
		// ERC20 代币转账
		contractAddress := common.HexToAddress(dynamicFeeTx.ContractAddress)
		buildData = BuildErc20Data(toAddress, amount)
		finalToAddress = contractAddress
		finalAmount = big.NewInt(0)
	}

	// 6. Create dynamic fee transaction
	dFeeTx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     dynamicFeeTx.Nonce,
		GasTipCap: maxPriorityFeePerGas,
		GasFeeCap: maxFeePerGas,
		Gas:       dynamicFeeTx.GasLimit,
		To:        &finalToAddress,
		Value:     finalAmount,
		Data:      buildData,
	}

	// 以太坊动态费用交易结构 、解析后的原始请求结构、err
	return dFeeTx, &dynamicFeeTx, nil
}

// 判断是否为 ETH 转账
func isEthTransfer(tx *Eip1559DynamicFeeTx) bool {
	// 检查合约地址是否为空或零地址
	if tx.ContractAddress == "" ||
		tx.ContractAddress == "0x0000000000000000000000000000000000000000" ||
		tx.ContractAddress == "0x00" {
		return true
	}
	return false
}

// 解压公钥（兼容压缩和非压缩格式）
func decompressPublicKey(pubKeyBytes []byte) (*btcec.PublicKey, error) {
	if len(pubKeyBytes) == 33 {
		// 压缩公钥（0x02/0x03 开头）
		return btcec.ParsePubKey(pubKeyBytes)
	} else if len(pubKeyBytes) == 65 && pubKeyBytes[0] == 0x04 {
		// 非压缩公钥（0x04 开头）
		return btcec.ParsePubKey(pubKeyBytes)
	} else {
		return nil, fmt.Errorf("invalid public key format")
	}
}
