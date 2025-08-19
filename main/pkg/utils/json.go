package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ToJsonString 结构体转json
func ToJsonString(data any) string {
	jsonBytes, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

func FromJson(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}

func ToJson(data any) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

// JsonToStruct 对象转结构体
func JsonToStruct(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}

// StrToStruct 字符串转对象
func StrToStruct(source string, destination any) error {
	err := json.Unmarshal([]byte(source), destination)
	return err
}

// StrToMap 字符转Map
func StrToMap(source string) (map[string]any, error) {
	res := make(map[string]any)
	err := json.Unmarshal([]byte(source), &res)
	return res, err
}

// MapToStruct Map转结构体
func MapToStruct(m map[string]any, s any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	return nil
}

// StructToStruct 结构体转结构体，并可选择过滤某些字段
func StructToStruct(data any, newdata any, filtrationKeys ...string) error {
	result, err := json.Marshal(data)
	if err != nil {
		return err
	}
	str := string(result)
	// 过滤字段
	if len(filtrationKeys) != 0 {
		keysArr := strings.Split(filtrationKeys[0], ",")
		for _, v := range keysArr {
			str = strings.Replace(string(str), `"`+v+`":`, `"_`+v+`":`, 1)
		}
	}
	// 转新的结构体
	err = json.Unmarshal([]byte(str), newdata)
	if err != nil {
		return err
	}
	//
	return nil
}

// JsonToStr 对象转换成字符串
func JsonToStr(data any) string {
	result, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(result)
}

// NumToStr 数字转字符串
func NumToStr(num any) string {
	return fmt.Sprintf("%v", num)
}

// StructToMap 结构体转map
func StructToMap(data any) (map[string]any, error) {
	var result map[string]any
	result, err := StrToMap(ToJsonString(data))
	return result, err
}

// MergeMaps 合并map
func MergeMaps(maps ...map[string]any) map[string]any {
	result := make(map[string]any)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
