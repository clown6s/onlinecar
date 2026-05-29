package grpcx

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	grpcRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "onlinecar",
		Subsystem: "grpc",
		Name:      "requests_total",
		Help:      "Total number of gRPC requests",
	}, []string{"method", "code"})

	grpcRequestDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "onlinecar",
		Subsystem: "grpc",
		Name:      "request_duration_seconds",
		Help:      "gRPC request duration in seconds",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"method"})
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

// UnaryInterceptor 返回 gRPC Unary 拦截器（统一超时 + 统一日志 + panic recovery + 业务错误码透传）
func UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now()
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("[gRPC] panic recovered in %s: %v", info.FullMethod, r)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		// 服务端超时保护：如果客户端未设置 deadline，默认 10 秒超时
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
		}

		resp, err = handler(ctx, req)
		duration := time.Since(startTime)
		code := codes.OK

		if err != nil {
			code = status.Code(err)
			// 区分客户端错误（Info）和服务端错误（Error）
			switch code {
			case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
				codes.PermissionDenied, codes.Unauthenticated:
				logx.Infof("[gRPC] %s 耗时=%v code=%s err=%v",
					info.FullMethod, duration, code.String(), err)
			default:
				logx.Errorf("[gRPC] %s 耗时=%v code=%s err=%v",
					info.FullMethod, duration, code.String(), err)
			}
		} else {
			logx.Infof("[gRPC] %s 耗时=%v code=%s",
				info.FullMethod, duration, codes.OK.String())
		}

		grpcRequestsTotal.WithLabelValues(info.FullMethod, code.String()).Inc()
		grpcRequestDurationSeconds.WithLabelValues(info.FullMethod).Observe(duration.Seconds())

		return resp, err
	}
}

// StartMetricsServer 在后台启动 Prometheus metrics HTTP 端点
// port 为 HTTP 端口号（通常与 gRPC 端口不同），服务关闭时通过 cancel 停止。
func StartMetricsServer(port int, serviceName string) (cancel context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:    net.JoinHostPort("0.0.0.0", strconv.Itoa(port)),
		Handler: mux,
	}

	go func() {
		logx.Infof("[%s] metrics server listening on :%d/metrics", serviceName, port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.Errorf("[%s] metrics server error: %v", serviceName, err)
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
	}()

	return cancel
}
