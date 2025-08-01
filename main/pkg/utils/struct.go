package utils

import (
	"reflect"
	"strings"
)

// GetStructFieldsMap 获取结构体字段名称和json tag的映射
// 返回 map[string]string，key 是字段名，value 是对应的 json tag
func GetStructFieldsMap(obj any) map[string]string {
	result := make(map[string]string)

	// 获取结构体类型信息
	t := reflect.TypeOf(obj)

	// 如果是指针，获取其底层元素
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 确保是结构体
	if t.Kind() != reflect.Struct {
		return result
	}

	// 遍历所有字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 获取json/form/uri tag
		var tag string
		for _, tagKey := range []string{"json", "form", "uri"} {
			if tag = field.Tag.Get(tagKey); tag != "" {
				break
			}
		}
		if tag == "" {
			continue
		}

		// 处理tag中包含选项的情况（如 `json:"name,omitempty"`）
		tagParts := strings.Split(tag, ",")
		jsonName := tagParts[0]

		// 保存字段名和json tag的映射
		result[field.Name] = jsonName
	}

	return result
}

// GetStructFieldsMapRecursive 递归获取结构体（包括嵌套结构体）的字段名称和json tag的映射
func GetStructFieldsMapRecursive(obj any) map[string]string {
	result := make(map[string]string)

	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return result
	}

	var getFields func(t reflect.Type, prefix string)
	getFields = func(t reflect.Type, prefix string) {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// 处理嵌套结构体
			if field.Type.Kind() == reflect.Struct {
				// 如果字段是匿名的（嵌入的），不添加前缀
				newPrefix := prefix
				if !field.Anonymous {
					if prefix == "" {
						newPrefix = field.Name
					} else {
						newPrefix = prefix + "." + field.Name
					}
				}
				getFields(field.Type, newPrefix)
				continue
			}

			// 获取json/form/uri tag
			var tag string
			for _, tagKey := range []string{"json", "form", "uri"} {
				if tag = field.Tag.Get(tagKey); tag != "" {
					break
				}
			}
			if tag == "" {
				continue
			}

			// 处理tag中包含选项的情况
			tagParts := strings.Split(tag, ",")
			jsonName := tagParts[0]

			// 构建完整的字段名
			fieldName := field.Name
			if prefix != "" {
				fieldName = prefix + "." + fieldName
			}

			result[fieldName] = jsonName
		}
	}

	getFields(t, "")
	return result
}
