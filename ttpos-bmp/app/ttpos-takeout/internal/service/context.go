// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/internal/model"
)

type (
	IContextService interface {
		Get(ctx context.Context) *model.Context
	}
)

var (
	localContextService IContextService
)

func ContextService() IContextService {
	if localContextService == nil {
		panic("implement not found for interface IContextService, forgot register?")
	}
	return localContextService
}

func RegisterContextService(i IContextService) {
	localContextService = i
}
