// Package pkg 提供打印机相关功能
package pkg

import (
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// PrintTextHelper 提供文本处理和格式化功能用于打印
type PrintTextHelper struct{}

// countThaiGraphemeClusters 计算泰语字符串中的字素簇数量
// 这个函数模拟PHP的grapheme_strlen函数的行为
func (h *PrintTextHelper) countThaiGraphemeClusters(text string) int {
	// 泰语元音和声调标记，这些通常与前面的辅音组合成一个字素簇
	combiningMarks := map[rune]bool{
		'่': true, // 声调符号
		'้': true, // 声调符号
		'๊': true, // 声调符号
		'๋': true, // 声调符号
		'ั': true, // 元音符号
		'ิ': true, // 元音符号
		'ี': true, // 元音符号
		'ึ': true, // 元音符号
		'ื': true, // 元音符号
		'ุ': true, // 元音符号
		'ู': true, // 元音符号
		'ำ': true, // 元音符号组合
		'็': true, // 其他符号
	}

	count := 0
	skipNext := false
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		count++
		// 检查下一个字符是否是组合标记
		if i+1 < len(runes) && combiningMarks[runes[i+1]] {
			skipNext = true
		}
	}

	return count
}

// GetTextWidth 获取文本宽度
func (h *PrintTextHelper) GetTextWidth(text string) int {
	// 泰语检测 - 对应PHP: preg_match_all('/[\p{Thai}฿]+/u', $text, $matches);
	thaiRegex := regexp.MustCompile(`[\p{Thai}฿]+`)
	if thaiRegex.MatchString(text) {
		// 1. 提取所有泰语字符序列 - 对应PHP: $ttf = implode('', $matches[0]);
		thaiMatches := thaiRegex.FindAllString(text, -1)
		ttf := strings.Join(thaiMatches, "")
		// 2. 计算泰语字符的基本宽度 - 对应PHP: $w = grapheme_strlen($ttf);
		// 使用自定义函数计算泰语字素簇数量，而不是简单的Unicode字符数
		w := h.countThaiGraphemeClusters(ttf)
		// 检查特殊字符
		for _, r := range ttf {
			if string(r) == "ำ" {
				w++
			}
		}
		// 3. 从原文本中移除泰语字符
		textWithoutThai := thaiRegex.ReplaceAllString(text, "")
		// 4. 转换为GBK并添加长度 - 对应PHP: $w += strlen(iconv("UTF-8", "GBK//IGNORE", $text));
		if textWithoutThai != "" {
			encoder := simplifiedchinese.GBK.NewEncoder()
			gbkStr, _, err := transform.String(encoder, textWithoutThai)
			if err != nil {
				w += len(textWithoutThai)
			} else {
				w += len(gbkStr)
			}
		}
		//
		return w
	}

	// 缅甸语检测
	myanmarRegex := regexp.MustCompile(`[\p{Myanmar}]`)
	if myanmarRegex.MatchString(text) {
		// 1. 提取所有缅甸语字符
		myanmarMatches := myanmarRegex.FindAllString(text, -1)
		myanmar := strings.Join(myanmarMatches, "")

		// 2. 计算缅甸语字符的基本宽度
		w := utf8.RuneCountInString(myanmar)

		// 3. 从原文本中移除缅甸语字符
		textWithoutMyanmar := myanmarRegex.ReplaceAllString(text, "")

		// 4. 转换为GBK并添加长度
		encoder := simplifiedchinese.GBK.NewEncoder()
		gbkStr, _, err := transform.String(encoder, textWithoutMyanmar)
		if err == nil {
			w += len(gbkStr)
		} else {
			w += len(textWithoutMyanmar)
		}

		return w
	}

	// 韩语检测
	koreanRegex := regexp.MustCompile(`[\p{Hangul}]`)
	if koreanRegex.MatchString(text) {
		encoder := korean.EUCKR.NewEncoder()
		eucKrStr, _, _ := transform.String(encoder, text)
		return len(eucKrStr)
	}

	// 默认情况，转换为GBK计算长度
	encoder := simplifiedchinese.GBK.NewEncoder()
	gbkStr, _, _ := transform.String(encoder, text)
	return len(gbkStr)
}

