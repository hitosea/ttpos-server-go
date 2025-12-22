// ImgTemplateParser 模板解析器
// @author: wfs
// @date: 2024/12/20
package pkg

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

type ImgBaseData struct {
	Language             string `json:"language"`               // 语言
	CurrencyUnit         string `json:"currency_unit"`          // 货币单位
	CurrencyUnitPosition int    `json:"currency_unit_position"` // 货币单位位置
	Is58mmPrinter        bool   `json:"is_58mm_printer"`        // 是否58mm打印机
}

// ImgTemplateMetadata 模板元数据
type ImgTemplateMetadata struct {
	Name        string `json:"name"`        // 模板名称
	Description string `json:"description"` // 模板描述
	PaperWidth  int    `json:"paper_width"` // 纸张宽度（毫米）
	Version     string `json:"version"`     // 模板版本
	Thousandth  bool   `json:"thousandth"`  // 是否千分位
}

// ImgTemplateBlock 模板块定义
type ImgTemplateBlock struct {
	BlockID           string                   `json:"block_id"`            // 块唯一标识
	BlockType         string                   `json:"block_type"`          // 块类型 label, value, label:auto:value, column, img, qrcode, barcode, blank_line
	BlockLabel        interface{}              `json:"block_label"`         // 标签（支持字符串或多语言）
	BlockValue        interface{}              `json:"block_value"`         // 块值
	BlockBeforeLabel  interface{}              `json:"block_before_label"`  // 块前标签（支持字符串或多语言）
	BlockAfterLabel   interface{}              `json:"block_after_label"`   // 块后标签（支持字符串或多语言）
	BlockExpandLabels []ImgTemplateExpandLabel `json:"block_expand_labels"` // 块扩展标签（支持条件显示）
	BlockBrotherID    string                   `json:"block_brother_id"`    // 兄弟块ID
	BlockAttr         ImgTemplateBlockAttr     `json:"block_attr"`          // 块属性
	Rows              [][]ImgTemplateBlock     `json:"rows,omitempty"`      // 嵌套行
	Conditions        []ImgTemplateCondition   `json:"conditions"`          // 条件显示规则
}

// ImgTemplateBlockAttr 块属性定义
type ImgTemplateBlockAttr struct {
	// 样式配置
	FontSize   int    `json:"font_size"`   // 字体大小
	FontBold   bool   `json:"font_bold"`   // 是否粗体
	FontWeight int    `json:"font_weight"` // 字体粗细 (1-3)
	Align      string `json:"align"`       // 对齐方式

	// 布局配置
	Width              interface{} `json:"width"`                // 宽度百分比（支持固定值或按语言动态配置）
	Width58            interface{} `json:"width_58"`             // 58mm宽度百分比
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
	CanEditValue       bool        `json:"can_edit_value"`       // 是否可编辑值
	Disabled           bool        `json:"disabled"`             // 是否禁用

	// 数据绑定
	DefaultValue string `json:"default_value"` // 默认值
}

// ImgTemplateCondition 条件显示规则
type ImgTemplateCondition struct {
	Field    string      `json:"field"`    // 字段路径
	Operator string      `json:"operator"` // 操作符
	Value    interface{} `json:"value"`    // 比较值
}

// ImgTemplateExpandLabel 扩展标签定义
type ImgTemplateExpandLabel struct {
	Text       interface{}            `json:"text"`       // 标签文本（支持字符串或多语言）
	Conditions []ImgTemplateCondition `json:"conditions"` // 显示条件
}

// ImgTemplate JSON模板定义
type ImgTemplate struct {
	Metadata ImgTemplateMetadata  `json:"metadata"` // 模板元数据
	Rows     [][]ImgTemplateBlock `json:"rows"`     // 模板行
}

// ImgTemplateParser JSON模板解析器
type ImgTemplateParser struct {
	template *ImgTemplate           // 模板配置
	data     map[string]interface{} // 数据源
	baseData ImgBaseData            // 基础数据
}

