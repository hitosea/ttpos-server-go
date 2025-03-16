package pkg

import (
	"encoding/hex"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"unicode"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// PrintTicket 打印
func PrintTicket(printerIP string, content string, printMethod int) error {
	if printMethod == constant.Yes {
		return ExecutePrinting(printerIP, content)
	}
	return ExecuteImgPrinting(printerIP, content)
}

// 图片打印
func ExecuteImgPrinting(printerIP string, content string) error {
	// 连接打印机（TCP连接，端口9100是标准打印机端口）
	conn, err := net.DialTimeout("tcp", printerIP+":9100", 3*time.Second)
	if err != nil {
		return fmt.Errorf("连接打印机出错: %v", err)
	}
	defer conn.Close()

	// 转换十六进制字符串为二进制数据
	binaryContent := hex2bin(content)

	// 初始化打印机
	initCmd := []byte{0x1B, 0x40}
	_, err = conn.Write(initCmd)
	if err != nil {
		return fmt.Errorf("初始化打印机失败: %v", err)
	}

	// 发送打印数据
	_, err = conn.Write([]byte(binaryContent))
	if err != nil {
		return fmt.Errorf("发送打印数据失败: %v", err)
	}

	return nil
}

// ExecutePrinting 连接打印机并发送打印内容
// 支持多种语言（泰语、韩语、中文等）的字符编码处理
func ExecutePrinting(printerIP string, content string) error {
	// 连接打印机（TCP连接，端口9100是标准打印机端口）
	conn, err := net.DialTimeout("tcp", printerIP+":9100", 3*time.Second)
	if err != nil {
		return fmt.Errorf("连接打印机出错: %v", err)
	}
	defer conn.Close()

	// 转换十六进制字符串为二进制数据
	content = hex2bin(content)
	// 替换特殊字符
	content = strings.ReplaceAll(content, "ー", "-")
	// 使用正则表达式分割文本，保留泰语、韩语、泰铢符号
	thaiHangulRegex := regexp.MustCompile(`([\p{Thai}\p{Hangul}฿]+)`)
	segments := thaiHangulRegex.Split(content, -1)
	matches := thaiHangulRegex.FindAllString(content, -1)
	// 合并分割结果和匹配结果，按原始顺序
	allSegments := make([]string, 0)
	matchIndex := 0
	for _, seg := range segments {
		if seg != "" {
			allSegments = append(allSegments, seg)
		}
		if matchIndex < len(matches) {
			allSegments = append(allSegments, matches[matchIndex])
			matchIndex++
		}
	}
	// 处理每个文本段落，根据语言类型选择对应的编码
	for _, segment := range allSegments {
		if segment == "" {
			continue
		}
		// 检查是否包含泰语字符或泰铢符号
		isThai := containsThai(segment) || strings.Contains(segment, "฿")
		isKorean := containsKorean(segment)
		isTurkish := containsTurkish(segment)
		if isThai {
			// 泰语处理
			// 切换到泰语字符集
			_, err = conn.Write([]byte{0x1C, 0x2E})
			if err != nil {
				return err
			}
			// 转换为CP874编码（泰语字符集）
			encoded, err := encodeTo(segment, "cp874")
			if err != nil {
				return err
			}
			conn.Write(encoded)
		} else if isKorean {
			// 韩语处理
			// 切换到韩语字符集
			_, err = conn.Write([]byte{0x1C, 0x26})
			if err != nil {
				return err
			}
			// 转换为CP949编码（韩语字符集）
			encoded, err := encodeTo(segment, "cp949")
			if err != nil {
				return err
			}
			conn.Write(encoded)
		} else if isTurkish {
			// 土耳其语处理
			// 切换到土耳其语字符集
			_, err = conn.Write([]byte{0x1B, 0x74, 0x02}) // ESC t 2 - 选择PC857编码（土耳其语字符集）
			if err != nil {
				return err
			}
			// 转换为ISO-8859-9编码（土耳其语字符集）
			encoded, err := encodeTo(segment, "iso8859-9")
			if err != nil {
				return err
			}
			conn.Write(encoded)
		} else {
			// 其他语言（默认使用中文GBK编码）
			// 切换到中文字符集
			_, err = conn.Write([]byte{0x1C, 0x26})
			if err != nil {
				return err
			}
			// 转换为GBK编码（中文字符集）
			encoded, err := encodeTo(segment, "gbk")
			if err != nil {
				return err
			}
			conn.Write(encoded)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// 检查字符串是否包含泰语字符
func containsThai(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Thai, r) {
			return true
		}
	}
	return false
}

// 检查字符串是否包含韩语字符
func containsKorean(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Hangul, r) {
			return true
		}
	}
	return false
}

// 检查字符串是否包含土耳其语特有字符
func containsTurkish(s string) bool {
	turkishChars := []rune{'ğ', 'ı', 'İ', 'ö', 'ş', 'ü', 'Ç', 'ç', 'Ğ', 'Ö', 'Ş', 'Ü'} // 土耳其语特有字符: ğ, ı, İ, ö, ş, ü, Ç, ç, Ğ, Ö, Ş, Ü
	for _, r := range s {
		for _, tc := range turkishChars {
			if r == tc {
				return true
			}
		}
	}
	return false
}

// 将字符串转换为指定编码
func encodeTo(s string, encoding string) ([]byte, error) {
	switch encoding {
	case "cp874": // 泰语
		enc := charmap.Windows874.NewEncoder()
		return enc.Bytes([]byte(s))
	case "cp949": // 韩语
		enc := korean.EUCKR.NewEncoder()
		return enc.Bytes([]byte(s))
	case "gbk": // 中文
		enc := simplifiedchinese.GBK.NewEncoder()
		return enc.Bytes([]byte(s))
	case "iso8859-9", "latin5": // 土耳其语
		enc := charmap.ISO8859_9.NewEncoder()
		return enc.Bytes([]byte(s))
	default:
		return []byte(s), nil
	}
}

// hex2bin 将十六进制字符串转换为二进制数据
// 类似于PHP中的同名函数
func hex2bin(hexStr string) string {
	// 移除可能存在的空格
	hexStr = strings.ReplaceAll(hexStr, " ", "")

	// 解码十六进制字符串
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		fmt.Printf("解析十六进制字符串出错: %v\n", err)
		return ""
	}

	// 返回解码后的字符串
	return string(decoded)
}
