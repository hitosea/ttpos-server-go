package pkg

import (
	"bytes"
	"fmt"
	"sync"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// EncodingConverter 更好的编码转换器
type EncodingConverter struct {
	gbkEncoder   *encoding.Encoder
	gbkTransform transform.Transformer
	once         sync.Once
}

var (
	// 全局编码转换器实例
	globalConverter *EncodingConverter
	converterOnce   sync.Once
)

// GetConverter 获取全局编码转换器实例（单例模式）
func GetConverter() *EncodingConverter {
	converterOnce.Do(func() {
		globalConverter = &EncodingConverter{}
		globalConverter.init()
	})
	return globalConverter
}

// init 初始化编码器
func (c *EncodingConverter) init() {
	c.gbkEncoder = simplifiedchinese.GBK.NewEncoder()
	c.gbkTransform = simplifiedchinese.GBK.NewEncoder()
}

// UTF8ToGBK 方案1：使用 transform.String（推荐）
// 优势：支持 IGNORE 模式，自动跳过不支持的字符，不会报错
func (c *EncodingConverter) UTF8ToGBK(text string) ([]byte, error) {
	if text == "" {
		return []byte{}, nil
	}

	// 使用 transform.String，自动处理不支持的字符
	result, _, err := transform.String(c.gbkTransform, text)
	if err != nil {
		// 如果还是出错，使用 IGNORE 模式
		return c.UTF8ToGBKIgnore(text)
	}

	return []byte(result), nil
}

// UTF8ToGBKIgnore 方案2：忽略不支持的字符
// 优势：绝不报错，会跳过无法转换的字符
func (c *EncodingConverter) UTF8ToGBKIgnore(text string) ([]byte, error) {
	var result bytes.Buffer

	for _, r := range text {
		// 尝试转换单个字符
		charBytes, _, err := transform.Bytes(c.gbkTransform, []byte(string(r)))
		if err == nil {
			result.Write(charBytes)
		} else {
			// 转换失败，使用替代字符或跳过
			// 可以用 '?' 替代，或者直接跳过
			result.WriteByte('?')
		}
	}

	return result.Bytes(), nil
}

// UTF8ToGBKSafe 方案3：安全转换（最推荐）
// 优势：结合多种策略，确保转换成功
func (c *EncodingConverter) UTF8ToGBKSafe(text string) ([]byte, error) {
	if text == "" {
		return []byte{}, nil
	}

	// 方法1：直接转换
	if result, err := c.UTF8ToGBK(text); err == nil {
		return result, nil
	}

	// 方法2：逐字符转换
	var result bytes.Buffer
	for _, r := range text {
		char := string(r)

		// 先检查是否是ASCII字符
		if r < 128 {
			result.WriteByte(byte(r))
			continue
		}

		// 尝试转换中文字符
		if charBytes, err := c.UTF8ToGBK(char); err == nil && len(charBytes) > 0 {
			result.Write(charBytes)
		} else {
			// 使用备用方案
			if backup := c.getCharacterBackup(r); backup != nil {
				result.Write(backup)
			} else {
				// 最后使用占位符
				result.WriteByte('?')
			}
		}
	}

	return result.Bytes(), nil
}

// getCharacterBackup 获取字符的备用编码
func (c *EncodingConverter) getCharacterBackup(r rune) []byte {
	// 常见字符的GBK备用编码映射
	backupMap := map[rune][]byte{
		'？': {0xA3, 0xBF}, // 全角问号
		'！': {0xA3, 0xA1}, // 全角感叹号
		'（': {0xA3, 0xA8}, // 全角左括号
		'）': {0xA3, 0xA9}, // 全角右括号
		'，': {0xA3, 0xAC}, // 全角逗号
		'。': {0xA3, 0xAE}, // 全角句号
		'：': {0xA3, 0xBA}, // 全角冒号
		'；': {0xA3, 0xBB}, // 全角分号
	}

	if backup, exists := backupMap[r]; exists {
		return backup
	}

	return nil
}

// UTF8ToGBKWithFallback 方案4：带回退机制的转换
func (c *EncodingConverter) UTF8ToGBKWithFallback(text string) ([]byte, error) {
	// 首先尝试标准转换
	result, _, err := transform.String(c.gbkTransform, text)
	if err == nil {
		return []byte(result), nil
	}

	// 如果失败，使用字符替换策略
	return c.convertWithReplacement(text), nil
}

// convertWithReplacement 使用字符替换的转换方法
func (c *EncodingConverter) convertWithReplacement(text string) []byte {
	var result bytes.Buffer

	for _, r := range text {
		// ASCII字符直接保留
		if r < 128 {
			result.WriteByte(byte(r))
			continue
		}

		// 尝试转换中文字符
		char := string(r)
		if gbkBytes, err := c.gbkEncoder.Bytes([]byte(char)); err == nil {
			result.Write(gbkBytes)
		} else {
			// 转换失败，使用相似字符替换
			replacement := c.findSimilarCharacter(r)
			if replacement != nil {
				result.Write(replacement)
			} else {
				// 使用问号替代
				result.WriteByte('?')
			}
		}
	}

	return result.Bytes()
}

// findSimilarCharacter 查找相似字符
func (c *EncodingConverter) findSimilarCharacter(r rune) []byte {
	// 一些生僻字的常用替代字符
	replacements := map[rune]string{
		'龘': "龙", // 生僻的龙字
		'鱻': "鲜", // 生僻的鲜字
		'麤': "粗", // 生僻的粗字
		'厤': "历", // 生僻的历字
	}

	if replacement, exists := replacements[r]; exists {
		if gbkBytes, err := c.gbkEncoder.Bytes([]byte(replacement)); err == nil {
			return gbkBytes
		}
	}

	return nil
}

// 便捷函数：直接转换UTF-8字符串到GBK字节
func UTF8ToGBKBytes(text string) ([]byte, error) {
	return GetConverter().UTF8ToGBKSafe(text)
}

// 便捷函数：转换UTF-8字符串到GBK，忽略错误
func UTF8ToGBKIgnoreErrors(text string) []byte {
	result, _ := GetConverter().UTF8ToGBKSafe(text)
	return result
}

// 性能优化：批量转换
func BatchUTF8ToGBK(texts []string) ([][]byte, error) {
	converter := GetConverter()
	results := make([][]byte, len(texts))

	for i, text := range texts {
		if gbkBytes, err := converter.UTF8ToGBKSafe(text); err != nil {
			return nil, fmt.Errorf("转换第%d个文本失败: %v", i, err)
		} else {
			results[i] = gbkBytes
		}
	}

	return results, nil
}

// BetterEncodeTo 改进的编码转换函数，替代原来的 encodeTo
func BetterEncodeTo(text string, encoding string) ([]byte, error) {
	switch encoding {
	case "gbk", "GB2312":
		// 使用改进的GBK转换
		return UTF8ToGBKBytes(text)
	case "utf8", "UTF-8":
		// 直接返回UTF-8字节
		return []byte(text), nil
	default:
		// 对于其他编码，使用原来的方法
		return encodeTo(text, encoding)
	}
}

// 测试函数：对比不同转换方法的效果
func TestEncodingMethods(text string) {
	converter := GetConverter()

	fmt.Printf("原始文本: %s\n", text)

	// 方法1：原来的方法
	if oldResult, err := encodeTo(text, "gbk"); err != nil {
		fmt.Printf("原方法失败: %v\n", err)
	} else {
		fmt.Printf("原方法结果: %v\n", oldResult)
	}

	// 方法2：新的安全方法
	if newResult, err := converter.UTF8ToGBKSafe(text); err != nil {
		fmt.Printf("新方法失败: %v\n", err)
	} else {
		fmt.Printf("新方法结果: %v\n", newResult)
	}

	// 方法3：忽略错误方法
	ignoreResult := UTF8ToGBKIgnoreErrors(text)
	fmt.Printf("忽略错误方法结果: %v\n", ignoreResult)
}
