// TextTemplateParser 模板解析器
// @author: wfs
// @date: 2024/12/20
package pkg

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"ttpos-server-go/pkg/utils"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

type TextBaseData struct {
	Language             string `json:"language"`               // 语言
	CurrencyUnit         string `json:"currency_unit"`          // 货币单位
	CurrencyUnitPosition int    `json:"currency_unit_position"` // 货币单位位置
	Is58mmPrinter        bool   `json:"is_58mm_printer"`        // 是否58mm打印机
}

// TextTemplateMetadata 模板元数据
type TextTemplateMetadata struct {
	Name        string `json:"name"`        // 模板名称
	Description string `json:"description"` // 模板描述
	PaperWidth  int    `json:"paper_width"` // 纸张宽度（毫米）
	Version     string `json:"version"`     // 模板版本
	Thousandth  bool   `json:"thousandth"`  // 是否千分位
}

// TextTemplateBlock 模板块定义
type TextTemplateBlock struct {
	BlockID           string                    `json:"block_id"`            // 块唯一标识
	BlockType         string                    `json:"block_type"`          // 块类型 label, value, label:auto:value, array, Text, qrcode, barcode, blank_line
	BlockLabel        interface{}               `json:"block_label"`         // 标签（支持字符串或多语言）
	BlockBeforeLabel  interface{}               `json:"block_before_label"`  // 块前标签（支持字符串或多语言）
	BlockAfterLabel   interface{}               `json:"block_after_label"`   // 块后标签（支持字符串或多语言）
	BlockExpandLabels []TextTemplateExpandLabel `json:"block_expand_labels"` // 块扩展标签（支持条件显示）
	BlockBrotherID    string                    `json:"block_brother_id"`    // 兄弟块ID
	BlockAttr         TextTemplateBlockAttr     `json:"block_attr"`          // 块属性
	Rows              [][]TextTemplateBlock     `json:"rows,omitempty"`      // 嵌套行
	Conditions        []TextTemplateCondition   `json:"conditions"`          // 条件显示规则
}

// TextTemplateBlockAttr 块属性定义
type TextTemplateBlockAttr struct {
	// 样式配置
	FontSize   int    `json:"font_size"`   // 字体大小
	FontBold   bool   `json:"font_bold"`   // 是否粗体
	FontWeight int    `json:"font_weight"` // 字体粗细 (1-3)
	Align      string `json:"align"`       // 对齐方式

	// 布局配置
	Width              interface{} `json:"width"`                // 宽度百分比（支持固定值或按语言动态配置）
	LabelWidth         float64     `json:"label_width"`          // 标签宽度百分比
	Height             int         `json:"height"`               // 固定高度
	DividingLine       bool        `json:"dividing_line"`        // 是否显示分割线
	LeadingBlankLines  int         `json:"leading_blank_lines"`  // 前置空行数
	TrailingBlankLines int         `json:"trailing_blank_lines"` // 后置空行数
	LineHeight         int         `json:"line_height"`          // 行高倍数
	WordWrap           bool        `json:"word_wrap"`            // 是否自动换行
	PaddingRight       int         `json:"padding_right"`        // 右边距
	ShowCurrencyUnit   bool        `json:"show_currency_unit"`   // 是否显示货币单位
	NotShowEmpty       bool        `json:"not_show_empty"`       // 是否不显示空值
	ShowColumnTitle    bool        `json:"show_column_title"`    // 是否显示列标题
	Disabled           bool        `json:"disabled"`             // 是否禁用

	// 数据绑定
	DefaultValue string `json:"default_value"` // 默认值
}

// TextTemplateCondition 条件显示规则
type TextTemplateCondition struct {
	Field    string      `json:"field"`    // 字段路径
	Operator string      `json:"operator"` // 操作符
	Value    interface{} `json:"value"`    // 比较值
}

// TextTemplateExpandLabel 扩展标签定义
type TextTemplateExpandLabel struct {
	Text       interface{}             `json:"text"`       // 标签文本（支持字符串或多语言）
	Conditions []TextTemplateCondition `json:"conditions"` // 显示条件
}

// TextTemplate JSON模板定义
type TextTemplate struct {
	Metadata TextTemplateMetadata  `json:"metadata"` // 模板元数据
	Rows     [][]TextTemplateBlock `json:"rows"`     // 模板行
}

