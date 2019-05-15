package utils

import "net/http"

type HttpError struct {
	HttpCode int    `json:"-"`
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
}


func NewHttpError(httpCode, code int, msg string) HttpError {
	return HttpError{
		HttpCode: httpCode,
		Code:     code,
		Msg:      msg,
	}
}

func (h HttpError) Error() string {
	return h.Msg
}

var (
	ParamError = NewHttpError(http.StatusBadRequest, http.StatusBadRequest, "Parameter error")
)