// InterceptText 截取文本
func (h *PrintTextHelper) InterceptText(text string, num int, intervalNum int) (string, string) {
	if num <= 0 || text == "" {
		return text, ""
	}

	afterText := ""
	nums := num - intervalNum

	// 检测中文/日文
	cjkRegex := regexp.MustCompile(`[\p{Han}\p{Hiragana}\p{Katakana}]`)
	if cjkRegex.MatchString(text) {
		tmpText := SubString(text, 0, int(math.Ceil(float64(nums)/2)))
		nonCJK := regexp.MustCompile(`[^\p{Han}\p{Hiragana}\p{Katakana}]`)
		matches := nonCJK.FindAllString(tmpText, -1)
		tmpTextCount := len(matches)

		if tmpTextCount > 1 {
			pos := int(math.Ceil(float64(nums)/2 + float64(tmpTextCount)/2))
			afterText = SubString(text, pos, 1000)
			text = SubString(text, 0, pos)
		} else {
			pos := int(math.Ceil(float64(nums) / 2))
			afterText = SubString(text, pos, 1000)
			text = tmpText
		}
	} else if thaiRegex := regexp.MustCompile(`[\p{Thai}]`); thaiRegex.MatchString(text) {
		// 泰语
		pos := int(math.Ceil(float64(nums) * 1.2))
		afterText = SubString(text, pos, 1000)
		text = SubString(text, 0, pos)
	} else if koreanRegex := regexp.MustCompile(`[\p{Hangul}]`); koreanRegex.MatchString(text) {
		// 韩语
		pos := int(math.Ceil(float64(nums) / 1.7))
		afterText = SubString(text, pos, 1000)
		text = SubString(text, 0, pos)
	} else {
		// 其他
		afterText = SubString(text, nums, 1000)
		text = SubString(text, 0, nums)
	}

	// 处理换行符
	if strings.Contains(text, "\n") {
		texts := strings.SplitN(text, "\n", 2)
		text = texts[0]
		if len(texts) > 1 {
			afterText = texts[1] + afterText
		}
	}

	return text, afterText
}

// FilterCharacter 过滤特殊字符
func (h *PrintTextHelper) FilterCharacter(text string) string {
	text = strings.ReplaceAll(text, "​​", "")
	text = strings.ReplaceAll(text, "　", " ")
	text = strings.ReplaceAll(text, "ー", "-")
	text = strings.ReplaceAll(text, "グ", "ク")
	text = strings.ReplaceAll(text, "・", "·")
	return text
}

// PrintText 获取格式化的打印文本
func (h *PrintTextHelper) PrintText(
	leftText, centerText, rightText string,
	total, leftNum, centerNum, rightNum, intervalNum int,
) string {
	if leftText == "" {
		leftText = ""
	}
	if centerText == "" {
		centerText = ""
	}
	if rightText == "" {
		rightText = ""
	}

	// 过滤和修剪文本
	leftText = h.FilterCharacter(strings.TrimSpace(leftText))
	centerText = h.FilterCharacter(strings.TrimSpace(centerText))
	rightText = h.FilterCharacter(strings.TrimSpace(rightText))

	// 截取文本
	var afterLeftText, afterCenterText, afterRightText string
	leftText, afterLeftText = h.InterceptText(leftText, leftNum, intervalNum)
	centerText, afterCenterText = h.InterceptText(centerText, centerNum, intervalNum)
	rightText, afterRightText = h.InterceptText(rightText, rightNum, intervalNum)

	// 计算宽度和填充
	leftWidth := h.GetTextWidth(leftText)

	leftPadding := ""
	if leftNum-leftWidth > 0 {
		leftPadding = strings.Repeat(" ", leftNum-leftWidth)
	}

	leftPaddingWidth := h.GetTextWidth(leftPadding)
	centerWidth := 0
	if centerText != "!" {
		centerWidth = h.GetTextWidth(centerText)
	}

	rightWidth := h.GetTextWidth(rightText)
	centerPaddingWidth := total - leftWidth - leftPaddingWidth - centerWidth - rightWidth

	centerPadding := ""
	if centerPaddingWidth > 0 {
		centerPadding = strings.Repeat(" ", centerPaddingWidth)
	}

	// 生成内容
	content := leftText + leftPadding + centerText + centerPadding + rightText

	// 处理多行内容
	if afterLeftText != "" || afterCenterText != "" || afterRightText != "" {
		content += "\n" + h.PrintText(afterLeftText, afterCenterText, afterRightText, total, leftNum, centerNum, rightNum, intervalNum)
	}

	return content
}

// SubString 安全地截取字符串的指定范围（按Unicode字符计算）
func SubString(str string, start, length int) string {
	if length <= 0 {
		return ""
	}

	runes := []rune(str)
	strLen := len(runes)

	if start >= strLen {
		return ""
	}

	if start+length > strLen {
		length = strLen - start
	}

	return string(runes[start : start+length])
}

// NewPrintTextHelper 创建PrintTextHelper的新实例
func NewPrintTextHelper() *PrintTextHelper {
	return &PrintTextHelper{}
}
