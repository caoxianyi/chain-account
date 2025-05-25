package dispatcher

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"chain-account/chain"
	"chain-account/chain/ethereum"
	"chain-account/common/global_const"
	"chain-account/config"
	"chain-account/rpc/account"
)

type CommonRequest interface {
	GetChain() string
}

// 错误返回值
type CommonReply struct {
	Code    int32  `json:"code"`
	Msg     string `json:"msg"`
	Support bool   `json:"support"`
}

type ChainDispatcher struct {
	registry map[string]chain.IChainAdaptor // 每一条链 都对应一套接口
}

// 初始化适配器
func NewChainDispatcher(conf *config.Config) (*ChainDispatcher, error) {
	dispatcher := ChainDispatcher{
		registry: make(map[string]chain.IChainAdaptor),
	}

	chainAdaptorFactoryMap := map[string]func(*config.Config) (chain.IChainAdaptor, error){
		ethereum.ChainName: ethereum.NewChainAdaptor,
	}
	supportedChains := []string{
		ethereum.ChainName,
	}

	for _, chainName := range conf.Chains {
		if factory, ok := chainAdaptorFactoryMap[chainName]; ok {
			adaptor, err := factory(conf)
			if err != nil {
				log.Crit("failed to setup chain", "chain", chainName, "error", err)
			}
			dispatcher.registry[chainName] = adaptor
		} else {
			log.Error("unsupported chain", "chain", chainName, "supportedChains", supportedChains)
		}
	}
	return &dispatcher, nil
}

// 拦截器
func (d *ChainDispatcher) Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("panic error", "msg", e)
			log.Debug(string(debug.Stack()))
			err = status.Errorf(codes.Internal, "Panic err: %v", e)
		}
	}()

	pos := strings.LastIndex(info.FullMethod, "/")
	method := info.FullMethod[pos+1:]

	chainName := req.(CommonRequest).GetChain()
	log.Info(method, "chain", chainName, "req", req)

	resp, err = handler(ctx, req)
	log.Debug("Finish handling", "resp", resp, "err", err)
	return
}

// 预处理
func (d *ChainDispatcher) preHandler(req interface{}) (*CommonReply, string) {
	chainName := req.(CommonRequest).GetChain()
	if _, ok := d.registry[chainName]; !ok {
		return &CommonReply{
			Code:    global_const.ReturnCode_ERROR,
			Msg:     config.UnsupportedOperation,
			Support: false,
		}, chainName
	}
	return nil, chainName
}

func (d *ChainDispatcher) GetSupportChains(ctx context.Context, request *account.SupportChainsRequest) (*account.SupportChainsResponse, error) {
	fmt.Println("进这里 1")
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.SupportChainsResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  config.UnsupportedOperation,
		}, nil
	}
	return d.registry[chainName].GetSupportChains(request)
}

func (d *ChainDispatcher) ConvertAddress(ctx context.Context, request *account.ConvertAddressRequest) (*account.ConvertAddressResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.ConvertAddressResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "covert address fail at pre handle",
		}, nil
	}
	return d.registry[chainName].ConvertAddress(request)
}
func (d *ChainDispatcher) ValidAddress(ctx context.Context, request *account.ValidAddressRequest) (*account.ValidAddressResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.ValidAddressResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "valid address error at pre handle",
		}, nil
	}
	return d.registry[chainName].ValidAddress(request)
}

func (d *ChainDispatcher) GetBlockByNumber(ctx context.Context, request *account.BlockNumberRequest) (*account.BlockResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.BlockResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get block by number fail at pre handle",
		}, nil
	}
	return d.registry[chainName].GetBlockByNumber(request)
}

func (d *ChainDispatcher) GetBlockByHash(ctx context.Context, request *account.BlockHashRequest) (*account.BlockResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.BlockResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get block by hash fail at pre handle",
		}, nil
	}
	return d.registry[chainName].GetBlockByHash(request)
}

func (d *ChainDispatcher) GetBlockHeaderByHash(ctx context.Context, request *account.BlockHeaderHashRequest) (*account.BlockHeaderResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.BlockHeaderResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get block header by hash fail at pre handle",
		}, nil
	}
	return d.registry[chainName].GetBlockHeaderByHash(request)
}

func (d *ChainDispatcher) GetBlockHeaderByNumber(ctx context.Context, request *account.BlockHeaderNumberRequest) (*account.BlockHeaderResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.BlockHeaderResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get block header by number fail at pre handle",
		}, nil
	}
	return d.registry[chainName].GetBlockHeaderByNumber(request)
}

