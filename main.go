package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"chain-account/config"
	"chain-account/dispatcher"
	"chain-account/rpc/account"
)

func main() {
	// 读取yaml配置文件
	var f = flag.String("c", "config.yml", "config path")
	flag.Parse()
	conf, err := config.NewConfig(*f)
	if err != nil {
		panic(err)
	}
	// 初始化适配器
	dispatch, err := dispatcher.NewChainDispatcher(conf)
	if err != nil {
		log.Error("Setup dispatcher failed", "err", err)
		panic(err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(dispatch.Interceptor))
	defer grpcServer.GracefulStop()

	account.RegisterWalletAccountServiceServer(grpcServer, dispatch) // 注册服务

	listen, err := net.Listen("tcp", fmt.Sprintf(":"+conf.Server.Port))
	if err != nil {
		log.Error("net listen failed", "err", err)
		panic(err)
	}
	reflection.Register(grpcServer)

	log.Info("wallet rpc services start success", "port", conf.Server.Port)

	if err := grpcServer.Serve(listen); err != nil {
		log.Error("grpc server serve failed", "err", err)
		panic(err)
	}
}
