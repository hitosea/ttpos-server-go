package template

import (
	"embed"
	"fmt"
)

//go:embed statement_order_config.json
//go:embed statement_order_data.json
//go:embed statement_order_tmp.json
//go:embed statement_pre_config.json
//go:embed statement_pre_data.json
//go:embed statement_pre_tmp.json
//go:embed invoice_config.json
//go:embed invoice_data.json
//go:embed invoice_tmp.json
//go:embed takeout_merchant_receipt_config.json
//go:embed takeout_merchant_receipt_data.json
//go:embed takeout_merchant_receipt_tmp.json
//go:embed takeout_customer_receipt_config.json
//go:embed takeout_customer_receipt_data.json
//go:embed takeout_customer_receipt_tmp.json
//go:embed takeout_refund_receipt_config.json
//go:embed takeout_refund_receipt_data.json
//go:embed takeout_refund_receipt_tmp.json
var TemplateJsonFS embed.FS

// StatementOrderTemplateType 结账单模板类型枚举
type StatementOrderTemplateType string

const (
	StatementOrder         StatementOrderTemplateType = "statement_order"          // StatementOrder 结账单
	PreStatementOrder      StatementOrderTemplateType = "statement_pre"            // PreStatementOrder 预结账单
	Invoice                StatementOrderTemplateType = "invoice"                  // Invoice 发票
	TakeoutMerchantReceipt StatementOrderTemplateType = "takeout_merchant_receipt" // TakeoutMerchantReceipt 外卖商家联
	TakeoutCustomerReceipt StatementOrderTemplateType = "takeout_customer_receipt" // TakeoutCustomerReceipt 外卖顾客联
	TakeoutRefundReceipt   StatementOrderTemplateType = "takeout_refund_receipt"   // TakeoutRefundReceipt 外卖退单联
)

// 中文名称到枚举的映射表（用于从数据库中文名称查找）
var chineseNameMap = map[string]StatementOrderTemplateType{
	"结账单":   StatementOrder,
	"预结账单":  PreStatementOrder,
	"发票":    Invoice,
	"外卖商家联": TakeoutMerchantReceipt,
	"外卖顾客联": TakeoutCustomerReceipt,
	"外卖退单联": TakeoutRefundReceipt,
}

// IsValid 验证枚举值是否有效
func (t StatementOrderTemplateType) IsValid() bool {
	return t == StatementOrder || t == PreStatementOrder || t == Invoice || t == TakeoutMerchantReceipt || t == TakeoutCustomerReceipt || t == TakeoutRefundReceipt
}

// ParseFromString 从字符串解析模板类型（支持中文和英文名称）
// 主要用于从数据库中的中文名称查找对应的英文文件
// name: 模板名称，可以是中文（"结账单"、"预结账单"）或英文（"statement_order"、"statement_pre"）
func ParseFromString(name string) StatementOrderTemplateType {
	// 优先从中文名称映射表查找（数据库通常存储中文名称）
	if templateType, ok := chineseNameMap[name]; ok {
		return templateType
	}
	// 未找到匹配的类型，返回空值
	return ""
}

// GetTemplateFileNameFromChineseName 根据中文名称直接获取英文文件名（便捷方法）
// chineseName: 数据库中的中文名称（如"结账单"、"预结账单"）
// 返回: 英文文件名（不含扩展名），如果未找到则返回空字符串
func GetTemplateFileNameFromChineseName(chineseName string) string {
	templateType := ParseFromString(chineseName)
	if !templateType.IsValid() {
		return ""
	}
	return string(templateType)
}

// GetTemplateJsonData 获取模板JSON数据
// templateType: 模板类型枚举
func GetTemplateJsonData(templateName string) ([]byte, error) {
	fileName := GetTemplateFileNameFromChineseName(templateName) + "_tmp.json"
	data, err := TemplateJsonFS.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("读取模板文件失败 [%s]: %w", fileName, err)
	}
	return data, nil
}

// GetTemplateConfigJsonData 获取模板配置JSON数据
// templateType: 模板类型枚举
func GetTemplateConfigJsonData(templateName string) ([]byte, error) {
	fileName := GetTemplateFileNameFromChineseName(templateName) + "_config.json"
	data, err := TemplateJsonFS.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("读取模板配置文件失败 [%s]: %w", fileName, err)
	}
	return data, nil
}

// GetTemplateDataJsonData 获取模板数据JSON数据
// templateType: 模板类型枚举
func GetTemplateDataJsonData(templateName string) ([]byte, error) {
	fileName := GetTemplateFileNameFromChineseName(templateName) + "_data.json"
	data, err := TemplateJsonFS.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("读取模板数据文件失败 [%s]: %w", fileName, err)
	}
	return data, nil
}
