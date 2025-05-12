package ethereum

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"chain-account/common/global_const"
	"chain-account/common/helpers"
	"chain-account/common/retry"
)

const (
	defaultDialTimeout    = 5 * time.Second
	defaultDialAttempts   = 5
	defaultRequestTimeout = 10 * time.Second
)

// 定义交易列表
type TransactionList struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Hash  string `json:"hash"`
	Value string `json:"value"`
}

// 定义rpc的block
type RpcBlock struct {
	Hash         common.Hash       `json:"hash"`
	Number       string            `json:"number"`
	BaseFee      string            `json:"baseFeePerGas"`
	Transactions []TransactionList `json:"transactions"`
}

// // 转换为uint64
func (b *RpcBlock) NumberUint64() (uint64, error) {
	return hexutil.DecodeUint64(b.Number)
}

// 定义日志
type Logs struct {
	Logs          []types.Log
	ToBlockHeader *types.Header
}

// 定义rpc接口
type IRpc interface {
	Close()
	CallContext(ctx context.Context, result any, method string, args ...any) error
	BatchCallContext(ctx context.Context, b []rpc.BatchElem) error
}

// 定义eth的接口
type IEth interface {
	// 区块数据相关
	BlockHeaderByNumber(*big.Int) (*types.Header, error)
	BlockHeaderByHash(common.Hash) (*types.Header, error)
	BlockHeadersByRange(*big.Int, *big.Int, uint) ([]types.Header, error)
	BlockByNumber(*big.Int) (*RpcBlock, error)
	BlockByHash(common.Hash) (*RpcBlock, error)
	LatestSafeBlockHeader() (*types.Header, error)
	LatestFinalizedBlockHeader() (*types.Header, error)
	// 账户与交易
	TxCountByAddress(common.Address) (hexutil.Uint64, error)
	SendRawTransaction(rawTx string) (*common.Hash, error)
	TxByHash(common.Hash) (*types.Transaction, error)
	TxReceiptByHash(common.Hash) (*types.Receipt, error)
	StorageHash(common.Address, *big.Int) (common.Hash, error)
	EthGetCode(common.Address) (string, error)
	GetBalance(common.Address) (*big.Int, error)
	// Gas 费用估算
	SuggestGasPrice() (*big.Int, error)
	SuggestGasTipCap() (*big.Int, error)
	// 智能合约交互
	FilterLogs(ethereum.FilterQuery, uint) (Logs, error)
	Close()
}

// 定义Eth客户端
type EthClient struct {
	rpc IRpc
}

// 初始化Eth客户端 需要全部实现IEth的接口
func NewEthClient(ctx context.Context, rpcUrl string) (IEth, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDialTimeout) // 超时设置
	defer cancel()

	bOff := retry.Exponential()
	// 尝试 重试链接 5次
	rpcClient, err := retry.Do(ctx, defaultDialAttempts, bOff, func() (*rpc.Client, error) {
		if !helpers.IsURLAvailable(rpcUrl) {
			return nil, fmt.Errorf("address unavailable (%s)", rpcUrl)
		}

		client, err := rpc.DialContext(ctx, rpcUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to dial address (%s): %w", rpcUrl, err)
		}

		return client, nil
	})
	if err != nil {
		return nil, err
	}

	return &EthClient{
		rpc: NewRPC(rpcClient), // 初始化rpc客户端
	}, nil
}

// 通过区块号获取区块头  number: 目标区块号	验证区块高度
func (e *EthClient) BlockHeaderByNumber(number *big.Int) (*types.Header, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var header *types.Header
	err := e.rpc.CallContext(ctx, &header, "eth_getBlockByNumber", toBlockNumArg(number), false)
	if err != nil {
		log.Error("Call eth_getBlockByNumber method fail", "err", err)
		return nil, err
	} else if header == nil {
		log.Warn("header not found")
		return nil, ethereum.NotFound
	}
	return header, nil
}