func (d *ChainDispatcher) GetBlockHeaderByRange(ctx context.Context, request *account.BlockByRangeRequest) (*account.BlockByRangeResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.BlockByRangeResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get block range header fail at pre handle",
		}, nil
	}
	return d.registry[chainName].GetBlockByRange(request)
}

func (d *ChainDispatcher) GetAccount(ctx context.Context, request *account.AccountRequest) (*account.AccountResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.AccountResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get account information fail at pre handle",
		}, nil
	}
	return d.registry[chainName].GetAccount(request)
}

func (d *ChainDispatcher) GetFee(ctx context.Context, request *account.FeeRequest) (*account.FeeResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.FeeResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get fee fail at pre handle",
		}, nil
	}
	return d.registry[chainName].GetFee(request)
}

func (d *ChainDispatcher) SendTx(ctx context.Context, request *account.SendTxRequest) (*account.SendTxResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.SendTxResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "send tx fail at pre handle",
		}, nil
	}
	return d.registry[chainName].SendTx(request)
}

func (d *ChainDispatcher) GetTxByAddress(ctx context.Context, request *account.TxAddressRequest) (*account.TxAddressResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.TxAddressResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get tx by address fail pre handle",
		}, nil
	}
	return d.registry[chainName].GetTxByAddress(request)
}

func (d *ChainDispatcher) GetTxByHash(ctx context.Context, request *account.TxHashRequest) (*account.TxHashResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.TxHashResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get tx by hash fail at pre handle",
		}, nil
	}
	return d.registry[chainName].GetTxByHash(request)
}

func (d *ChainDispatcher) GetBlockByRange(ctx context.Context, request *account.BlockByRangeRequest) (*account.BlockByRangeResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.BlockByRangeResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get blcok by range fail at pre handle",
		}, nil
	}
	return d.registry[chainName].GetBlockByRange(request)
}

func (d *ChainDispatcher) BuildUnSignTransaction(ctx context.Context, request *account.UnSignTransactionRequest) (*account.UnSignTransactionResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.UnSignTransactionResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get un sign tx fail at pre handle",
		}, nil
	}
	return d.registry[chainName].BuildUnSignTransaction(request)
}

func (d *ChainDispatcher) BuildSignedTransaction(ctx context.Context, request *account.SignedTransactionRequest) (*account.SignedTransactionResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.SignedTransactionResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "signed tx fail at pre handle",
		}, nil
	}
	return d.registry[chainName].BuildSignedTransaction(request)
}

func (d *ChainDispatcher) DecodeTransaction(ctx context.Context, request *account.DecodeTransactionRequest) (*account.DecodeTransactionResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.DecodeTransactionResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "decode tx fail at pre handle",
		}, nil
	}
	return d.registry[chainName].DecodeTransaction(request)
}

func (d *ChainDispatcher) VerifySignedTransaction(ctx context.Context, request *account.VerifyTransactionRequest) (*account.VerifyTransactionResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.VerifyTransactionResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "verify tx fail at pre handle",
		}, nil
	}
	return d.registry[chainName].VerifySignedTransaction(request)
}

func (d *ChainDispatcher) GetExtraData(ctx context.Context, request *account.ExtraDataRequest) (*account.ExtraDataResponse, error) {
	resp, chainName := d.preHandler(request)
	if resp != nil {
		return &account.ExtraDataResponse{
			Code: global_const.ReturnCode_ERROR,
			Msg:  "get extra data fail at pre handle",
		}, nil
	}
	return d.registry[chainName].GetExtraData(request)
}

func (d *ChainDispatcher) GetNftListByAddress(ctx context.Context, request *account.NftAddressRequest) (*account.NftAddressResponse, error) {
	panic("implement me")
}

func (d *ChainDispatcher) GetNftCollection(ctx context.Context, request *account.NftCollectionRequest) (*account.NftCollectionResponse, error) {
	panic("implement me")
}

func (d *ChainDispatcher) GetNftDetail(ctx context.Context, request *account.NftDetailRequest) (*account.NftDetailResponse, error) {
	panic("implement me")
}

func (d *ChainDispatcher) GetNftHolderList(ctx context.Context, request *account.NftHolderListRequest) (*account.NftHolderListResponse, error) {
	panic("implement me")
}

func (d *ChainDispatcher) GetNftTradeHistory(ctx context.Context, request *account.NftTradeHistoryRequest) (*account.NftTradeHistoryResponse, error) {
	panic("implement me")
}

func (d *ChainDispatcher) GetAddressNftTradeHistory(ctx context.Context, request *account.AddressNftTradeHistoryRequest) (*account.AddressNftTradeHistoryResponse, error) {
	panic("implement me")
}
