package config

import (
	"os"

	"gopkg.in/yaml.v2"

	"github.com/ethereum/go-ethereum/log"
)

type Server struct {
	Port string `yaml:"port"`
}

type Node struct {
	RpcUrl       string `yaml:"rpc_url"`
	RpcUser      string `yaml:"rpc_user"`
	RpcPass      string `yaml:"rpc_pass"`
	DataApiUrl   string `yaml:"data_api_url"`
	DataApiKey   string `yaml:"data_api_key"`
	DataApiToken string `yaml:"data_api_token"`
	TimeOut      uint64 `yaml:"time_out"`
}

type WalletNode struct {
	Eth Node `yaml:"eth"`
}

type Config struct {
	Server     Server     `yaml:"server"`
	WalletNode WalletNode `yaml:"wallet_node"`
	NetWork    string     `yaml:"network"`
	Chains     []string   `yaml:"chains"`
}

func NewConfig(path string) (*Config, error) {
	config := &Config{}
	h := log.NewTerminalHandler(os.Stdout, true)
	log.SetDefault(log.NewLogger(h))
	// 读取文件
	file, err := os.ReadFile(path)
	if err != nil {
		log.Error("read config file error", "err", err)
		return nil, err
	}
	// 将文件内容解析成结构体
	err = yaml.Unmarshal(file, config)
	if err != nil {
		log.Error("unmarshal config file error", "err", err)
		return nil, err
	}
	return config, nil
}

const UnsupportedOperation = "Unsupport chain"