// 通过区块Hash获取区块头 hash: 区块哈希值 追踪特定区块
func (e *EthClient) BlockHeaderByHash(hash common.Hash) (*types.Header, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var header *types.Header
	err := e.rpc.CallContext(ctx, &header, "eth_getBlockByHash", hash, false)
	if err != nil {
		log.Error("Call eth_getBlockByHash method fail", "err", err)
		return nil, err
	} else if header == nil {
		log.Warn("header not found")
		return nil, ethereum.NotFound
	}
	if header.Hash() != hash {
		return nil, errors.New("header mismatch")
	}
	return header, nil
}

// 批量获取区块头范围 start, end: 起始/结束区块号	区块数据同步
func (e *EthClient) BlockHeadersByRange(startHeight, endHeight *big.Int, chainId uint) ([]types.Header, error) {
	/*
		1：左侧值 (startHeight) > 右侧值 (endHeight)
		0：两侧值相等
		-1：左侧值 < 右侧值
	*/
	// 判断是否请求单个区块
	if startHeight.Cmp(endHeight) == 0 {
		header, err := e.BlockHeaderByNumber(startHeight)
		if err != nil {
			return nil, err
		}
		return []types.Header{*header}, nil
	}
	// 计算两个区块之间的区块总数 区块数 = (end - start) + 1
	count := new(big.Int).Sub(endHeight, startHeight).Uint64() + 1
	headers := make([]types.Header, count)
	batchElems := make([]rpc.BatchElem, count)
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	if chainId == uint(global_const.ZkFairSepoliaChainId) || chainId == uint(global_const.ZkFairChainId) {
		groupSize := 100
		var wg sync.WaitGroup
		numGroups := (int(count)-1)/groupSize + 1
		wg.Add(numGroups)

		for i := 0; i < int(count); i += groupSize {
			start := i
			end := i + groupSize - 1
			if end > int(count) {
				end = int(count) - 1
			}
			go func(start, end int) {
				defer wg.Done()
				for j := start; j <= end; j++ {
					height := new(big.Int).Add(startHeight, new(big.Int).SetUint64(uint64(j)))
					// 创建 BatchElem 请求
					batchElems[j] = rpc.BatchElem{
						Method: "eth_getBlockByNumber",
						Result: new(types.Header),
						Error:  nil,
					}
					header := new(types.Header)
					batchElems[j].Error = e.rpc.CallContext(ctx, header, batchElems[j].Method, toBlockNumArg(height), false)
					batchElems[j].Result = header
				}
			}(start, end)
		}
		wg.Wait()
	} else {
		for i := uint64(0); i < count; i++ {
			height := new(big.Int).Add(startHeight, new(big.Int).SetUint64(i))
			// 创建 BatchElem 请求
			batchElems[i] = rpc.BatchElem{
				Method: "eth_getBlockByNumber",
				Args: []interface{}{
					toBlockNumArg(height),
					false,
				},
				Result: &headers[i],
			}
		}
		err := e.rpc.BatchCallContext(ctx, batchElems)
		if err != nil {
			return nil, err
		}
	}

	fmt.Println("batchElems", batchElems)
	size := 0
	for i, batchElem := range batchElems {
		header, ok := batchElem.Result.(*types.Header)
		if !ok {
			return nil, fmt.Errorf("unable to transform rpc response %v into types.Header", batchElem.Result)
		}
		headers[i] = *header
		size = size + 1
	}
	headers = headers[:size]
	return headers, nil
}

// 通过区块号获取区块数据  number: 目标区块号	分析区块内容
func (e *EthClient) BlockByNumber(number *big.Int) (*RpcBlock, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var block *RpcBlock
	err := e.rpc.CallContext(ctx, &block, "eth_getBlockByNumber", toBlockNumArg(number), true)
	if err != nil {
		log.Error("Call eth_getBlockByNumber method fail", "err", err)
		return nil, err
	} else if block == nil {
		log.Warn("header not found")
		return nil, ethereum.NotFound
	}
	return block, nil
}

// 通过区块Hash获取区块数据 hash: 区块哈希值	交易追溯
func (e *EthClient) BlockByHash(hash common.Hash) (*RpcBlock, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var block *RpcBlock
	err := e.rpc.CallContext(ctx, &block, "eth_getBlockByHash", hash, true)
	if err != nil {
		log.Error("Call eth_getBlockByHash method fail", "err", err)
		return nil, err
	} else if block == nil {
		log.Warn("header not found")
		return nil, ethereum.NotFound
	}
	return block, nil
}

