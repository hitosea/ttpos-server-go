package rpc

import (
	"ttpos-bmp/app/ttpos-erp/api"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// ApiSuccess 成功响应
// 参数：msg 成功消息
// 返回：标准成功响应
func ApiSuccess(msg string) *api.ResponseInfo {
	return &api.ResponseInfo{
		Code:    "0",
		Message: msg,
	}
}

// ApiSuccessWithData 成功响应带数据
// 参数：msg 成功消息，data 成功数据
// 返回：标准成功响应
func ApiSuccessWithData(msg string, data proto.Message) *api.ResponseInfo {
	resp := &api.ResponseInfo{
		Code:    "0",
		Message: msg,
	}

	if data != nil {
		if anyData, err := anypb.New(data); err == nil {
			resp.Data = anyData
		}
	}

	return resp
}

// ApiError 错误响应
// 参数：msg 错误消息
// 返回：标准错误响应
func ApiError(msg string) *api.ResponseInfo {
	return &api.ResponseInfo{
		Code:    "1",
		Message: msg,
	}
}

// ApiErrorWithData 带数据的错误响应
// 参数：msg 错误消息，data 错误数据
// 返回：标准错误响应
func ApiErrorWithData(msg string, data proto.Message) *api.ResponseInfo {
	resp := &api.ResponseInfo{
		Code:    "1",
		Message: msg,
	}

	if data != nil {
		if anyData, err := anypb.New(data); err == nil {
			resp.Data = anyData
		}
	}

	return resp
}
