package cachekey

import (
	"strconv"
)

// 叫号相关
const (
	PendingDeviceListKey     = "{binding}:pending_devices"
	pendingDeviceKeyPrefix   = "{binding}:pending_device:"
	bindCodeKeyPrefix        = "{binding}:bind_code:"
	bindedDeviceKeyPrefix    = "bind:device:"        //绑定设备信息， bind:device:<deviceId>
	callBoardChangeKeyPrefix = "callboard_change:"   //叫号展示变化信息， callboard_change:<companyUuid>
	boardLanguageKeyPrefix   = "callboard_language:" //叫号语言信息， callboard_language:<companyUuid>
)

// 叫号相关
func GetPendingDeviceKey(deviceId string) string {
	return pendingDeviceKeyPrefix + deviceId
}

func GetBindCodeKey(bindCode string) string {
	return bindCodeKeyPrefix + bindCode
}

func GetBindedDeviceKey(deviceId string) string {
	return bindedDeviceKeyPrefix + deviceId
}

func GetCallBoardChangeKey(companyUuid uint64) string {
	return callBoardChangeKeyPrefix + strconv.FormatUint(companyUuid, 10)
}

func GetBoardLanguageKey(companyUuid uint64) string {
	return boardLanguageKeyPrefix + strconv.FormatUint(companyUuid, 10)
}
