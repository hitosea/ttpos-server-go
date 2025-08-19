// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
)

type (
	ITakeout interface {
		Get(ctx context.Context, shopOrderUuid string) (*entity.Job, error)
	}
)

var (
	localTakeout ITakeout
)

func Takeout() ITakeout {
	if localTakeout == nil {
		panic("implement not found for interface ITakeout, forgot register?")
	}
	return localTakeout
}

func RegisterTakeout(i ITakeout) {
	localTakeout = i
}
