package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"ttpos-ai/config"
	"ttpos-ai/pkg/database"
	"ttpos-ai/pkg/logger"

	"go.uber.org/zap"
)

// SQLTool SQL查询工具
type SQLTool struct {
	maxRows     int
	tablePrefix string
	denyTables  []string
	allowTables []string
}

// NewSQLTool 创建SQL工具
func NewSQLTool() *SQLTool {
	return &SQLTool{
		maxRows:     config.Database.MaxRows,
		tablePrefix: config.Database.TablePrefix,
		denyTables:  config.Database.DenyTables,
		allowTables: config.Database.AllowTables,
	}
}

// SQLQueryInput SQL查询输入参数
type SQLQueryInput struct {
	SQL string `json:"sql" description:"要执行的SQL查询语句，只支持SELECT"`
}

// SQLQueryOutput SQL查询输出
type SQLQueryOutput struct {
	Success bool              `json:"success"`
	Data    []json.RawMessage `json:"data,omitempty"`
	Count   int               `json:"count"`
	Error   string            `json:"error,omitempty"`
}

// Execute 执行SQL查询
func (t *SQLTool) Execute(ctx context.Context, input *SQLQueryInput) (*SQLQueryOutput, error) {
	logger.Logger.Info("执行SQL查询", zap.String("sql", input.SQL))

	// 安全校验
	if err := t.validateSQL(input.SQL); err != nil {
		logger.Logger.Warn("SQL校验失败", zap.Error(err), zap.String("sql", input.SQL))
		return &SQLQueryOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 添加LIMIT限制
	sql := t.addLimit(input.SQL)

	// 执行查询
	db := database.GetDB()
	if db == nil {
		return &SQLQueryOutput{
			Success: false,
			Error:   "数据库未连接",
		}, nil
	}

	var results []map[string]any
	if err := db.Raw(sql).Scan(&results).Error; err != nil {
		logger.Logger.Error("SQL执行失败", zap.Error(err), zap.String("sql", sql))
		return &SQLQueryOutput{
			Success: false,
			Error:   fmt.Sprintf("查询失败: %v", err),
		}, nil
	}

	// 转换结果
	data := make([]json.RawMessage, 0, len(results))
	for _, row := range results {
		rowJSON, _ := json.Marshal(row)
		data = append(data, rowJSON)
	}

	logger.Logger.Info("SQL查询完成", zap.Int("count", len(results)))

	return &SQLQueryOutput{
		Success: true,
		Data:    data,
		Count:   len(results),
	}, nil
}

// validateSQL 验证SQL安全性
func (t *SQLTool) validateSQL(sql string) error {
	sql = strings.TrimSpace(sql)
	upperSQL := strings.ToUpper(sql)

	// 只允许SELECT语句
	if !strings.HasPrefix(upperSQL, "SELECT") {
		return fmt.Errorf("只允许SELECT查询")
	}

	// 禁止危险关键字
	dangerousKeywords := []string{
		"INSERT", "UPDATE", "DELETE", "DROP", "TRUNCATE",
		"ALTER", "CREATE", "GRANT", "REVOKE", "EXECUTE",
		"INTO OUTFILE", "INTO DUMPFILE", "LOAD_FILE",
		"BENCHMARK", "SLEEP", "WAITFOR",
	}
	for _, keyword := range dangerousKeywords {
		if strings.Contains(upperSQL, keyword) {
			return fmt.Errorf("禁止使用关键字: %s", keyword)
		}
	}

	// 禁止子查询修改数据
	if strings.Contains(upperSQL, ";") {
		return fmt.Errorf("禁止多语句执行")
	}

	// 检查表名黑名单
	tables := t.extractTables(sql)
	for _, table := range tables {
		tableName := strings.TrimPrefix(table, t.tablePrefix)
		for _, deny := range t.denyTables {
			if strings.EqualFold(tableName, deny) || strings.EqualFold(table, deny) {
				return fmt.Errorf("禁止查询表: %s", table)
			}
		}
	}

	// 如果设置了白名单，检查是否在白名单内
	if len(t.allowTables) > 0 {
		for _, table := range tables {
			tableName := strings.TrimPrefix(table, t.tablePrefix)
			allowed := false
			for _, allow := range t.allowTables {
				if strings.EqualFold(tableName, allow) || strings.EqualFold(table, allow) {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("表不在允许列表中: %s", table)
			}
		}
	}

	return nil
}

// extractTables 从SQL中提取表名
func (t *SQLTool) extractTables(sql string) []string {
	tables := make([]string, 0)

	// 匹配 FROM table 和 JOIN table
	patterns := []string{
		`(?i)FROM\s+` + "`?" + `(\w+)` + "`?",
		`(?i)JOIN\s+` + "`?" + `(\w+)` + "`?",
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(sql, -1)
		for _, match := range matches {
			if len(match) > 1 {
				tables = append(tables, match[1])
			}
		}
	}

	return tables
}

// addLimit 添加LIMIT限制
func (t *SQLTool) addLimit(sql string) string {
	upperSQL := strings.ToUpper(sql)

	// 如果已有LIMIT，检查是否超过限制
	if strings.Contains(upperSQL, "LIMIT") {
		// 替换过大的LIMIT
		re := regexp.MustCompile(`(?i)LIMIT\s+(\d+)`)
		sql = re.ReplaceAllStringFunc(sql, func(match string) string {
			return fmt.Sprintf("LIMIT %d", t.maxRows)
		})
		return sql
	}

	// 添加LIMIT
	sql = strings.TrimSuffix(strings.TrimSpace(sql), ";")
	return fmt.Sprintf("%s LIMIT %d", sql, t.maxRows)
}

// GetToolInfo 获取工具信息（用于LLM）
func (t *SQLTool) GetToolInfo() map[string]any {
	return map[string]any{
		"name":        "execute_sql",
		"description": "执行SQL查询语句，用于查询数据库中的数据。只支持SELECT语句，禁止修改数据。表名需要加上前缀 ttpos_",
		"parameters": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"sql": map[string]any{
					"type":        "string",
					"description": "要执行的SQL SELECT语句",
				},
			},
			"required": []string{"sql"},
		},
	}
}

