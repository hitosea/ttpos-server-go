package rpc

import (
	"ttpos-bmp/app/ttpos-erp/api"
)

func ApiSuccess(msg string) (resp *api.ResponseInfo) {
	resp = &api.ResponseInfo{
		Code:    "0",
		Message: msg,
	}
	return resp
}

func ApiError(msg string) (resp *api.ResponseInfo) {
	resp = &api.ResponseInfo{
		Code:    "1",
		Message: msg,
	}
	return resp
}
