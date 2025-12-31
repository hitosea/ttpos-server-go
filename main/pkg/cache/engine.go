package cache

import (
	"context"

	"golang.org/x/sync/singleflight"
)

// singleflightEngine 提供了对 golang.org/x/sync/singleflight 的泛型封装
// 确保高并发下相同 Key 的请求只会被执行一次
type singleflightEngine[T any] struct {
	group singleflight.Group
}

// newSingleflightEngine 创建一个新的 Singleflight 引擎实例
func newSingleflightEngine[T any]() *singleflightEngine[T] {
	return &singleflightEngine[T]{}
}

// do 执行合并请求
// key: 并发合并的唯一标识
// fn: 实际要执行的业务函数
func (e *singleflightEngine[T]) do(ctx context.Context, key string, fn func(ctx context.Context) (T, error)) (T, error) {
	// 调用标准库的 singleflight.Group.Do
	// 返回值 val 是 interface{}，需要断言回泛型 T
	val, err, _ := e.group.Do(key, func() (interface{}, error) {
		return fn(ctx)
	})

	if err != nil {
		var zero T
		return zero, err
	}

	// 泛型断言，由于 fn 返回的是 T，这里断言一定是安全的
	return val.(T), nil
}
