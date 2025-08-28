package utility

import "strings"

// ConvertToTitleCase 将下划线分隔的字符串转换为标题格式
// 参数：
//   - input: 输入字符串，如 "custom_field"
//
// 返回：
//   - string: 转换后的标题格式字符串，如 "Custom Field"
func ConvertToTitleCase(input string) string {
	// 参数验证
	if input == "" {
		return ""
	}

	// 按下划线分割字符串
	words := strings.Split(input, "_")

	// 将每个单词的首字母大写
	for i, word := range words {
		if len(word) > 0 {
			// 将单词转换为首字母大写，其余小写
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	// 用空格连接所有单词
	return strings.Join(words, " ")
}