// NewImgTemplateParser 创建新的JSON模板解析器
func NewImgTemplateParser(baseData ImgBaseData, templateJSON string, data map[string]interface{}) (*ImgTemplateParser, error) {
	var template ImgTemplate
	err := json.Unmarshal([]byte(templateJSON), &template)
	if err != nil {
		return nil, fmt.Errorf("解析模板JSON失败: %v", err)
	}

	// 如果58mm打印机，且纸张宽度为0，则设置为58mm
	if baseData.Is58mmPrinter && template.Metadata.PaperWidth == 0 {
		template.Metadata.PaperWidth = 58
	}

	// 设置默认语言
	if baseData.Language == "" {
		baseData.Language = "zh"
	}

	return &ImgTemplateParser{
		template: &template,
		data:     data,
		baseData: baseData,
	}, nil
}

// Parse 解析模板并生成打印内容
func (p *ImgTemplateParser) Parse() (*ImgFont, error) {
	defer func() {
		if r := recover(); r != nil {
			// 捕获 panic 并记录日志
			logger.Logger.Error("解析模板并生成打印内容 panic", zap.Any("panic_value", r), zap.Stack("stack_trace"))
		}
	}()
	// 根据纸张宽度创建图片打印对象
	var img *ImgFont

	switch p.template.Metadata.PaperWidth {
	case 80:
		img = NewImgFont(568, 0, 0) // 80mm纸张
	case 58:
		img = NewImgFont(384, 50, 0) // 其他规格
	default:
		if p.baseData.Is58mmPrinter {
			img = NewImgFont(384, 50, 0) // 其他规格
		} else {
			img = NewImgFont(568, 0, 0) // 80mm纸张
		}
	}

	// 设置内边距
	img.SetImagePadding(0)

	// 解析模板行
	err := p.parseRows(img, p.template.Rows, 0, false)
	if err != nil {
		return nil, fmt.Errorf("解析模板行失败: %v", err)
	}

	return img, nil
}