// 获取最新安全区块头 PoS 链状态监控
func (e *EthClient) LatestSafeBlockHeader() (*types.Header, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var header *types.Header
	err := e.rpc.CallContext(ctx, &header, "eth_getBlockByNumber", "safe", false)
	if err != nil {
		return nil, err
	} else if header == nil {
		return nil, ethereum.NotFound
	}
	return header, nil
}

// 获取最终确认区块头 PoS 链状态监控
func (e *EthClient) LatestFinalizedBlockHeader() (*types.Header, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var header *types.Header
	err := e.rpc.CallContext(ctx, &header, "eth_getBlockByNumber", "finalized", false)
	if err != nil {
		return nil, err
	} else if header == nil {
		return nil, ethereum.NotFound
	}
	return header, nil
}

// 获取地址交易次数(Nonce) address: 钱包地址	最新交易Nonce计算
func (e *EthClient) TxCountByAddress(address common.Address) (hexutil.Uint64, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var nonce hexutil.Uint64
	err := e.rpc.CallContext(ctx, &nonce, "eth_getTransactionCount", address, "latest")
	if err != nil {
		log.Error("Call eth_getTransactionCount method fail", "err", err)
		return 0, err
	}
	log.Info("get nonce by address success", "nonce", nonce)
	return nonce, err
}

// 广播签名交易 rawTx: 16进制签名	交易提交 返回交易哈希
func (e *EthClient) SendRawTransaction(rawTx string) (*common.Hash, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var txHash common.Hash
	if err := e.rpc.CallContext(ctx, &txHash, "eth_sendRawTransaction", rawTx); err != nil {
		return nil, err
	}
	log.Info("send tx to ethereum success", "txHash", txHash.Hex())
	return &txHash, nil
}

// 按Hash获取交易详情 hash: 交易哈希 交易状态查询
func (e *EthClient) TxByHash(hash common.Hash) (*types.Transaction, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var tx *types.Transaction
	err := e.rpc.CallContext(ctx, &tx, "eth_getTransactionByHash", hash)
	fmt.Println("**********0", err)
	if err != nil {
		return nil, err
	} else if tx == nil {
		return nil, ethereum.NotFound
	}

	return tx, nil
}

// 获取交易收据	hash: 交易哈希	交易结果验证
func (e *EthClient) TxReceiptByHash(hash common.Hash) (*types.Receipt, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var txReceipt *types.Receipt
	err := e.rpc.CallContext(ctx, &txReceipt, "eth_getTransactionReceipt", hash)
	if err != nil {
		return nil, err
	} else if txReceipt == nil {
		return nil, ethereum.NotFound
	}

	return txReceipt, nil
}

// 获取合约存储哈希 address: 合约地址 blockNumber: 区块号 状态验证
func (e *EthClient) StorageHash(address common.Address, blockNumber *big.Int) (common.Hash, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	proof := struct {
		StorageHash common.Hash
	}{}
	err := e.rpc.CallContext(ctx, &proof, "eth_getProof", address, nil, toBlockNumArg(blockNumber))
	if err != nil {
		return common.Hash{}, err
	}

	return proof.StorageHash, nil
}

// 获取合约字节码 account: 合约地址 合约验证
func (e *EthClient) EthGetCode(account common.Address) (string, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var result hexutil.Bytes
	err := e.rpc.CallContext(ctx, &result, "eth_getCode", account, "latest")
	if err != nil {
		return "", err
	}
	if result.String() == "0x" {
		return "eoa", nil // 普通账户
	} else {
		return "contract", nil // 合约账户
	}
}

// 查询地址余额	address: 钱包地址	余额监控
func (e *EthClient) GetBalance(address common.Address) (*big.Int, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var result hexutil.Big
	err := e.rpc.CallContext(ctx, &result, "eth_getBalance", address, "latest")
	if err != nil {
		return nil, fmt.Errorf("get balance failed: %w", err)
	}
	balance := (*big.Int)(&result)
	return balance, nil
}