// TextTemplateParser JSON模板解析器
type TextTemplateParser struct {
	template *TextTemplate          // 模板配置
	data     map[string]interface{} // 数据源
	baseData TextBaseData           // 基础数据
}

// NewTextTemplateParser 创建新的JSON模板解析器
func NewTextTemplateParser(baseData TextBaseData, templateJSON string, data map[string]interface{}) (*TextTemplateParser, error) {
	var template TextTemplate
	err := json.Unmarshal([]byte(templateJSON), &template)
	if err != nil {
		return nil, fmt.Errorf("解析模板JSON失败: %v", err)
	}

	// 设置默认语言
	if baseData.Language == "" {
		baseData.Language = "zh"
	}

	return &TextTemplateParser{
		template: &template,
		data:     data,
		baseData: baseData,
	}, nil
}

// Parse 解析模板并生成打印内容
func (p *TextTemplateParser) Parse(IsEnableSound bool) (string, error) {
	// 根据纸张宽度创建 ESC 打印机对象
	var printer *Printers
	paperWidth := p.template.Metadata.PaperWidth

	switch paperWidth {
	case 80:
		printer = NewPrinter(568) // 80mm纸张
	case 58:
		printer = NewPrinter(384) // 58mm纸张
	default:
		if p.baseData.Is58mmPrinter {
			printer = NewPrinter(384) // 58mm纸张
		} else {
			printer = NewPrinter(568) // 80mm纸张
		}
	}

	// 初始化打印机设置
	printer.RestoreDefaultSettings()
	printer.SetUtf8Mode(1) // 启用UTF-8模式

	// 解析模板行
	err := p.parseRows(printer, p.template.Rows, 0, false)
	if err != nil {
		return "", fmt.Errorf("解析模板行失败: %v", err)
	}

	// 打印并退出页面模式
	printer.PrintAndExitPageMode()
	printer.LineFeed(2)
	printer.CutPaper(IsEnableSound)

	//
	return printer.GetOrderData(), nil
}

// fixEncodingIssues 修复编码问题
func (p *TextTemplateParser) fixEncodingIssues(hexData string) string {
	// 转换为二进制数据
	binaryData := hex2bin(hexData)

	// 查找和替换UTF-8编码的中文字符
	result := []byte{}
	i := 0

	for i < len(binaryData) {
		// 检查是否是ESC指令
		if i < len(binaryData)-1 && binaryData[i] == 0x1B {
			// 这是ESC指令，直接复制
			result = append(result, binaryData[i])
			i++
			continue
		}

		// 检查是否是中文字符（UTF-8 3字节）
		if i < len(binaryData)-2 {
			b1, b2, b3 := binaryData[i], binaryData[i+1], binaryData[i+2]
			if (b1 >= 0xE4 && b1 <= 0xE9) && (b2 >= 0x80 && b2 <= 0xBF) && (b3 >= 0x80 && b3 <= 0xBF) {
				// 这是UTF-8编码的中文字符，转换为GBK
				utf8Bytes := []byte{b1, b2, b3}
				if utf8.Valid(utf8Bytes) {
					char := string(utf8Bytes)
					// 转换为GBK编码
					gbkBytes, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(char))
					if err == nil {
						result = append(result, gbkBytes...)
						i += 3
						continue
					}
				}
			}
		}

		// 其他字节直接复制
		result = append(result, binaryData[i])
		i++
	}

	// 转换回16进制字符串
	fixedHexData := hex.EncodeToString(result)

	return fixedHexData
}

// Thousandth 计算金额的千分位
func (p *TextTemplateParser) Thousandth(amount float64) string {
	if amount == 0 {
		return "0"
	}
	// 先格式化为2位小数
	formattedAmount := strconv.FormatFloat(amount, 'f', 2, 64)

	// 分割整数和小数部分
	parts := strings.Split(formattedAmount, ".")
	integerPart := parts[0]
	decimalPart := ""
	if len(parts) > 1 {
		decimalPart = parts[1]
	}

	// 处理整数部分，添加千分位
	var result strings.Builder
	length := len(integerPart)
	for i := length - 1; i >= 0; i-- {
		result.WriteByte(integerPart[i])
		if i > 0 && (length-i)%3 == 0 {
			result.WriteByte(',')
		}
	}

	// 反转字符串
	reversed := result.String()
	runes := []rune(reversed)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	integerPart = string(runes)

	// 处理小数部分，去除尾部的0
	decimalPart = strings.TrimRight(decimalPart, "0")

	// 组合结果
	if decimalPart != "" {
		return integerPart + "." + decimalPart
	}
	return integerPart
}