// Thousandth 计算金额的千分位
func (p *ImgTemplateParser) Thousandth(amount float64) string {
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
func (p *ImgTemplateParser) parseRows(img *ImgFont, rows [][]ImgTemplateBlock, level int, isLast bool) error {
	for i, row := range rows {
		err := p.parseRow(img, row, p.getNextValidRow(rows, i+1), level, isLast)
		if err != nil {
			return fmt.Errorf("解析行失败: %v", err)
		}
	}
	return nil
}

// getNextValidBlock 获取下一行有效的block
func (p *ImgTemplateParser) validBlock(block ImgTemplateBlock) bool {
	// 如果设置了不显示空值，且值为空或零，则返回空字符串
	if block.BlockAttr.NotShowEmpty {
		if block.BlockType == "label" {
			if p.getLabel(block.BlockLabel) == "" {
				return false
			}
		} else if block.BlockType == "label:value" && block.BlockValue != nil {
			if p.getLabel(block.BlockValue) == "" {
				return false
			}
		} else if p.isEmptyOrZero(p.getDataValue(block.BlockID)) {
			return false
		}
	}
	// 检查条件显示
	if !p.checkConditions(block.Conditions) {
		return false
	}
	return true
}

// getNextValidRow 获取下一行有效的blocks
func (p *ImgTemplateParser) getNextValidRow(rows [][]ImgTemplateBlock, startIndex int) []ImgTemplateBlock {
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
	return []ImgTemplateBlock{}
}

// parseRow 解析单行
func (p *ImgTemplateParser) parseRow(
	img *ImgFont,
	blocks []ImgTemplateBlock,
	nextBlocks []ImgTemplateBlock,
	level int,
	isLast bool,
) error {
	if len(blocks) == 0 {
		return nil
	}

	// 获取下一行的 blocks 作为 nextBlocks
	var nextBlock ImgTemplateBlock
	if len(nextBlocks) > 0 {
		nextBlock = nextBlocks[0]
	}

	block := blocks[0]

	if block.BlockAttr.Disabled || !p.validBlock(block) {
		return nil
	}

	// 设置前置空行
	if block.BlockAttr.LeadingBlankLines > 0 {
		for i := 0; i < block.BlockAttr.LeadingBlankLines; i++ {
			img.LineFeed(block.BlockAttr.LeadingBlankLines, 12)
		}
	}

	// 设置字体样式
	p.applyFontStyle(img, block.BlockAttr)

	// 设置对齐方式
	img.SetAlignment(p.convertAlign(block.BlockAttr.Align))

	// 如果块ID不是空白行，则处理
	if block.BlockID != "blank_line" {

		// 如果只有一个块，直接处理
		if len(blocks) == 1 && block.BlockType != "column" {
			// 获取宽度
			width := p.getWidthValue(block.BlockAttr.Width)
			if p.template.Metadata.PaperWidth == 58 && block.BlockAttr.Width58 != nil && block.BlockAttr.Width58 != 0 {
				if width58 := p.getWidthValue(block.BlockAttr.Width58); width58 != 0 {
					width = width58
				}
			}
			widthInt := int(width)
			text := p.getBlockText(block)

			// 添加文本
			if block.BlockType == "img" {
				if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
					// 使用永久保存，下次可以直接使用缓存
					tempPath, downloadErr := utils.DownloadImageToLocal(text, true)
					if downloadErr != nil {
						return fmt.Errorf("下载图片失败: %v, URL: %s\n", downloadErr, text)
					}
					text = tempPath
				}
				img.SetTextLineHeight(25)
				img.AppendImg(text, widthInt, false, 0)
				img.RecoverDefaultTextLineHeight()
			} else if block.BlockType == "qrcode" {
				if strings.HasPrefix(text, "data:image/png;base64,") {
					img.SetTextTotalHeight(20)
					img.SetTextLineHeight(5)
					img.AppendQrcodeDisableBorder(text, widthInt, 0, true)
				} else if strings.HasPrefix(text, "./tmp") {
					img.SetTextLineHeight(20)
					img.AppendImg(text, widthInt, false, 0)
				} else {
					widthInt = widthInt - 20
					img.SetTextTotalHeight(0)
					img.SetTextLineHeight(65)
					img.AppendQrcodeDisableBorder(text, widthInt, 0, false)
				}
				img.RecoverDefaultTextLineHeight()
			} else if block.BlockType == "barcode" {
				img.AppendBarcode(text, widthInt, 120)
				img.LineFeed(1, 12)
				img.SetAlignment(AlignCenter)
				img.AppendText(text)
				img.LineFeed(1)
				img.RecoverDefaultTextLineHeight()
			} else if strings.Contains(block.BlockType, "label:auto:value") {
				fontWeight := p.getFontWeight(block.BlockAttr)
				lineHeight := block.BlockAttr.LineHeight
				rightPadding := block.BlockAttr.PaddingRight
				img.PrintInColumns(
					ColumnConfig{Text: p.getLabel(block.BlockLabel), Width: p.getBlockWidth(img, block) - rightPadding - 50, Align: AlignLeft, FontWeight: 1, LineHeight: lineHeight, RightPadding: rightPadding},
					ColumnConfig{Text: p.getBlockText(block), Width: 0, Align: AlignRight, FontWeight: fontWeight, LineHeight: lineHeight, RightPadding: rightPadding},
				)
			} else if !strings.Contains(block.BlockType, "array") {
				// 获取显示文本
				img.AppendText(text)
				// 添加换行
				if len(text) > 0 {
					// 如果下一行是空白行，则缩小间隔换行，否则正常换行
					if nextBlock.BlockType == "blank_line" {
						img.LineFeed(1, 40)
					} else {
						img.LineFeed(1)
					}
				}
			}

		} else if (len(block.Rows) == 0 || level > 0) && block.BlockType != "column" {
			// 创建列配置
			columns := make([]ColumnConfig, 0, len(blocks))
			for _, block := range blocks {
				// 检查条件显示
				if block.BlockAttr.Disabled || !p.checkConditions(block.Conditions) {
					continue
				}
				lineHeight := block.BlockAttr.LineHeight
				if isLast && lineHeight > 33 {
					lineHeight = 33
				}
				column := ColumnConfig{
					Text:         p.getBlockText(block),
					Width:        p.getBlockWidth(img, block), // 转换为像素宽度
					Align:        p.convertAlign(block.BlockAttr.Align),
					FontWeight:   p.getFontWeight(block.BlockAttr),
					LineHeight:   lineHeight,
					RightPadding: block.BlockAttr.PaddingRight,
				}
				columns = append(columns, column)
			}
			// 如果有有效的列，进行分列打印
			if len(columns) > 0 {
				img.PrintInColumns(columns...)
			}

			// 处理嵌套行 - 规格，备注等
			if level > 0 {
				for _, block := range blocks {
					if len(block.Rows) > 0 {
						err := p.parseRows(img, block.Rows, level+1, isLast)
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
						err := p.parseRows(img, block.Rows, level+1, isLast)
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
	if block.BlockAttr.DividingLine || (block.BlockType == "blank_line" && block.BlockAttr.DividingLine) {
		img.AppendSplitLine(WithLineFeed(true))
	}

	// 设置后置空行
	if block.BlockAttr.TrailingBlankLines > 0 {
		for i := 0; i < block.BlockAttr.TrailingBlankLines; i++ {
			img.LineFeed(block.BlockAttr.TrailingBlankLines, 12)
		}
	}

	// 恢复默认值
	img.RestoreDefault()

	return nil
}

// getBlockWidth 获取块的宽度
func (p *ImgTemplateParser) getBlockWidth(img *ImgFont, block ImgTemplateBlock) int {
	width := p.getWidthValue(block.BlockAttr.Width)
	if p.template.Metadata.PaperWidth == 58 && block.BlockAttr.Width58 != nil && block.BlockAttr.Width58 != 0 {
		if width58 := p.getWidthValue(block.BlockAttr.Width58); width58 != 0 {
			width = width58
		}
	}
	if width > 0 {
		return int(float64(img.ImageWidth) * (width / 100))
	}
	return 0
}

// getWidthValue 获取宽度值（支持固定值或按语言动态配置）
func (p *ImgTemplateParser) getWidthValue(width interface{}) float64 {
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
func (p *ImgTemplateParser) convertToFloat64(value interface{}) float64 {
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
	case uint:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
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
func (p *ImgTemplateParser) getBlockText(block ImgTemplateBlock) string {
	// 如果有数据源，检查是否需要隐藏空值
	if block.BlockID != "" && (strings.Contains(block.BlockType, "value") || block.BlockType == "img" || block.BlockType == "qrcode" || block.BlockType == "barcode") {
		value := p.getDataValue(block.BlockID)
		if block.BlockValue != nil {
			// 支持多语言的 BlockValue
			value = p.getLabel(block.BlockValue)
		}
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
func (p *ImgTemplateParser) GetBlockText(block ImgTemplateBlock) string {
	return p.getBlockText(block)
}

// SetLanguage 设置语言
func (p *ImgTemplateParser) SetLanguage(language string) {
	p.baseData.Language = language
}

// GetBlockWidth 公开方法，获取块的宽度
func (p *ImgTemplateParser) GetBlockWidth(img *ImgFont, block ImgTemplateBlock) int {
	return p.getBlockWidth(img, block)
}

// getLabel 获取当前语言的标签
func (p *ImgTemplateParser) getLabel(labels interface{}) string {
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
		// 处理标签中的变量替换
		return p.replaceVariables(label)
	}

	// 如果是map[string]string类型（向后兼容）
	if labelsMap, ok := labels.(map[string]string); ok {
		var label string
		// 获取当前语言的标签
		if l, exists := labelsMap[p.baseData.Language]; exists && l != "" {
			label = l
		}
		// 处理标签中的变量替换
		return p.replaceVariables(label)
	}

	return ""
}

// getLabelText 获取标签文本（支持变量替换）
func (p *ImgTemplateParser) getLabelText(labels interface{}) string {
	if labels == nil {
		return ""
	}

	// 使用 getLabel 方法获取当前语言的标签
	return p.getLabel(labels)
}

// getExpandLabelsText 获取扩展标签文本（支持条件显示）
func (p *ImgTemplateParser) getExpandLabelsText(expandLabels []ImgTemplateExpandLabel) string {
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
func (p *ImgTemplateParser) combineLabels(before, main, after string) string {
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
func (p *ImgTemplateParser) combineLabelsWithExpand(before, main, after, expand string) string {
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
func (p *ImgTemplateParser) checkCondition(condition ImgTemplateCondition) bool {
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
func (p *ImgTemplateParser) compareValues(a, b interface{}) int {
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
func (p *ImgTemplateParser) containsValue(fieldValue, targetValue interface{}) bool {
	if fieldValue == nil || targetValue == nil {
		return false
	}

	fieldStr := fmt.Sprintf("%v", fieldValue)
	targetStr := fmt.Sprintf("%v", targetValue)

	return strings.Contains(fieldStr, targetStr)
}

// isInValue 检查值是否在指定列表中
func (p *ImgTemplateParser) isInValue(fieldValue, listValue interface{}) bool {
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
func (p *ImgTemplateParser) isEmptyOrZero(value interface{}) bool {
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
func (p *ImgTemplateParser) replaceVariables(text string) string {
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
func (p *ImgTemplateParser) getData(path string) []map[string]interface{} {
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
func (p *ImgTemplateParser) getDataValue(path string) interface{} {
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
func (p *ImgTemplateParser) formatValue(value interface{}, attr ImgTemplateBlockAttr) string {
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
func (p *ImgTemplateParser) checkConditions(conditions []ImgTemplateCondition) bool {
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

// applyFontStyle 应用字体样式
func (p *ImgTemplateParser) applyFontStyle(img *ImgFont, attr ImgTemplateBlockAttr) {
	if attr.FontSize > 0 {
		img.SetFontSize(attr.FontSize)
	}

	fontWeight := p.getFontWeight(attr)
	if fontWeight > 0 {
		img.SetFontWeight(fontWeight)
	}

	if attr.LineHeight > 0 {
		img.SetTextLineHeight(attr.LineHeight)
	}
}

// getFontWeight 获取字体粗细
func (p *ImgTemplateParser) getFontWeight(attr ImgTemplateBlockAttr) int {
	if attr.FontWeight > 0 {
		return attr.FontWeight
	}

	if attr.FontBold {
		return 2
	}

	return 1
}

// convertAlign 转换对齐方式
func (p *ImgTemplateParser) convertAlign(align string) int {
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
func (p *ImgTemplateParser) ValidateTemplate() error {
	// 验证模板元数据
	if p.template.Metadata.Name == "" {
		return fmt.Errorf("模板名称不能为空")
	}
	// 验证模板行
	return p.validateRows(p.template.Rows)
}

// validateRows 验证模板行
func (p *ImgTemplateParser) validateRows(rows [][]ImgTemplateBlock) error {
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
func (p *ImgTemplateParser) validateRow(blocks []ImgTemplateBlock, rowIndex int) error {
	totalWidth := 0.0

	for j, block := range blocks {
		// 验证块ID唯一性
		if block.BlockID == "" {
			return fmt.Errorf("第%d行第%d个块的ID不能为空", rowIndex+1, j+1)
		}
		// 验证宽度
		if block.BlockType != "img" && block.BlockType != "qrcode" && block.BlockType != "barcode" {
			width := p.getWidthValue(block.BlockAttr.Width)
			if p.template.Metadata.PaperWidth == 58 && block.BlockAttr.Width58 != nil {
				if width58 := p.getWidthValue(block.BlockAttr.Width58); width58 != 0 {
					width = width58
				}
			}
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

	// 验证总宽度必须大于0
	if totalWidth < 0 {
		return fmt.Errorf("第%d行所有块的总宽度必须大于0", rowIndex+1)
	}

	return nil
}
