package health

import (
	"context"
	"net/http"
	"time"

	"onlinecar-gateway/common/response"
	"onlinecar-gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
			details["driver_rpc"] = "ok"
		} else {
			details["driver_rpc"] = "unhealthy"
		}

		passengerOk := checkGRPCHealth(svcCtx.PassengerConn, "passenger.rpc")
		if passengerOk {
			details["passenger_rpc"] = "ok"
		} else {
			details["passenger_rpc"] = "unhealthy"
		}

		status := "ok"
		if !driverOk || !passengerOk {
			status = "degraded"
		}

		code := http.StatusOK
		if status == "degraded" {
			code = http.StatusServiceUnavailable
		}

		if status == "degraded" {
			response.SuccessWithCode(w, code, HealthResponse{Status: status, Details: details})
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
