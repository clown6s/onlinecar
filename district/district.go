package main

import (
	"context"
	"crypto/subtle"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"time"

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
	"google.golang.org/grpc/metadata"
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
		// StreamInterceptor：panic recovery + 内部认证（与 UnaryInterceptor 保持一致）
		grpc.StreamInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logx.Errorf("[gRPC Stream] panic recovered in %s: %v", info.FullMethod, r)
					err = status.Errorf(codes.Internal, "internal server error")
				}
			}()

			// 内部认证：校验 x-internal-token
			isHealthCheck := info.FullMethod == "/grpc.health.v1.Health/Check" ||
				info.FullMethod == "/grpc.health.v1.Health/Watch"
			if !isHealthCheck {
				expected := os.Getenv("ONLINECAR_INTERNAL_TOKEN")
				if expected == "" {
					logx.Severef("[gRPC Stream] ONLINECAR_INTERNAL_TOKEN 未设置，拒绝所有非健康检查请求")
					return status.Errorf(codes.Unauthenticated, "internal token not configured")
				}
				md, ok := metadata.FromIncomingContext(ss.Context())
				if !ok {
					return status.Errorf(codes.Unauthenticated, "missing metadata")
				}
				tokens := md.Get("x-internal-token")
				if len(tokens) == 0 || subtle.ConstantTimeCompare([]byte(tokens[0]), []byte(expected)) != 1 {
					return status.Errorf(codes.Unauthenticated, "invalid internal token")
				}
			}

			return handler(srv, ss)
		}),
		// UnaryInterceptor：统一日志、panic recovery + 简单认证
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			// 简单认证：校验 x-internal-token（健康检查路径跳过）
			isHealthCheck := info.FullMethod == "/grpc.health.v1.Health/Check" ||
				info.FullMethod == "/grpc.health.v1.Health/Watch"
			if !isHealthCheck {
				expected := os.Getenv("ONLINECAR_INTERNAL_TOKEN")
				if expected == "" {
					logx.Severef("[gRPC] ONLINECAR_INTERNAL_TOKEN 未设置，拒绝所有非健康检查请求")
					return nil, status.Errorf(codes.Unauthenticated, "internal token not configured")
				}
				md, ok := metadata.FromIncomingContext(ctx)
				if !ok {
					return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
				}
				tokens := md.Get("x-internal-token")
				if len(tokens) == 0 || subtle.ConstantTimeCompare([]byte(tokens[0]), []byte(expected)) != 1 {
					return nil, status.Errorf(codes.Unauthenticated, "invalid internal token")
				}
			}

			// panic recovery
			var err error
			defer func() {
				if r := recover(); r != nil {
					logx.Errorf("[gRPC] panic recovered in %s: %v", info.FullMethod, r)
					err = status.Errorf(codes.Internal, "internal server error")
				}
			}()

			startTime := time.Now()
			resp, handlerErr := handler(ctx, req)
			err = handlerErr
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
	)
	defer srv.Stop()

	logx.Infof("Starting district rpc server at %s...", c.ListenOn)
	srv.Start()
}
