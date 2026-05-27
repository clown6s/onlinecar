package response

import (
	"errors"
	"net/http"

	"onlinecar-gateway/common/errorx"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Body struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func Response(w http.ResponseWriter, resp any, err error) {
	var body Body
	if err != nil {
		var codeErr *errorx.CodeError
		if errors.As(err, &codeErr) {
			body.Code = codeErr.Code
			body.Msg = codeErr.Msg
			httpx.WriteJson(w, codeErr.HttpCode(), body)
		} else {
			body.Code = -1
			body.Msg = "服务器内部错误，请稍后重试"
			httpx.WriteJson(w, http.StatusInternalServerError, body)
		}
		return
	}
	body.Code = 0
	body.Msg = "ok"
	body.Data = resp
	httpx.OkJson(w, body)
}

func Success(w http.ResponseWriter, data any) {
	httpx.OkJson(w, Body{Code: 0, Msg: "ok", Data: data})
}

// SuccessWithCode 返回成功响应但允许自定义 HTTP 状态码（如 503 degraded）
func SuccessWithCode(w http.ResponseWriter, httpCode int, data any) {
	httpx.WriteJson(w, httpCode, Body{Code: 0, Msg: "ok", Data: data})
}

func Error(w http.ResponseWriter, httpCode int, code int, msg string) {
	httpx.WriteJson(w, httpCode, Body{Code: code, Msg: msg})
}

func Errorx(w http.ResponseWriter, err *errorx.CodeError) {
	httpx.WriteJson(w, err.HttpCode(), Body{Code: err.Code, Msg: err.Msg})
}
