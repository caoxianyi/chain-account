package ethereum

import (
	"encoding/hex"
	"math/big"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

// 构建符合 ERC20 标准的代币转账数据 toAddress 接收方地址 amount 余额
func BuildErc20Data(toAddress common.Address, amount *big.Int) []byte {
	var data []byte

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := crypto.Keccak256Hash(transferFnSignature)
	methodId := hash[:4]
	dataAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	dataAmount := common.LeftPadBytes(amount.Bytes(), 32)

	data = append(data, methodId...)
	data = append(data, dataAddress...)
	data = append(data, dataAmount...)

	return data
}

// 构建符合 ERC721 标准的代币转移（NFT 转账）数据
// fromAddress 当前 NFT 持有者的地址（转账发起方）
// NFT  toAddress 接收方地址
// tokenId 要转移的 NFT 唯一标识符
func BuildErc721Data(fromAddress, toAddress common.Address, tokenId *big.Int) []byte {
	var data []byte

	transferFnSignature := []byte("safeTransferFrom(address,address,uint256)")
	hash := crypto.Keccak256Hash(transferFnSignature)
	methodId := hash[:4]

	dataFromAddress := common.LeftPadBytes(fromAddress.Bytes(), 32)
	dataToAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	dataTokenId := common.LeftPadBytes(tokenId.Bytes(), 32)

	data = append(data, methodId...)
	data = append(data, dataFromAddress...)
	data = append(data, dataToAddress...)
	data = append(data, dataTokenId...)

	return data
}

// 生成 未签名的传统以太坊交易（Legacy Transaction）的哈希
// txData 传统交易数据结构
func CreateLegacyUnSignTx(txData *types.LegacyTx, chainId *big.Int) string {
	tx := types.NewTx(txData)
	signer := types.LatestSignerForChainID(chainId)
	txHash := signer.Hash(tx)
	return txHash.String()
}

// 创建符合EIP-1559标准的未签名交易的哈希
// 构建未签名的 rawTx 32位的massageHash
func CreateEip1559UnSignTx(txData *types.DynamicFeeTx, chainId *big.Int) (string, error) {
	tx := types.NewTx(txData)
	// 序列化交易
	signer := types.LatestSignerForChainID(chainId)
	rawTx := signer.Hash(tx)
	return rawTx.String(), nil
}

// 对传统以太坊交易（Legacy Transaction）进行签名并序列化
// txData 传统交易数据
// signature 原始 ECDSA 签名（65 字节，R|S|V）
// chainId 区块链网络 ID（如以太坊主网为 1）。
func CreateLegacySignedTx(txData *types.LegacyTx, signature []byte, chainId *big.Int) (string, string, error) {
	tx := types.NewTx(txData)
	signer := types.LatestSignerForChainID(chainId)
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return "", "", errors.New("tx with signature fail")
	}
	signedTxData, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return "", "", errors.New("encode tx to byte fail")
	}
	// 编码后的交易数据、交易哈希（交易唯一标识）、err
	return "0x" + hex.EncodeToString(signedTxData), signedTx.Hash().String(), nil
}

// 对符合 EIP-1559 标准的动态费用交易进行签名并序列化
// EIP-1559 交易数据（含 GasTipCap、GasFeeCap 等）
func CreateEip1559SignedTx(txData *types.DynamicFeeTx, signature []byte, chainId *big.Int) (types.Signer, *types.Transaction, string, string, error) {
	tx := types.NewTx(txData)
	// 序列化交易
	signer := types.LatestSignerForChainID(chainId)
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return nil, nil, "", "", errors.New("tx with signature fail")
	}
	// RLP 编码
	signedTxData, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, nil, "", "", errors.New("encode tx to byte fail")
	}
	//用于验证签名的签名器实例、包含完整签名数据的交易对象、编码的原始交易数据（用来广播）、交易哈希（唯一标识）、err
	return signer, signedTx, "0x" + hex.EncodeToString(signedTxData)[4:], signedTx.Hash().String(), nil
}
