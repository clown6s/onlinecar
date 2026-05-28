package grpcx

import (
	"context"
	"time"

	"onlinecar-common/errorx"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamInterceptor 返回 gRPC Stream 拦截器（panic recovery）
func StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("[gRPC Stream] panic recovered in %s: %v", info.FullMethod, r)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}

// UnaryInterceptor 返回 gRPC Unary 拦截器（统一日志 + panic recovery + 业务错误码透传）
func UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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
	}
}
