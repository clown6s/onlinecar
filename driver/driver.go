package main

import (
	"flag"
	"path/filepath"
	"runtime"

	"onlinecar-driver/internal/config"
	driverserver "onlinecar-driver/internal/server/driver"
	"onlinecar-driver/internal/svc"
	"onlinecar-driver/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func main() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	configFile := flag.String("f", filepath.Join(dir, "etc", "driver.yaml"), "the config file")
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.MustSetup(c.Log)
	logx.DisableStat()

	svcCtx := svc.NewServiceContext(c)

	srv := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterDriverServer(grpcServer, driverserver.NewDriverServer(svcCtx))
	})
	defer srv.Stop()

	logx.Infof("Starting driver rpc server at %s...", c.ListenOn)
	srv.Start()
}
