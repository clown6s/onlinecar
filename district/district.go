package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"onlinecar-common/errorx"
	"onlinecar-district/internal/config"
	districtserviceServer "onlinecar-district/internal/server/districtservice"
	"onlinecar-district/internal/svc"
	"onlinecar-district/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

func main() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	configFile := flag.String("f", filepath.Join(dir, "etc", "district.yaml"), "the config file")
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.MustSetup(c.Log)
	logx.DisableStat()

	svcCtx := svc.NewServiceContext(c)

	srv := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterDistrictServiceServer(grpcServer, districtserviceServer.NewDistrictServiceServer(svcCtx))
	})
	// 设置服务端参数
	srv.AddOptions(
		// 限制最大消息大小为 4MB（默认值显式声明，可根据业务调整）
		grpc.MaxRecvMsgSize(4*1024*1024),
		grpc.MaxSendMsgSize(4*1024*1024),
		// Keepalive 参数
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			MaxConnectionAge:  30 * time.Minute,
			Time:              30 * time.Second,
			Timeout:           10 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             10 * time.Second,
			PermitWithoutStream: false,
		}),
		// StreamInterceptor：panic recovery
		grpc.StreamInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logx.Errorf("[gRPC Stream] panic recovered in %s: %v", info.FullMethod, r)
					err = status.Errorf(codes.Internal, "internal server error")
				}
			}()

			return handler(srv, ss)
		}),
		// UnaryInterceptor：统一日志、panic recovery
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			startTime := time.Now()
			defer func() {
				if r := recover(); r != nil {
					logx.Errorf("[gRPC] panic recovered in %s: %v", info.FullMethod, r)
					err = status.Errorf(codes.Internal, "internal server error")
				}
			}()

			resp, err = handler(ctx, req)
			duration := time.Since(startTime)

			if err != nil {
				// 将业务错误码写入 gRPC trailer，供 Gateway MapRpcError 提取
				if bizCode := errorx.ErrorToBizCode(err); bizCode > 0 {
					errorx.SetGrpcBizCode(ctx, bizCode)
				}
				code := status.Code(err)
				logx.Errorf("[gRPC] %s 耗时=%v code=%s err=%v",
					info.FullMethod, duration, code.String(), err)
			} else {
				logx.Infof("[gRPC] %s 耗时=%v code=%s",
					info.FullMethod, duration, codes.OK.String())
			}

			return resp, err
		}),
	)
	defer srv.Stop()

	// 优雅关闭：监听信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logx.Info("收到关闭信号，开始优雅关闭...")
		srv.Stop()
		logx.Info("gRPC 服务已停止")
		os.Exit(0)
	}()

	logx.Infof("Starting district rpc server at %s...", c.ListenOn)
	srv.Start()
}
