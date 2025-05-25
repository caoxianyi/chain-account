package ethereum

import (
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/dapplink-labs/chain-explorer-api/common/account"
	"github.com/dapplink-labs/chain-explorer-api/common/chain"
	"github.com/dapplink-labs/chain-explorer-api/explorer/etherscan"
)

type EthData struct {
	EthScanCli *etherscan.ChainExplorerAdaptor
}

// 初始化EthData
func NewEthData(baseUrl, apiKey string, timeout time.Duration) (*EthData, error) {
	ethScanCli, err := etherscan.NewChainExplorerAdaptor(apiKey, baseUrl, false, timeout)
	if err != nil {
		log.Error("New etherscan client fail", "err", err)
		return nil, err
	}
	return &EthData{
		EthScanCli: ethScanCli,
	}, err
}

// 通过地址获取交易记录
func (ed *EthData) GetTxByAddress(pageNum, pageSize uint64, address string, action account.ActionType) (*account.TransactionResponse[account.AccountTxResponse], error) {
	request := &account.AccountTxRequest{
		PageRequest: chain.PageRequest{
			Page:  pageNum,
			Limit: pageSize,
		},
		Action:  action,
		Address: address,
	}
	txData, err := ed.EthScanCli.GetTxByAddress(request)
	if err != nil {
		return nil, err
	}
	return txData, nil
}

// 通过合约地址 用户地址查询账户代币余额
func (ed *EthData) GetBalanceByAddress(contractAddr, address string) (*account.AccountBalanceResponse, error) {
	accountItem := []string{address}
	symbol := []string{"ETH"}
	contractAddress := []string{contractAddr}
	protocolType := []string{""}
	page := []string{"1"}
	limit := []string{"10"}
	acbr := &account.AccountBalanceRequest{
		ChainShortName:  "ETH",           // 链标识（以太坊主网）
		ExplorerName:    "etherescan",    // 使用的区块链浏览器名称
		Account:         accountItem,     // 目标地址列表
		Symbol:          symbol,          // 查询的代币符号
		ContractAddress: contractAddress, // 代币合约地址列表
		ProtocolType:    protocolType,    // 协议类型过滤
		Page:            page,
		Limit:           limit,
	}
	etherscanResp, err := ed.EthScanCli.GetAccountBalance(acbr)
	if err != nil {
		log.Error("get account balance error", "err", err)
		return nil, err
	}
	return etherscanResp, nil
}
