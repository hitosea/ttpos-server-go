// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"takeout/internal/model"
)

type (
	IEcho interface {
		Msg(ctx context.Context, in *model.EchoMsgInput) (out *model.EcoMsgOutput, err error)
	}
)

var (
	localEcho IEcho
)

func Echo() IEcho {
	if localEcho == nil {
		panic("implement not found for interface IEcho, forgot register?")
	}
	return localEcho
}

func RegisterEcho(i IEcho) {
	localEcho = i
}