// parseRows 解析模板行
func (p *TextTemplateParser) parseRows(printer *Printers, rows [][]TextTemplateBlock, level int, isLast bool) error {
	for _, row := range rows {
		err := p.parseRow(printer, row, level, isLast)
		if err != nil {
			return fmt.Errorf("解析行失败: %v", err)
		}
	}
	return nil
}

// appendSplitLine 添加分割线（适配ESC打印机）
func (p *TextTemplateParser) appendSplitLine(printer *Printers) {
	// 根据打印机点数生成分割线
	if printer.dotsPerLine >= 568 {
		printer.AppendText("------------------------------------------------")
	} else if printer.dotsPerLine >= 384 {
		printer.AppendText("--------------------------------")
	}
	printer.LineFeed(1)
}

// restoreDefault 恢复默认设置（适配ESC打印机）
func (p *TextTemplateParser) restoreDefault(printer *Printers) {
	printer.SetPrintModes(false, false, false) // 恢复正常字体
	printer.SetAlignment(AlignLeft)            // 恢复左对齐
}

// getQRCodeSize 根据宽度获取二维码模块大小
func (p *TextTemplateParser) getQRCodeSize(width interface{}) int {
	w := p.getWidthValue(width)
	if w > 50 {
		return 8 // 大尺寸
	} else if w > 30 {
		return 6 // 中等尺寸
	}
	return 4 // 小尺寸
}

// getNextValidBlock 获取下一行有效的block
func (p *TextTemplateParser) validBlock(block TextTemplateBlock) bool {
	// 如果设置了不显示空值，且值为空或零，则返回空字符串
	if block.BlockAttr.NotShowEmpty && p.isEmptyOrZero(p.getDataValue(block.BlockID)) {
		return false
	}
	// 检查条件显示
	if !p.checkConditions(block.Conditions) {
		return false
	}
	return true
}

// getNextValidRow 获取下一行有效的blocks
func (p *TextTemplateParser) getNextValidRow(rows [][]TextTemplateBlock, startIndex int) []TextTemplateBlock {
	// 从指定索引开始查找下一行有效的blocks
	for i := startIndex; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue // 跳过空行
		}
		// 检查这一行是否有效（至少有一个有效的block）
		for _, block := range row {
			// 检查条件显示
			if !p.validBlock(block) {
				continue
			}
			// 找到有效的block，返回这一行
			return row
		}
	}
	// 没有找到有效的行，返回空切片
	return []TextTemplateBlock{}
}

