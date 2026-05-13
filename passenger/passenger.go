package main

import (
	"flag"
	"path/filepath"
	"runtime"

	"onlinecar-passenger/internal/config"
	passengerserver "onlinecar-passenger/internal/server/passenger"
	"onlinecar-passenger/internal/svc"
	"onlinecar-passenger/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func main() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	configFile := flag.String("f", filepath.Join(dir, "etc", "passenger.yaml"), "the config file")
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.MustSetup(c.Log)
	logx.DisableStat()

	svcCtx := svc.NewServiceContext(c)

	srv := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterPassengerServer(grpcServer, passengerserver.NewPassengerServer(svcCtx))
	})
	defer srv.Stop()

	logx.Infof("Starting passenger rpc server at %s...", c.ListenOn)
	srv.Start()
}
