package callboard

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	MaxPendingDeviceCount = 100 // 最大等待设备数量
	BindCodeTTL           = 300 // 绑定码有效时间，单位：秒
)

// 设置设备绑定码，使用 lua 脚本,保证原子性
// 参数：deviceId, bindCode, ttl，注意顺序
// 成功时返回：{bindCode, expireTime}
// 失败时抛出 Redis 错误：BIND_CODE_CONFLICT, TOO_MANY_DEVICES

const LuaScriptSetDeviceBindCode = `
local deviceId = ARGV[1]
local bindCode = ARGV[2]
local ttl = tonumber(ARGV[3])
local maxPendingDeviceCount = tonumber(ARGV[4])

local pendingKey = KEYS[1]
local deviceKey = KEYS[2]
local bindCodeKey = KEYS[3]
local currentTime = tonumber(redis.call('TIME')[1])

-- 1. 检查设备是否已申请过绑定码
local existingScore = tonumber(redis.call('ZSCORE', pendingKey, deviceId))
if existingScore and existingScore > currentTime then
    -- 设备未过期，直接返回现有绑定码
    local existingBindCode = redis.call('GET', deviceKey)
    if existingBindCode then
        return {existingBindCode, existingScore}
    end
end

-- 2. 检查新绑定码是否重复
if redis.call('EXISTS', bindCodeKey) == 1 then
    return redis.error_reply('BIND_CODE_CONFLICT')
end

-- 3. 清理过期设备并检查数量
redis.call('ZREMRANGEBYSCORE', pendingKey, '-inf', currentTime)
local count = redis.call('ZCARD', pendingKey)
if count >= maxPendingDeviceCount then
    return redis.error_reply('TOO_MANY_DEVICES')
end

-- 4. 创建新的绑定申请
local expireTime = currentTime + ttl
redis.call('ZADD', pendingKey, expireTime, deviceId)
redis.call('SETEX', deviceKey, ttl, bindCode)
redis.call('SETEX', bindCodeKey, ttl, deviceId)

return {bindCode, expireTime}
`

const LuaScriptClearPendingDevice = `
local pendingKey = KEYS[1]
local deviceKey = KEYS[2]
local bindCodeKey = KEYS[3]
local deviceId = ARGV[1]

redis.call('DEL', deviceKey)
redis.call('DEL', bindCodeKey)
redis.call('ZREM', pendingKey, deviceId)

return true
`

var (
	ErrBindCodeConflict = errors.New("bind_code_conflict")
	ErrTooManyDevices   = errors.New("too_many_devices")
)

// ParseScriptError 解析 Lua 脚本错误并转换为业务错误
func ParseScriptError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Redis 脚本相关错误
	if strings.Contains(errStr, "NOSCRIPT") {
		return ErrBindCodeConflict
	}

	// 业务错误
	if strings.Contains(errStr, "BIND_CODE_CONFLICT") {
		return ErrBindCodeConflict
	}
	if strings.Contains(errStr, "TOO_MANY_DEVICES") {
		return ErrTooManyDevices
	}

	return err
}

// ParseBindCodeResult 解析 Lua 脚本返回的绑定码结果
// 期望格式：[bindCode, expireTime]
func ParseBindCodeResult(result interface{}) (string, int64, error) {

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) != 2 {
		return "", 0, errors.New("Lua脚本返回值格式错误")
	}
	bindCode, ok := resultSlice[0].(string)
	if !ok {
		return "", 0, errors.New("绑定码类型错误")
	}

	// expireTime 在 Redis 中返回为 string
	expireTimeStr, ok := resultSlice[1].(int64)
	if !ok {
		return "", 0, errors.New("过期时间类型错误")
	}

	expireTime, err := strconv.ParseInt(strconv.FormatInt(expireTimeStr, 10), 10, 64)
	if err != nil {
		return "", 0, errors.Wrap(err, "过期时间转换失败")
	}

	return bindCode, expireTime, nil
}