// parseRow 解析单行
func (p *TextTemplateParser) parseRow(
	printer *Printers,
	blocks []TextTemplateBlock,
	level int,
	isLast bool,
) error {
	if len(blocks) == 0 {
		return nil
	}

	block := blocks[0]

	if block.BlockAttr.Disabled || !p.validBlock(block) {
		return nil
	}

	// 设置前置空行
	if block.BlockAttr.LeadingBlankLines > 0 {
		printer.SetLineSpacing(24)
		printer.LineFeed(block.BlockAttr.LeadingBlankLines)
	}

	// 设置字体样式
	p.applyFontStyle(printer, block.BlockAttr)

	// 设置对齐方式
	printer.SetAlignment(p.convertAlign(block.BlockAttr.Align))

	// 如果块ID不是空白行，则处理
	if block.BlockID != "blank_line" {

		// 如果只有一个块，直接处理
		if len(blocks) == 1 && block.BlockType != "column" {
			text := p.getBlockText(block)

			// 添加文本
			if block.BlockType == "Text" {
				printer.AppendText(text)
				printer.LineFeed(1)
			} else if block.BlockType == "qrcode" {
				// 设置二维码大小和错误纠正级别
				moduleSize := p.getQRCodeSize(block.BlockAttr.Width)
				printer.AppendQRcode(moduleSize, 1, text) // 默认错误纠正级别为1
				printer.LineFeed(1)
			} else if block.BlockType == "barcode" {
				// 添加条形码，使用默认参数
				printer.AppendBarcode(HriPosBelow, 120, 3, 73, text) // CODE128条形码
				printer.LineFeed(1)
			} else if strings.Contains(block.BlockType, "label:auto:value") {
				// 分列打印：标签和值
				label := p.getLabel(block.BlockLabel)
				value := p.getBlockText(block)

				// 设置列宽度，标签左对齐，值右对齐
				labelWidth := p.calculateColumnWidth(printer, 60) // 60%给标签
				valueWidth := p.calculateColumnWidth(printer, 40) // 40%给值

				printer.SetupColumns([]int{labelWidth, AlignLeft, 0}, []int{valueWidth, AlignRight, 0})
				printer.PrintInColumns(label, value)
			} else if !strings.Contains(block.BlockType, "array") {
				// 获取显示文本
				printer.AppendText(text)
				// 添加换行
				if len(text) > 0 {
					printer.LineFeed(1)
				}
			}

		} else if (len(block.Rows) == 0 || level > 0) && block.BlockType != "column" {
			// 创建列配置
			columnWidths := make([][]int, 0, len(blocks))
			columnTexts := make([]string, 0, len(blocks))

			for _, block := range blocks {
				// 检查条件显示
				if !p.checkConditions(block.Conditions) {
					continue
				}

				// 计算列宽度
				width := p.calculateColumnWidth(printer, p.getWidthValue(block.BlockAttr.Width))
				align := p.convertAlign(block.BlockAttr.Align)
				flag := 0
				if p.getFontWeight(block.BlockAttr) > 1 {
					flag |= ColumnFlagBold
				}

				columnWidths = append(columnWidths, []int{width, align, flag})
				columnTexts = append(columnTexts, p.getBlockText(block))
			}

			// 如果有有效的列，进行分列打印
			if len(columnWidths) > 0 {
				printer.SetupColumns(columnWidths...)
				printer.PrintInColumns(columnTexts...)
			}

			// 处理嵌套行 - 规格，备注等
			if level > 0 {
				for _, block := range blocks {
					if len(block.Rows) > 0 {
						err := p.parseRows(printer, block.Rows, level+1, isLast)
						if err != nil {
							return err
						}
					}
				}
			}
		} else {
			// 处理嵌套行
			for _, block := range blocks {
				if len(block.Rows) > 0 {
					originalData := p.data
					blockDatas := p.getData(block.BlockID)
					for index, blockData := range blockDatas {
						p.data = blockData
						isLast := index == len(blockDatas)-1 && block.BlockAttr.DividingLine
						err := p.parseRows(printer, block.Rows, level+1, isLast)
						p.data = originalData // 恢复原始数据
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	// 添加分割线
	if block.BlockAttr.DividingLine || block.BlockType == "blank_line" {
		p.appendSplitLine(printer)
	}

	// 设置后置空行
	if block.BlockAttr.TrailingBlankLines > 0 {
		printer.SetLineSpacing(24)
		printer.LineFeed(block.BlockAttr.TrailingBlankLines)
	}

	// 恢复默认值
	p.restoreDefault(printer)

	return nil
}

// calculateColumnWidth 计算列宽度（适配ESC打印机）
func (p *TextTemplateParser) calculateColumnWidth(printer *Printers, widthPercent float64) int {
	// 根据打印机点数计算列宽度
	if widthPercent > 0 {
		return int(float64(printer.dotsPerLine) * (widthPercent / 100))
	}
	return printer.dotsPerLine / 4 // 默认宽度
}

// getWidthValue 获取宽度值（支持固定值或按语言动态配置）
func (p *TextTemplateParser) getWidthValue(width interface{}) float64 {
	if width == nil {
		return 0
	}

	// 如果是数字类型，直接返回
	switch v := width.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	}

	// 如果是map类型，按语言获取宽度
	if widthMap, ok := width.(map[string]interface{}); ok {
		// 获取当前语言的宽度
		if langWidth, exists := widthMap[p.baseData.Language]; exists {
			if w := p.convertToFloat64(langWidth); w > 0 {
				return w
			}
		}

		// 回退到中文
		if langWidth, exists := widthMap["zh"]; exists {
			if w := p.convertToFloat64(langWidth); w > 0 {
				return w
			}
		}

		// 回退到英文
		if langWidth, exists := widthMap["en"]; exists {
			if w := p.convertToFloat64(langWidth); w > 0 {
				return w
			}
		}

		// 返回第一个可用的宽度
		for _, langWidth := range widthMap {
			if w := p.convertToFloat64(langWidth); w > 0 {
				return w
			}
		}
	}

	return 0
}

// convertToFloat64 将interface{}转换为float64
func (p *TextTemplateParser) convertToFloat64(value interface{}) float64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		// 尝试解析字符串为数字
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return 0
	default:
		return 0
	}
}

// getBlockText 获取块的显示文本
func (p *TextTemplateParser) getBlockText(block TextTemplateBlock) string {
	// 如果有数据源，检查是否需要隐藏空值
	if block.BlockID != "" && (strings.Contains(block.BlockType, "value") || block.BlockType == "Text" || block.BlockType == "qrcode" || block.BlockType == "barcode") {
		value := p.getDataValue(block.BlockID)
		formatted := p.formatValue(value, block.BlockAttr)
		// 获取前标签文本
		beforeLabel := p.getLabelText(block.BlockBeforeLabel)
		// 获取主标签文本
		label := utils.IfString(strings.Contains(block.BlockType, "label"), p.getLabel(block.BlockLabel), "")
		// 获取后标签文本
		afterLabel := p.getLabelText(block.BlockAfterLabel)
		// 获取扩展标签文本
		expandLabel := p.getExpandLabelsText(block.BlockExpandLabels)
		if strings.Contains(block.BlockType, "label:auto:value") {
			// 组合前标签、值、后标签和扩展标签
			return p.combineLabelsWithExpand(beforeLabel, formatted, afterLabel, expandLabel)
		}
		if label != "" {
			// 组合前标签、主标签、值、后标签和扩展标签
			return p.combineLabelsWithExpand(beforeLabel, fmt.Sprintf("%s%s", label, formatted), afterLabel, expandLabel)
		}
		// 组合前标签、值、后标签和扩展标签
		return p.combineLabelsWithExpand(beforeLabel, formatted, afterLabel, expandLabel)
	}

	// 获取前标签文本
	beforeLabel := p.getLabelText(block.BlockBeforeLabel)
	// 获取主标签文本
	label := utils.IfString(strings.Contains(block.BlockType, "label"), p.getLabel(block.BlockLabel), "")
	// 获取后标签文本
	afterLabel := p.getLabelText(block.BlockAfterLabel)
	// 获取扩展标签文本
	expandLabel := p.getExpandLabelsText(block.BlockExpandLabels)
	// 组合前标签、主标签、后标签和扩展标签
	return p.combineLabelsWithExpand(beforeLabel, label, afterLabel, expandLabel)
}

// GetBlockText 公开方法，获取块的显示文本
func (p *TextTemplateParser) GetBlockText(block TextTemplateBlock) string {
	return p.getBlockText(block)
}

// SetLanguage 设置语言
func (p *TextTemplateParser) SetLanguage(language string) {
	p.baseData.Language = language
}

// GetBlockWidth 公开方法，获取块的宽度（已适配ESC打印机）
func (p *TextTemplateParser) GetBlockWidth(printer *Printers, block TextTemplateBlock) int {
	return p.calculateColumnWidth(printer, p.getWidthValue(block.BlockAttr.Width))
}

// getLabel 获取当前语言的标签
func (p *TextTemplateParser) getLabel(labels interface{}) string {
	if labels == nil {
		return ""
	}

	// 如果是字符串类型，直接返回
	if labelStr, ok := labels.(string); ok {
		return p.replaceVariables(labelStr)
	}

	// 如果是map类型，按语言获取标签
	if labelsMap, ok := labels.(map[string]interface{}); ok {
		var label string

		// 获取当前语言的标签
		if l, exists := labelsMap[p.baseData.Language]; exists {
			if labelStr, ok := l.(string); ok && labelStr != "" {
				label = labelStr
			}
		}

		// 回退到中文
		if label == "" {
			if l, exists := labelsMap["zh"]; exists {
				if labelStr, ok := l.(string); ok && labelStr != "" {
					label = labelStr
				}
			}
		}

		// 回退到英文
		if label == "" {
			if l, exists := labelsMap["en"]; exists {
				if labelStr, ok := l.(string); ok && labelStr != "" {
					label = labelStr
				}
			}
		}

		// 返回第一个可用的标签
		if label == "" {
			for _, l := range labelsMap {
				if labelStr, ok := l.(string); ok && labelStr != "" {
					label = labelStr
					break
				}
			}
		}

		// 处理标签中的变量替换
		return p.replaceVariables(label)
	}

	// 如果是map[string]string类型（向后兼容）
	if labelsMap, ok := labels.(map[string]string); ok {
		var label string

		// 获取当前语言的标签
		if l, exists := labelsMap[p.baseData.Language]; exists && l != "" {
			label = l
		} else if l, exists := labelsMap["zh"]; exists && l != "" {
			// 回退到中文
			label = l
		} else if l, exists := labelsMap["en"]; exists && l != "" {
			// 回退到英文
			label = l
		} else {
			// 返回第一个可用的标签
			for _, l := range labelsMap {
				if l != "" {
					label = l
					break
				}
			}
		}

		// 处理标签中的变量替换
		return p.replaceVariables(label)
	}

	return ""
}

// getLabelText 获取标签文本（支持变量替换）
func (p *TextTemplateParser) getLabelText(labels interface{}) string {
	if labels == nil {
		return ""
	}

	// 使用 getLabel 方法获取当前语言的标签
	return p.getLabel(labels)
}

// getExpandLabelsText 获取扩展标签文本（支持条件显示）
func (p *TextTemplateParser) getExpandLabelsText(expandLabels []TextTemplateExpandLabel) string {
	if len(expandLabels) == 0 {
		return ""
	}

	var result []string
	for _, expandLabel := range expandLabels {
		// 检查所有条件是否都满足
		allConditionsMet := true
		for _, condition := range expandLabel.Conditions {
			if !p.checkCondition(condition) {
				allConditionsMet = false
				break
			}
		}

		// 如果所有条件都满足，获取标签文本
		if allConditionsMet {
			labelText := p.getLabel(expandLabel.Text)
			if labelText != "" {
				result = append(result, labelText)
			}
		}
	}

	return strings.Join(result, " ")
}

// combineLabels 组合标签文本
func (p *TextTemplateParser) combineLabels(before, main, after string) string {
	var parts []string
	// 添加前标签
	if before != "" {
		parts = append(parts, before)
	}
	// 添加主内容
	if main != "" {
		parts = append(parts, main)
	}
	// 添加后标签
	if after != "" {
		parts = append(parts, after)
	}
	// 用空格连接所有部分
	return strings.Join(parts, "")
}

// combineLabelsWithExpand 组合标签文本（包含扩展标签）
func (p *TextTemplateParser) combineLabelsWithExpand(before, main, after, expand string) string {
	var parts []string
	// 添加扩展标签
	if expand != "" {
		parts = append(parts, expand)
	}
	// 添加前标签
	if before != "" {
		parts = append(parts, before)
	}
	// 添加主内容
	if main != "" {
		parts = append(parts, main)
	}
	// 添加后标签
	if after != "" {
		parts = append(parts, after)
	}
	// 用空格连接所有部分
	return strings.Join(parts, "")
}

// checkCondition 检查条件是否满足
func (p *TextTemplateParser) checkCondition(condition TextTemplateCondition) bool {
	fieldValue := p.getDataValue(condition.Field)

	switch condition.Operator {
	case "eq": // 等于
		return p.convertToFloat64(fieldValue) == p.convertToFloat64(condition.Value)
	case "ne": // 不等于
		return p.convertToFloat64(fieldValue) != p.convertToFloat64(condition.Value)
	case "gt": // 大于
		return p.compareValues(fieldValue, condition.Value) > 0
	case "lt": // 小于
		return p.compareValues(fieldValue, condition.Value) < 0
	case "gte": // 大于等于
		return p.compareValues(fieldValue, condition.Value) >= 0
	case "lte": // 小于等于
		return p.compareValues(fieldValue, condition.Value) <= 0
	case "empty": // 空
		return p.isEmptyOrZero(fieldValue)
	case "not_empty": // 非空
		return !p.isEmptyOrZero(fieldValue)
	case "contains": // 包含
		return p.containsValue(fieldValue, condition.Value)
	case "not_contains": // 不包含
		return !p.containsValue(fieldValue, condition.Value)
	case "in": // 在
		return p.isInValue(fieldValue, condition.Value)
	case "not_in": // 不在
		return !p.isInValue(fieldValue, condition.Value)
	default:
		return false
	}
}

// compareValues 比较两个值的大小
func (p *TextTemplateParser) compareValues(a, b interface{}) int {
	// 转换为float64进行比较
	valA := p.convertToFloat64(a)
	valB := p.convertToFloat64(b)

	if valA > valB {
		return 1
	} else if valA < valB {
		return -1
	}
	return 0
}

// containsValue 检查值是否包含指定内容
func (p *TextTemplateParser) containsValue(fieldValue, targetValue interface{}) bool {
	if fieldValue == nil || targetValue == nil {
		return false
	}

	fieldStr := fmt.Sprintf("%v", fieldValue)
	targetStr := fmt.Sprintf("%v", targetValue)

	return strings.Contains(fieldStr, targetStr)
}

// isInValue 检查值是否在指定列表中
func (p *TextTemplateParser) isInValue(fieldValue, listValue interface{}) bool {
	if fieldValue == nil || listValue == nil {
		return false
	}

	// 如果listValue是切片，检查fieldValue是否在其中
	if list, ok := listValue.([]interface{}); ok {
		for _, item := range list {
			if item == fieldValue {
				return true
			}
		}
	}

	return false
}

// isEmptyOrZero 判断值是否为空或零
func (p *TextTemplateParser) isEmptyOrZero(value interface{}) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return v == "" || v == "0" || v == "0.0" || v == "false"
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float32, float64:
		return v == 0.0
	case bool:
		return !v
	case []interface{}:
		return len(v) == 0
	case []map[string]interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		// 对于其他类型，尝试转换为字符串判断
		str := fmt.Sprintf("%v", v)
		return str == "" || str == "0" || str == "0.0" || str == "false"
	}
}

