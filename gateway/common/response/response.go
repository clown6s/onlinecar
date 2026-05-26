package response

import (
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
		if codeErr, ok := err.(*errorx.CodeError); ok {
			body.Code = codeErr.Code
			body.Msg = codeErr.Msg
		} else {
			body.Code = -1
			body.Msg = err.Error()
		}
	} else {
		body.Code = 0
		body.Msg = "ok"
		body.Data = resp
	}
	httpx.OkJson(w, body)
}

func Success(w http.ResponseWriter, data any) {
	httpx.OkJson(w, Body{Code: 0, Msg: "ok", Data: data})
}

func Error(w http.ResponseWriter, code int, msg string) {
	httpx.OkJson(w, Body{Code: code, Msg: msg})
}
