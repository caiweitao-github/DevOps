package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Type interface {
	~bool | ~[]string | ~int | ~string | ~[]Result | ~[]OrderDtat | ~[]SfpsData | NginxForbidData | ReportIP
}

type Response[T Type] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data *T     `json:"data,omitempty"`
}

var (
	InvalidArgs     = err[int](400, "invalid params")
	Unauthorized    = err[int](401, "unauthorized")
	Forbidden       = err[int](403, "forbidden")
	NotFound        = err[int](404, "not found")
	TooManyRequests = err[int](429, "too many requests")
	ResultError     = err[int](500, "response error")
	ServerError     = err[int](500, "server error")

	Ok = success[int](200, "success!")
)

func resp[T Type](c *gin.Context, code int, msg string, data T) {
	c.JSON(http.StatusOK, Response[T]{
		code,
		msg,
		&data,
	})
}

func err[T Type](code int, msg string) Response[T] {
	return Response[T]{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

func success[T Type](code int, msg string) Response[T] {
	return Response[T]{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

func OK[T Type](c *gin.Context, data T) {
	OkWithMsg(c, Ok.Msg, data)
}

func OkWithMsg[T Type](c *gin.Context, msg string, data T) {
	resp(c, Ok.Code, msg, data)
}

func Fail[T Type](c *gin.Context, err Response[int]) {
	var data T
	resp(c, err.Code, err.Msg, data)
}

func FailWithMsg[T Type](c *gin.Context, err Response[int], msg string) {
	var data T
	resp(c, err.Code, msg, data)
}