// replaceVariables 替换文本中的变量
func (p *TextTemplateParser) replaceVariables(text string) string {
	if text == "" || p.data == nil {
		return text
	}
	// 使用正则表达式匹配 {variable.path} 格式的变量
	re := regexp.MustCompile(`\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		// 提取变量名（去掉花括号）
		variableName := match[1 : len(match)-1]
		// 从数据源获取变量值
		value := p.getDataValue(variableName)
		// 格式化值
		if value == nil {
			return match // 如果找不到值，保持原样
		}
		// 根据值的类型进行格式化
		switch v := value.(type) {
		case string:
			return v
		case int, int8, int16, int32, int64:
			return fmt.Sprintf("%d", v)
		case uint, uint8, uint16, uint32, uint64:
			return fmt.Sprintf("%d", v)
		case float32, float64:
			// 使用 %g 自动去除不必要的小数位零
			return fmt.Sprintf("%g", v)
		case bool:
			return fmt.Sprintf("%t", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	})
}

// getData 从数据源获取值
func (p *TextTemplateParser) getData(path string) []map[string]interface{} {
	if p.data == nil {
		return nil
	}
	keys := strings.Split(path, ".")
	var current interface{} = p.data
	for _, key := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[key]
		case map[interface{}]interface{}:
			current = v[key]
		}
	}
	// 处理不同类型的数据
	switch data := current.(type) {
	case map[string]interface{}:
		// 如果是map，返回包含一个元素的map数组
		return []map[string]interface{}{data}
	case []interface{}:
		// 如果是数组，遍历每个元素，收集所有map[string]interface{}
		var result []map[string]interface{}
		for _, item := range data {
			if itemMap, ok := item.(map[string]interface{}); ok {
				result = append(result, itemMap)
			}
		}
		return result
	case []map[string]interface{}:
		// 如果本身就是map数组，直接返回
		return data
	default:
		// 其他类型，跳过处理
		return nil
	}
}

// getDataValue 从数据源获取值
func (p *TextTemplateParser) getDataValue(path string) interface{} {
	if p.data == nil {
		return nil
	}

	// 支持点号分隔的路径
	keys := strings.Split(path, ".")
	var current interface{} = p.data
	for _, key := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[key]
		case map[interface{}]interface{}:
			current = v[key]
		default:
			// 尝试使用反射获取字段值
			rv := reflect.ValueOf(current)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			if rv.Kind() == reflect.Struct {
				field := rv.FieldByName(strings.Title(key))
				if field.IsValid() {
					current = field.Interface()
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
		if current == nil {
			return nil
		}
	}

	return current
}

// formatValue 格式化数据值
func (p *TextTemplateParser) formatValue(value interface{}, attr TextTemplateBlockAttr) string {
	if value == nil {
		return attr.DefaultValue
	}
	// 转换为字符串
	valueStr := fmt.Sprintf("%v", value)
	if attr.ShowCurrencyUnit && p.baseData.CurrencyUnit != "" {
		if p.template.Metadata.Thousandth {
			// 安全地转换为float64
			if floatVal := p.convertToFloat64(value); floatVal != 0 {
				valueStr = p.Thousandth(floatVal)
			}
		}
		if p.baseData.CurrencyUnitPosition == 1 {
			valueStr = valueStr + p.baseData.CurrencyUnit
		} else {
			valueStr = p.baseData.CurrencyUnit + valueStr
		}
	}
	return valueStr
}

// checkConditions 检查条件显示规则
func (p *TextTemplateParser) checkConditions(conditions []TextTemplateCondition) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, condition := range conditions {
		if !p.checkCondition(condition) {
			return false
		}
	}

	return true
}

// applyFontStyle 应用字体样式（适配ESC打印机）
func (p *TextTemplateParser) applyFontStyle(printer *Printers, attr TextTemplateBlockAttr) {
	// ESC打印机字体样式设置
	fontWeight := p.getFontWeight(attr)

	// 设置字体大小和样式
	if attr.FontSize > 20 || fontWeight > 1 {
		// 大字体或粗体，使用字符放大
		doubleH := attr.FontSize > 25
		doubleW := attr.FontSize > 30
		bold := fontWeight > 1
		printer.SetPrintModes(bold, doubleH, doubleW)
	} else {
		// 普通字体
		printer.SetPrintModes(false, false, false)
	}
}

// getFontWeight 获取字体粗细
func (p *TextTemplateParser) getFontWeight(attr TextTemplateBlockAttr) int {
	if attr.FontWeight > 0 {
		return attr.FontWeight
	}

	if attr.FontBold {
		return 2
	}

	return 1
}

// convertAlign 转换对齐方式
func (p *TextTemplateParser) convertAlign(align string) int {
	switch align {
	case "left":
		return AlignLeft
	case "center":
		return AlignCenter
	case "right":
		return AlignRight
	case "left-right":
		return AlignLeft // 特殊处理，在调用方处理
	default:
		return AlignLeft
	}
}

// ValidateTemplate 验证模板配置
func (p *TextTemplateParser) ValidateTemplate() error {
	// 验证模板元数据
	if p.template.Metadata.Name == "" {
		return fmt.Errorf("模板名称不能为空")
	}

	if p.template.Metadata.PaperWidth <= 0 {
		return fmt.Errorf("纸张宽度必须大于0")
	}

	// 验证模板行
	return p.validateRows(p.template.Rows)
}

// validateRows 验证模板行
func (p *TextTemplateParser) validateRows(rows [][]TextTemplateBlock) error {
	for i, row := range rows {
		err := p.validateRow(row, i)
		if err != nil {
			return err
		}

		// 验证嵌套行
		for _, block := range row {
			if len(block.Rows) > 0 {
				err := p.validateRows(block.Rows)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// validateRow 验证单行
func (p *TextTemplateParser) validateRow(blocks []TextTemplateBlock, rowIndex int) error {
	totalWidth := 0.0

	for j, block := range blocks {
		// 验证块ID唯一性
		if block.BlockID == "" {
			return fmt.Errorf("第%d行第%d个块的ID不能为空", rowIndex+1, j+1)
		}
		// 验证宽度
		if block.BlockType != "Text" && block.BlockType != "qrcode" && block.BlockType != "barcode" {
			width := p.getWidthValue(block.BlockAttr.Width)
			if width < 0 || width > 100 {
				return fmt.Errorf("块 '%s' 的宽度必须在0-100之间", block.BlockID)
			}
			totalWidth += width
		}
		// 验证字体大小
		if block.BlockAttr.FontSize < 0 || block.BlockAttr.FontSize > 48 {
			return fmt.Errorf("块 '%s' 的字体大小必须在0-48之间", block.BlockID)
		}
	}

	// 验证总宽度不超过100%
	if totalWidth > 100 {
		return fmt.Errorf("第%d行所有块的总宽度(%.1f%%)超过100%%", rowIndex+1, totalWidth)
	}

	return nil
}