// 获取基础Gas价 普通交易定价
func (e *EthClient) SuggestGasPrice() (*big.Int, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var hex hexutil.Big
	if err := e.rpc.CallContext(ctx, &hex, "eth_gasPrice"); err != nil {
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

// 获取优先费建议 EIP-1559 交易优化
func (e *EthClient) SuggestGasTipCap() (*big.Int, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()

	var hex hexutil.Big
	if err := e.rpc.CallContext(ctx, &hex, "eth_maxPriorityFeePerGas"); err != nil {
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

// 过滤事件日志 filterQuery: 过滤条件 chainId: 链ID 监听合约事件
func (e *EthClient) FilterLogs(query ethereum.FilterQuery, chainId uint) (Logs, error) {
	arg, err := toFilterArg(query)
	if err != nil {
		return Logs{}, err
	}
	var logs []types.Log
	var header types.Header

	batchElems := make([]rpc.BatchElem, 2)
	batchElems[0] = rpc.BatchElem{
		Method: "eth_getBlockByNumber",
		Args: []interface{}{
			toBlockNumArg(query.ToBlock),
			false,
		},
		Result: &header,
	}
	batchElems[1] = rpc.BatchElem{
		Method: "eth_getLogs",
		Args: []interface{}{
			arg,
		},
		Result: &logs,
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout*10)
	defer cancelFunc()

	if chainId == uint(global_const.ZkFairSepoliaChainId) || chainId == uint(global_const.ZkFairChainId) {
		batchElems[0].Error = e.rpc.CallContext(ctx, &header, batchElems[0].Method, toBlockNumArg(query.ToBlock), false)
		batchElems[1].Error = e.rpc.CallContext(ctx, &logs, batchElems[1].Method, arg)
	} else {
		err = e.rpc.BatchCallContext(ctx, batchElems)
		if err != nil {
			return Logs{}, err
		}
	}

	if batchElems[0].Error != nil {
		return Logs{}, fmt.Errorf("unable to query for the `FilterQuery#ToBlock` header: %w", batchElems[0].Error)
	}
	if batchElems[1].Error != nil {
		return Logs{}, fmt.Errorf("unable to query logs: %w", batchElems[1].Error)
	}
	return Logs{
		Logs:          logs,
		ToBlockHeader: &header,
	}, nil
}

// 关闭eth客户端连接
func (e *EthClient) Close() {
	e.rpc.Close()
}

// 定义rpc客户端
type RpcClient struct {
	rpc *rpc.Client
}

// 初始化rpc客户端 需要全部实现IRpc的接口
func NewRPC(client *rpc.Client) IRpc {
	return &RpcClient{
		client,
	}
}

// 关闭
func (r RpcClient) Close() {
	r.rpc.Close()
}

// 调用
func (r RpcClient) CallContext(ctx context.Context, result any, method string, args ...any) error {
	err := r.rpc.CallContext(ctx, result, method, args...)
	return err
}

// 批量调用
func (r RpcClient) BatchCallContext(ctx context.Context, b []rpc.BatchElem) error {
	err := r.rpc.BatchCallContext(ctx, b)
	return err
}

/*
toBlockNumArg(nil)          // "latest"（最新区块）
toBlockNumArg(big.NewInt(1000))  // "0x3e8"（十进制1000的十六进制）
toBlockNumArg(big.NewInt(-1))    // "earliest"（创世区块）
toBlockNumArg(big.NewInt(-2))    // "pending"（待确认的区块）
*/
func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	if number.Sign() >= 0 {
		return hexutil.EncodeBig(number)
	}
	return rpc.BlockNumber(number.Int64()).String()
}

func toFilterArg(query ethereum.FilterQuery) (interface{}, error) {
	arg := map[string]interface{}{"address": query.Addresses, "topics": query.Topics}
	if query.BlockHash != nil {
		arg["blockHash"] = *query.BlockHash
		if query.FromBlock != nil || query.ToBlock != nil {
			return nil, errors.New("cannot specify both BlockHash and FromBlock/ToBlock")
		}
	} else {
		if query.FromBlock == nil {
			arg["fromBlock"] = "0x0"
		} else {
			arg["fromBlock"] = toBlockNumArg(query.FromBlock)
		}
		arg["toBlock"] = toBlockNumArg(query.ToBlock)
	}
	return arg, nil
}