// GetSchemaInfo 获取数据库Schema信息
func (t *SQLTool) GetSchemaInfo(ctx context.Context) (string, error) {
	db := database.GetDB()
	if db == nil {
		return "", fmt.Errorf("数据库未连接")
	}

	// 获取所有表
	var tables []struct {
		TableName string `gorm:"column:TABLE_NAME"`
	}
	if err := db.Raw(`
		SELECT TABLE_NAME
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`, config.Database.Database).Scan(&tables).Error; err != nil {
		return "", err
	}

	var schema strings.Builder
	schema.WriteString("数据库表结构:\n\n")

	for _, table := range tables {
		// 检查是否在黑名单
		tableName := strings.TrimPrefix(table.TableName, t.tablePrefix)
		isDenied := false
		for _, deny := range t.denyTables {
			if strings.EqualFold(tableName, deny) {
				isDenied = true
				break
			}
		}
		if isDenied {
			continue
		}

		// 获取表的列信息
		var columns []struct {
			ColumnName string `gorm:"column:COLUMN_NAME"`
			ColumnType string `gorm:"column:COLUMN_TYPE"`
			Comment    string `gorm:"column:COLUMN_COMMENT"`
		}
		db.Raw(`
			SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_COMMENT
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION
		`, config.Database.Database, table.TableName).Scan(&columns)

		schema.WriteString(fmt.Sprintf("表: %s\n", table.TableName))
		for _, col := range columns {
			comment := col.Comment
			if comment == "" {
				comment = "-"
			}
			schema.WriteString(fmt.Sprintf("  - %s (%s): %s\n", col.ColumnName, col.ColumnType, comment))
		}
		schema.WriteString("\n")
	}

	return schema.String(), nil
}
