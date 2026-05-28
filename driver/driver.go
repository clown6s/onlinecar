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
	"google.golang.org/grpc/keepalive"
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

	// 后台 goroutine：定时清理过期的司机位置数据
	cleanCtx, cleanCancel := context.WithCancel(context.Background())
	defer cleanCancel()
	go cleanExpiredDriverLocations(cleanCtx, svcCtx)

	srv := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterDriverServer(grpcServer, driverserver.NewDriverServer(svcCtx))
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
const removeExpiredLocationScript = `
redis.call('ZREM', KEYS[1], ARGV[1])
redis.call('HDEL', KEYS[2], ARGV[1])
return 1
`

// cleanExpiredDriverLocations 每小时清理超过 24 小时未更新位置的司机数据
func cleanExpiredDriverLocations(ctx context.Context, svcCtx *svc.ServiceContext) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	locationKey := "driver:location"
	lastUpdateKey := "driver:last_update"
	expireThreshold := int64(driverLocationExpireThreshold.Seconds())

	for {
		select {
		case <-ctx.Done():
			logx.Info("司机位置清理 goroutine 已退出")
			return
		case <-ticker.C:
			logx.Info("开始清理过期司机位置数据...")

			var cursor uint64
			for {
				result, nextCursor, err := svcCtx.Redis.HscanCtx(ctx, lastUpdateKey, cursor, "*", 100)
				if err != nil {
					logx.Errorf("HSCAN driver:last_update 失败: %v", err)
					break
				}
				cursor = nextCursor

				// result 是扁平数组：[field1, value1, field2, value2, ...]
				now := time.Now().Unix()
				for i := 0; i+1 < len(result); i += 2 {
					driverId := result[i]
					var lastUpdate int64
					if _, err := fmt.Sscanf(result[i+1], "%d", &lastUpdate); err != nil {
						continue
					}
					if now-lastUpdate > expireThreshold {
						// 使用 Lua 脚本原子移除过期司机的位置数据和更新时间戳
						if _, err := svcCtx.Redis.Eval(removeExpiredLocationScript, []string{locationKey, lastUpdateKey}, driverId); err != nil {
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
