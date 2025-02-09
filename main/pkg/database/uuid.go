package database

import (
	"fmt"

	"github.com/sony/sonyflake"
)

// 参考：https://jicki.cn/golang-web-note-10/
// todo: 验证id是否会重复，重启多次后id是否会重复；时间偏移是否会重复
var (
	sonyFlake *sonyflake.Sonyflake
)

// 初始化 sonyFlake 配置
func InitSonyFlakeId() {
	st := sonyflake.Settings{}
	sonyFlake = sonyflake.NewSonyflake(st)
	return
}

// 获取全局 Uuid 的函数
func GetID() (id uint64, err error) {
	if sonyFlake == nil {
		err = fmt.Errorf("需要先初始化以后再执行 GetID 函数 err: %#v \n", err)
		return
	}
	return sonyFlake.NextID()
}
