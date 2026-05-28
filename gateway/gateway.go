package main

import (
	"context"
	"flag"
	"path/filepath"
	"runtime"

	"onlinecar-gateway/internal/config"
	"onlinecar-gateway/internal/handler"
	"onlinecar-gateway/internal/middleware"
	"onlinecar-gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

func main() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	configFile := flag.String("f", filepath.Join(dir, "etc", "gateway.yaml"), "the config file")
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.MustSetup(c.Log)
	logx.DisableStat()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	defer ctx.Close()
	defer middleware.StopRateLimiters()

	ctx.ProvinceCache.Start(context.Background())

	handler.RegisterHandlers(server, ctx)

	logx.Infof("Starting gateway at %s:%d...", c.Host, c.Port)
	server.Start()
}
