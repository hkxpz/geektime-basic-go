package handlefunc

import "github.com/golang-jwt/jwt/v5"

type UserClaims struct {
	ID        int64
	UserAgent string
	SSID      string
	jwt.RegisteredClaims
}

// Response 你可以通过在 Result 里面定义更加多的字段，来配合 Wrap 方法
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func InternalServerError() Response {
	return Response{Code: 5, Msg: "系统错误"}
}

func RespSuccess(msg string) Response {
	return Response{Code: 2, Msg: msg}
}

func BadRequestError(msg string) Response {
	return Response{Code: 4, Msg: msg}
}
