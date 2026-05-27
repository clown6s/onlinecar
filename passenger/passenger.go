package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"onlinecar-passenger/internal/config"
	passengerserver "onlinecar-passenger/internal/server/passenger"
	"onlinecar-passenger/internal/svc"
	"onlinecar-passenger/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
	// 设置服务端参数
	srv.AddOptions(
		// 限制最大消息大小为 4MB
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
		// UnaryInterceptor：统一日志、panic recovery、内部认证
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			startTime := time.Now()
			defer func() {
				if r := recover(); r != nil {
					logx.Errorf("[gRPC] panic recovered in %s: %v", info.FullMethod, r)
					err = status.Errorf(codes.Internal, "internal server error")
				}
			}()

			// 内部认证：校验 x-internal-token
			if !strings.HasPrefix(info.FullMethod, "/grpc.health.v1.Health/") {
				internalToken := os.Getenv("ONLINECAR_INTERNAL_TOKEN")
				if internalToken == "" {
					logx.Severef("[gRPC] ONLINECAR_INTERNAL_TOKEN 未设置，拒绝所有非健康检查请求")
					return nil, status.Errorf(codes.Unauthenticated, "internal token not configured")
				}
				md, ok := metadata.FromIncomingContext(ctx)
				if !ok {
					return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
				}
				tokens := md.Get("x-internal-token")
				if len(tokens) == 0 || tokens[0] != internalToken {
					return nil, status.Errorf(codes.Unauthenticated, "invalid internal token")
				}
			}

			resp, err = handler(ctx, req)
			duration := time.Since(startTime)

			if err != nil {
				code := status.Code(err)
				logx.Errorf("[gRPC] %s 耗时=%v code=%s err=%v",
					info.FullMethod, duration, code.String(), err)
			} else {
				logx.Infof("[gRPC] %s 耗时=%v code=%s",
					info.FullMethod, duration, codes.OK.String())
			}

			return resp, err
		}),
		// StreamInterceptor：panic recovery、内部认证
		grpc.StreamInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logx.Errorf("[gRPC] panic recovered in stream %s: %v", info.FullMethod, r)
					err = status.Errorf(codes.Internal, "internal server error")
				}
			}()

			// 内部认证：校验 x-internal-token
			if !strings.HasPrefix(info.FullMethod, "/grpc.health.v1.Health/") {
				internalToken := os.Getenv("ONLINECAR_INTERNAL_TOKEN")
				if internalToken == "" {
					logx.Severef("[gRPC] ONLINECAR_INTERNAL_TOKEN 未设置，拒绝所有非健康检查请求")
					return status.Errorf(codes.Unauthenticated, "internal token not configured")
				}
				md, ok := metadata.FromIncomingContext(ss.Context())
				if !ok {
					return status.Errorf(codes.Unauthenticated, "missing metadata")
				}
				tokens := md.Get("x-internal-token")
				if len(tokens) == 0 || tokens[0] != internalToken {
					return status.Errorf(codes.Unauthenticated, "invalid internal token")
				}
			}

			return handler(srv, ss)
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

	logx.Infof("Starting passenger rpc server at %s...", c.ListenOn)
	srv.Start()
}
