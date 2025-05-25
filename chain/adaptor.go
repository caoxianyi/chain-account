package chain

import "chain-account/rpc/account"

type IChainAdaptor interface {
	GetSupportChains(req *account.SupportChainsRequest) (*account.SupportChainsResponse, error)                // 获取支持的链
	ConvertAddress(req *account.ConvertAddressRequest) (*account.ConvertAddressResponse, error)                // 地址转换
	ValidAddress(req *account.ValidAddressRequest) (*account.ValidAddressResponse, error)                      // 地址校验
	GetBlockByNumber(req *account.BlockNumberRequest) (*account.BlockResponse, error)                          // 获取区块信息
	GetBlockByHash(req *account.BlockHashRequest) (*account.BlockResponse, error)                              // 获取区块信息
	GetBlockHeaderByHash(req *account.BlockHeaderHashRequest) (*account.BlockHeaderResponse, error)            // 获取区块头信息
	GetBlockHeaderByNumber(req *account.BlockHeaderNumberRequest) (*account.BlockHeaderResponse, error)        // 获取区块头信息
	GetAccount(req *account.AccountRequest) (*account.AccountResponse, error)                                  // 获取账户信息
	GetFee(req *account.FeeRequest) (*account.FeeResponse, error)                                              // 获取手续费
	SendTx(req *account.SendTxRequest) (*account.SendTxResponse, error)                                        // 发送交易
	GetTxByAddress(req *account.TxAddressRequest) (*account.TxAddressResponse, error)                          // 获取地址交易信息
	GetTxByHash(req *account.TxHashRequest) (*account.TxHashResponse, error)                                   // 获取交易信息
	GetBlockByRange(req *account.BlockByRangeRequest) (*account.BlockByRangeResponse, error)                   // 获取区块信息
	BuildUnSignTransaction(req *account.UnSignTransactionRequest) (*account.UnSignTransactionResponse, error)  // 构建未签名交易
	BuildSignedTransaction(req *account.SignedTransactionRequest) (*account.SignedTransactionResponse, error)  // 构建签名交易
	DecodeTransaction(req *account.DecodeTransactionRequest) (*account.DecodeTransactionResponse, error)       // 解码交易
	VerifySignedTransaction(req *account.VerifyTransactionRequest) (*account.VerifyTransactionResponse, error) // 验证签名交易
	GetExtraData(req *account.ExtraDataRequest) (*account.ExtraDataResponse, error)                            // 获取额外数据
	GetNftListByAddress(req *account.NftAddressRequest) (*account.NftAddressResponse, error)                   // 获取NFT列表
}
