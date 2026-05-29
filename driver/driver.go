package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"onlinecar-common/grpcx"
	"onlinecar-driver/internal/config"
	driverserver "onlinecar-driver/internal/server/driver"
	"onlinecar-driver/internal/svc"
	"onlinecar-driver/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

func main() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	configFile := flag.String("f", filepath.Join(dir, "etc", "driver.yaml"), "the config file")
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	c.DataSource = os.ExpandEnv(c.DataSource)

	logx.MustSetup(c.Log)
	logx.DisableStat()

	svcCtx := svc.NewServiceContext(c)

	// 后台 goroutine：定时清理过期的司机位置数据
	cleanCtx, cleanCancel := context.WithCancel(context.Background())
	defer cleanCancel()
	go cleanExpiredDriverLocations(cleanCtx, svcCtx)

	srv := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterDriverServer(grpcServer, driverserver.NewDriverServer(svcCtx))
		grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())
	})
	// 启动 Prometheus metrics HTTP 端点
	defer grpcx.StartMetricsServer(c.MetricsPort, "driver.rpc")()
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
		grpc.StreamInterceptor(grpcx.StreamInterceptor()),
		// UnaryInterceptor：统一日志、panic recovery、业务错误码透传
		grpc.UnaryInterceptor(grpcx.UnaryInterceptor()),
	)
	defer srv.Stop()

	// 优雅关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logx.Info("收到关闭信号，开始优雅关闭...")
		srv.Stop()
		logx.Info("gRPC 服务已停止")
		os.Exit(0)
	}()

	logx.Infof("Starting driver rpc server at %s...", c.ListenOn)
	srv.Start()
}

// driverLocationExpireThreshold 司机位置数据过期阈值（24 小时）
const driverLocationExpireThreshold = 24 * time.Hour

// removeExpiredLocationScript 原子移除过期司机的位置数据和更新时间戳
// 使用 DEL（per-driver key）而非 HDEL（hash field）
const removeExpiredLocationScript = `
redis.call('ZREM', KEYS[1], ARGV[1])
redis.call('DEL', KEYS[2])
return 1
`

// cleanExpiredDriverLocations 每小时清理超过 24 小时未更新位置的司机数据
// 使用 SCAN 遍历 driver:last_update:* per-driver key，而非 HSCAN 单个 hash。
// 每个 key 独立 TTL，活跃司机不会影响不活跃司机数据的过期。
func cleanExpiredDriverLocations(ctx context.Context, svcCtx *svc.ServiceContext) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	locationKey := "driver:location"
	expireThreshold := int64(driverLocationExpireThreshold.Seconds())

	for {
		select {
		case <-ctx.Done():
			logx.Info("司机位置清理 goroutine 已退出")
			return
		case <-ticker.C:
			if svcCtx.Redis == nil {
				logx.Error("Redis 未初始化，跳过司机位置清理")
				continue
			}
			logx.Info("开始清理过期司机位置数据...")

			var cursor uint64
			for {
				// SCAN 遍历所有 driver:last_update:* key
				keys, nextCursor, err := svcCtx.Redis.ScanCtx(ctx, cursor, "driver:last_update:*", 100)
				if err != nil {
					logx.Errorf("SCAN driver:last_update:* 失败: %v", err)
					break
				}
				cursor = nextCursor

				now := time.Now().Unix()
				for _, key := range keys {
					// GET per-driver 更新时间戳
					val, err := svcCtx.Redis.Get(key)
					if err != nil {
						logx.Errorf("GET %s 失败: %v", key, err)
						continue
					}
					var lastUpdate int64
					if _, err := fmt.Sscanf(val, "%d", &lastUpdate); err != nil {
						continue
					}
					if now-lastUpdate > expireThreshold {
						// 从 key 中提取 driverId（格式: driver:last_update:{id}）
						driverId := key[len("driver:last_update:"):]
						if _, err := svcCtx.Redis.Eval(removeExpiredLocationScript, []string{locationKey, key}, driverId); err != nil {
							logx.Errorf("清理过期司机位置失败, driverId=%s, err=%v", driverId, err)
							continue
						}
						logx.Infof("清理过期司机位置, driverId=%s", driverId)
					}
				}

				if cursor == 0 {
					break
				}
			}
		}
	}
}
