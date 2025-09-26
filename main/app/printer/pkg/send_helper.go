package pkg

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// PrintSunmiTicket 商米打印
func PrintSunmiTicket(config model.PrinterConfigJson, content string) error {
	printer := NewPrinter(567, config.APP_ID, config.APP_KEY)
	// 云打印
	if config.SN != "" && config.APP_ID != "" && config.APP_KEY != "" {
		printer.orderData = content
		res, err := printer.PushContent(config.SN, fmt.Sprintf("%d%s", time.Now().Unix(), hex.EncodeToString([]byte(uniqid()))[:8]), 1, 1, "", 1)
		if err != nil {
			fmt.Println("请求错误", err)
			return fmt.Errorf("请求错误")
		}

		// 获取code
		var code int
		switch v := res["code"].(type) {
		case int:
			code = v
		case float64:
			code = int(v)
		default:
			if v != nil {
				fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &code)
			}
		}

		resJSON, _ := json.Marshal(res)

		// 判断code
		switch code {
		case 1, 100001:
			return nil
		case 20000:
			return fmt.Errorf("网关校验缺少必要参数")
		case 20001:
			return fmt.Errorf("请求超过有效期")
		case 30000:
			return fmt.Errorf("开发者身份验证失败（APPID无效；访问IP不在配置的IP白名单列表）")
		case 30001:
			return fmt.Errorf("用户缺少相关权限（能力未关联）")
		case 40000:
			return fmt.Errorf("签名验证失败")
		case 50000:
			return fmt.Errorf("服务器异常")
		case 50001:
			return fmt.Errorf("网关异常")
		case 10071400:
			return fmt.Errorf("参数错误，请注意参数类型是否错误")
		case 10071500:
			return fmt.Errorf("服务器错误")
		case 10071701:
			return fmt.Errorf("未知设备，需要检查SN号是否错误")
		case 10071702:
			return fmt.Errorf("设备已绑定")
		case 10071703:
			return fmt.Errorf("设备绑定异常")
		case 10071704:
			return fmt.Errorf("该SN号设备不属于本渠道，需要联系销售绑定到贵司渠道下")
		case 10071705:
			return fmt.Errorf("订单已推送，需要确认订单号是否重复")
		case 10071706:
			return fmt.Errorf("订单未知")
		case 10071707:
			return fmt.Errorf("没有设备或不是此通道")
		default:
			return fmt.Errorf("异常错误: %s", string(resJSON))
		}
	} else if config.IP != "" && config.SN != "" {
		// 局域网打印
		url := "http://" + config.IP + "/cgi-bin/print.cgi?sn=" + config.SN + "&copies=1"
		_, err := printer.httpPost(url, map[string]interface{}{
			"content": content,
		})
		if err != nil {
			return fmt.Errorf("打印失败")
		}
		return nil
	}
	//
	return fmt.Errorf("打印配置不全")
}

// PrintTicket 打印
func PrintTicket(printerIP string, port string, content string, printMethod int) error {
	if printMethod == constant.Yes {
		return ExecutePrinting(printerIP, port, content)
	}
	return ExecuteImgPrinting(printerIP, port, content)
}

// 图片打印
func ExecuteImgPrinting(printerIP string, port string, content string) error {
	// 连接打印机（TCP连接，端口9100是标准打印机端口）
	conn, err := net.DialTimeout("tcp", printerIP+":"+port, 3*time.Second)
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
func ExecutePrinting(printerIP string, port string, content string) error {
	// 连接打印机（TCP连接，端口9100是标准打印机端口）
	conn, err := net.DialTimeout("tcp", printerIP+":"+port, 3*time.Second)
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
				fmt.Printf("转换为CP874编码出错: %v %v\n", segment, err)
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
				fmt.Printf("转换为CP949编码出错: %v %v\n", segment, err)
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
				fmt.Printf("转换为ISO-8859-9编码出错: %v %v\n", segment, err)
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
			encoded, err := BetterEncodeTo(segment, "gbk")
			// 转换为GBK编码（中文字符集）
			if err != nil {
				fmt.Printf("转换为GBK编码出错: %v %v\n", segment, err)
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

// Hex2bin 公开的hex2bin函数，用于外部调用
func Hex2bin(hexStr string) string {
	return hex2bin(hexStr)
}

// 生成唯一ID，模拟PHP的uniqid函数
func uniqid() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}
