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
	result, err := StrToMap(ToJson(data))
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

// ExtractNestedFieldToStruct 从 JSON 字符串中提取指定的嵌套字段并解析到目标结构体
// jsonStr: JSON 字符串
// fieldName: 要提取的字段名
// target: 目标结构体指针
// 返回错误信息
func ExtractNestedFieldToStruct(jsonStr string, fieldName string, target any) error {
	var dataMap map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &dataMap); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	fieldValue, ok := dataMap[fieldName]
	if !ok {
		return fmt.Errorf("字段 %s 不存在", fieldName)
	}

	// 将字段值序列化为 JSON，再解析到目标结构体
	fieldBytes, err := json.Marshal(fieldValue)
	if err != nil {
		return fmt.Errorf("序列化字段 %s 失败: %w", fieldName, err)
	}

	if err := json.Unmarshal(fieldBytes, target); err != nil {
		return fmt.Errorf("解析字段 %s 到结构体失败: %w", fieldName, err)
	}

	return nil
}
