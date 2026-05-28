package health

import (
	"context"
	"net/http"
	"time"

	"onlinecar-gateway/common/response"
	"onlinecar-gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type HealthResponse struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details"`
}

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		details := make(map[string]string)

		driverOk := checkGRPCHealth(svcCtx.DriverConn, "driver.rpc")
		if driverOk {
			details["service_a"] = "ok"
		} else {
			details["service_a"] = "unhealthy"
		}

		passengerOk := checkGRPCHealth(svcCtx.PassengerConn, "passenger.rpc")
		if passengerOk {
			details["service_b"] = "ok"
		} else {
			details["service_b"] = "unhealthy"
		}

		districtOk := checkGRPCHealth(svcCtx.DistrictConn, "district.rpc")
		if districtOk {
			details["service_c"] = "ok"
		} else {
			details["service_c"] = "unhealthy"
		}

		redisOk := checkRedisHealth(svcCtx.Redis)
		if redisOk {
			details["redis"] = "ok"
		} else {
			details["redis"] = "unhealthy"
		}

		status := "ok"
		if !driverOk || !passengerOk || !districtOk || !redisOk {
			status = "degraded"
		}

		if status == "degraded" {
			// 降级时仅返回状态和通用提示，不暴露后端服务拓扑
			response.Error(w, http.StatusServiceUnavailable, 50301, "service degraded")
		} else {
			response.Success(w, HealthResponse{Status: status, Details: details})
		}
	}
}

func checkGRPCHealth(conn *grpc.ClientConn, service string) bool {
	if conn == nil {
		return false
	}

	client := grpc_health_v1.NewHealthClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
	if err != nil {
		logx.Errorf("health check failed for %s: %v", service, err)
		return false
	}

	return resp.GetStatus() == grpc_health_v1.HealthCheckResponse_SERVING
}

func checkRedisHealth(rdb *redis.Redis) bool {
	if rdb == nil {
		return false
	}
	ok := rdb.Ping()
	if !ok {
		logx.Errorf("redis health check failed")
		return false
	}
	return true
}
