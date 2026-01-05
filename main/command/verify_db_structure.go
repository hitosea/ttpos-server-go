package command

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func init() {
	verifyDBStructureCmd.Flags().StringVarP(&referenceSigFile, "reference", "r", "", "参考签名文件路径 (默认: admin/database/migrations/database_structure_signature.json)")
	rootCommand.AddCommand(verifyDBStructureCmd)
}

var (
	referenceSigFile string
)

// getProjectRoot 获取项目根目录
// 通过向上查找直到找到包含 admin 目录和 makefile 的目录来确定项目根目录
func getProjectRoot() (string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前工作目录失败: %w", err)
	}

	dir := workDir
	for {
		// 检查是否存在项目根目录的标志：同时存在 admin 目录和 makefile
		adminDir := filepath.Join(dir, "admin")
		makefilePath := filepath.Join(dir, "makefile")
		gitPath := filepath.Join(dir, ".git")

		// 如果同时存在 admin 目录和 makefile，或者存在 .git 目录，则认为是项目根目录
		hasAdmin, _ := os.Stat(adminDir)
		hasMakefile, _ := os.Stat(makefilePath)
		hasGit, _ := os.Stat(gitPath)

		if hasAdmin != nil && hasMakefile != nil && hasAdmin.IsDir() {
			return dir, nil
		}
		if hasGit != nil && hasGit.IsDir() {
			return dir, nil
		}

		// 到达文件系统根目录，停止查找
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// 如果找不到项目根目录，返回当前工作目录
	return workDir, nil
}

// 数据库结构签名
type DatabaseSignature struct {
	Signature   string           `json:"signature"`
	TableCount  int              `json:"table_count"`
	TotalFields int              `json:"total_fields"`
	Structure   []TableStructure `json:"structure"`
	Database    string           `json:"database,omitempty"`
	GeneratedAt string           `json:"generated_at,omitempty"`
}

// 表结构
type TableStructure struct {
	Table      string   `json:"table"`
	Fields     []string `json:"fields"`
	FieldCount int      `json:"field_count"`
}

// 字段信息
type FieldInfo struct {
	Name    string
	Type    string
	Default sql.NullString // 默认值（可能为NULL）
	Comment string         // 字段注释
}

// 验证数据库结构命令
var verifyDBStructureCmd = &cobra.Command{
	Use:   "verify-db-structure",
	Short: "验证商家数据库结构完整性",
	Long:  `验证商家数据库结构是否与参考签名一致，用于确认删除迁移文件后不会影响新商家`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("%sFailed to initialize config: %v%s", redColor, err, resetColor)
		}
		config.Server.Mode = "release"

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		// 初始化Redis分布式并发锁
		lock.InitRedisLock(cacheConfig)
		lock.NewSystemLock()

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}

		// 初始化id生成器
		utils.InitIdGenerator()

		// 读取参考签名文件路径
		if referenceSigFile == "" {
			// 默认路径：admin/database/migrations/database_structure_signature.json
			workDir, err := getProjectRoot()
			if err != nil {
				log.Fatalf("获取项目根目录失败: %v", err)
			}
			referenceSigFile = filepath.Join(workDir, "admin", "database", "migrations", "database_structure_signature.json")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 自动创建测试商户并验证
		testCompanyUuid, err := createTestCompany()
		if err != nil {
			fmt.Printf("%s 错误: 创建测试商户失败: %v %s\n", redColor, err, resetColor)
			return
		}
		defer cleanupTestCompany(testCompanyUuid) // 确保清理

		// 创建测试数据库
		dbName := fmt.Sprintf("%s%d", constant.DBNamePrefix, testCompanyUuid)
		if err := createTestDatabase(dbName); err != nil {
			fmt.Printf("%s 错误: 创建测试数据库失败: %v %s\n", redColor, err, resetColor)
			return
		}

		// 第一步：执行 shop_01.sql，读取参考签名
		fmt.Printf("\n%s 第一步：执行 shop_01.sql 并提取参考签名... %s\n", blueColor, resetColor)
		if err := executeShop01SQL(dbName); err != nil {
			fmt.Printf("%s 错误: 执行 shop_01.sql 失败: %v %s\n", redColor, err, resetColor)
			return
		}

		// 连接数据库并提取参考签名
		companyDB, err := database.NewMySQLConnection(config.Database, dbName)
		if err != nil {
			fmt.Printf("%s 错误: 连接数据库失败: %v %s\n", redColor, err, resetColor)
			return
		}

		fmt.Printf("%s 正在提取参考签名（shop_01.sql 执行后）... %s\n", blueColor, resetColor)
		referenceSig, err := extractDatabaseStructure(companyDB, dbName)
		if err != nil {
			fmt.Printf("%s 错误: 提取参考签名失败: %v %s\n", redColor, err, resetColor)
			return
		}
		fmt.Printf("%s 参考签名提取完成: %s\n", greenColor, resetColor)
		fmt.Printf("  表数量: %d\n", referenceSig.TableCount)
		fmt.Printf("  字段总数: %d\n", referenceSig.TotalFields)
		fmt.Printf("  结构签名: %s...\n", referenceSig.Signature[:32])

		// 第二步：执行迁移文件，读取当前签名
		fmt.Printf("\n%s 第二步：执行迁移文件并提取当前签名... %s\n", blueColor, resetColor)
		if err := runMigrations(dbName); err != nil {
			fmt.Printf("%s 错误: 执行迁移文件失败: %v %s\n", redColor, err, resetColor)
			return
		}

		fmt.Printf("%s 正在提取当前签名（迁移文件执行后）... %s\n", blueColor, resetColor)
		currentSig, err := extractDatabaseStructure(companyDB, dbName)
		if err != nil {
			fmt.Printf("%s 错误: 提取当前签名失败: %v %s\n", redColor, err, resetColor)
			return
		}
		fmt.Printf("%s 当前签名提取完成: %s\n", greenColor, resetColor)
		fmt.Printf("  表数量: %d\n", currentSig.TableCount)
		fmt.Printf("  字段总数: %d\n", currentSig.TotalFields)
		fmt.Printf("  结构签名: %s...\n", currentSig.Signature[:32])

		// 第三步：对比签名
		fmt.Printf("\n%s 第三步：对比签名... %s\n", blueColor, resetColor)
		compareSignatures(referenceSig, currentSig, testCompanyUuid)
	},
}

// extractDatabaseStructure 从数据库提取结构
func extractDatabaseStructure(db *gorm.DB, dbName string) (*DatabaseSignature, error) {
	// 获取所有表（使用 information_schema 更可靠）
	var tableResults []struct {
		TableName string `gorm:"column:TABLE_NAME"`
	}

	query := fmt.Sprintf(`
		SELECT TABLE_NAME 
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = '%s' 
		AND TABLE_TYPE = 'BASE TABLE'
		AND TABLE_NAME NOT IN ('migrations', 'ttpos_migrations')
		ORDER BY TABLE_NAME
	`, dbName)

	if err := db.Raw(query).Scan(&tableResults).Error; err != nil {
		return nil, fmt.Errorf("获取表列表失败: %w", err)
	}

	tables := make([]string, 0, len(tableResults))
	for _, t := range tableResults {
		// 忽略 migrations 表（迁移系统表）
		if t.TableName == "migrations" || t.TableName == "ttpos_migrations" {
			continue
		}
		tables = append(tables, t.TableName)
	}

	// 移除表名前缀（如果有）
	cleanTables := make(map[string]string) // 原始表名 -> 清理后的表名
	for _, table := range tables {
		cleanTable := strings.TrimPrefix(table, "ttpos_")
		// 再次检查清理后的表名，确保过滤掉 migrations
		if cleanTable == "migrations" {
			continue
		}
		cleanTables[cleanTable] = table
	}

	// 提取每个表的结构
	structure := make([]TableStructure, 0, len(cleanTables))
	for cleanTable, originalTable := range cleanTables {
		fields, err := extractTableFields(db, originalTable)
		if err != nil {
			return nil, fmt.Errorf("提取表 %s 的字段失败: %w", originalTable, err)
		}

		// 生成字段签名列表（字段名:类型:默认值:注释）
		fieldSignatures := make([]string, 0, len(fields))
		for _, field := range fields {
			// 处理默认值
			defaultValue := "NULL"
			if field.Default.Valid {
				defaultValue = field.Default.String
			}

			// 转义注释中的特殊字符，避免影响签名解析
			comment := strings.ReplaceAll(field.Comment, ":", "&#58;")
			comment = strings.ReplaceAll(comment, "\n", " ")

			// 格式：字段名:类型:默认值:注释
			fieldSignatures = append(fieldSignatures, fmt.Sprintf("%s:%s:%s:%s",
				field.Name, field.Type, defaultValue, comment))
		}
		sort.Strings(fieldSignatures)

		structure = append(structure, TableStructure{
			Table:      cleanTable,
			Fields:     fieldSignatures,
			FieldCount: len(fieldSignatures),
		})
	}

	// 按表名排序
	sort.Slice(structure, func(i, j int) bool {
		return structure[i].Table < structure[j].Table
	})

	// 生成签名
	signatureJSON, err := json.Marshal(structure)
	if err != nil {
		return nil, fmt.Errorf("生成签名JSON失败: %w", err)
	}

	hash := sha256.Sum256(signatureJSON)
	signatureStr := fmt.Sprintf("%x", hash)

	totalFields := 0
	for _, s := range structure {
		totalFields += s.FieldCount
	}

	return &DatabaseSignature{
		Signature:   signatureStr,
		TableCount:  len(structure),
		TotalFields: totalFields,
		Structure:   structure,
		Database:    dbName,
	}, nil
}

// extractTableFields 提取表的字段信息
func extractTableFields(db *gorm.DB, tableName string) ([]FieldInfo, error) {
	var results []struct {
		Field   string         `gorm:"column:COLUMN_NAME"`
		Type    string         `gorm:"column:COLUMN_TYPE"`
		Default sql.NullString `gorm:"column:COLUMN_DEFAULT"`
		Comment string         `gorm:"column:COLUMN_COMMENT"`
	}

	// 使用 information_schema 获取完整的字段信息，包括注释
	query := fmt.Sprintf(`
		SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_DEFAULT, COLUMN_COMMENT
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`)

	if err := db.Raw(query, tableName).Scan(&results).Error; err != nil {
		return nil, err
	}

	fields := make([]FieldInfo, 0, len(results))
	for _, r := range results {
		fields = append(fields, FieldInfo{
			Name:    r.Field,
			Type:    r.Type, // 保留完整类型（包括长度）
			Default: r.Default,
			Comment: r.Comment,
		})
	}

	return fields, nil
}

// simplifyFieldType 简化字段类型
func simplifyFieldType(fullType string) string {
	// 移除括号及其内容，只保留基本类型
	// 例如: VARCHAR(255) -> VARCHAR, DECIMAL(22,4) -> DECIMAL
	parts := strings.Split(fullType, "(")
	if len(parts) > 0 {
		return strings.ToUpper(strings.TrimSpace(parts[0]))
	}
	return strings.ToUpper(strings.TrimSpace(fullType))
}

// loadOrGenerateReferenceSignature 加载或生成参考签名
func loadOrGenerateReferenceSignature(filePath string) (*DatabaseSignature, error) {
	// 先尝试加载现有文件
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("  读取现有参考签名文件...\n")
		data, err := os.ReadFile(filePath)
		if err == nil {
			var sig DatabaseSignature
			if err := json.Unmarshal(data, &sig); err == nil {
				fmt.Printf("%s  ✓ 已加载参考签名 %s\n", greenColor, resetColor)
				return &sig, nil
			}
		}
	}

	// 文件不存在或读取失败，自动生成
	fmt.Printf("  参考签名文件不存在，正在自动生成...\n")

	workDir, err := getProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("获取项目根目录失败: %w", err)
	}
	sqlFile := filepath.Join(workDir, "admin", "database", "seeds", "shop_01.sql")
	migrationsDir := filepath.Join(workDir, "admin", "database", "migrations")

	// 从 SQL 和迁移文件生成参考签名
	extractor := NewStructureExtractor()

	fmt.Printf("    从 shop_01.sql 提取基础结构...\n")
	if err := extractor.ExtractFromSQL(sqlFile); err != nil {
		return nil, fmt.Errorf("从 SQL 文件提取结构失败: %w", err)
	}
	fmt.Printf("    ✓ 提取了 %d 个表\n", len(extractor.tables))

	fmt.Printf("    从迁移文件提取结构变更...\n")
	migrationCount, err := extractor.ExtractFromMigrations(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("从迁移文件提取结构失败: %w", err)
	}
	fmt.Printf("    ✓ 处理了 %d 个迁移文件\n", migrationCount)

	// 生成签名
	sig := extractor.GenerateSignature()

	// 保存签名文件
	if err := saveSignatureToFile(sig, filePath); err != nil {
		return nil, fmt.Errorf("保存签名文件失败: %w", err)
	}

	fmt.Printf("%s  ✓ 参考签名已生成并保存 %s\n", greenColor, resetColor)
	return sig, nil
}

// StructureExtractor 数据库结构提取器
type StructureExtractor struct {
	tables map[string]*TableInfo
}

// TableInfo 表信息
type TableInfo struct {
	Fields []FieldInfo
	Source string
}

// NewStructureExtractor 创建结构提取器
func NewStructureExtractor() *StructureExtractor {
	return &StructureExtractor{
		tables: make(map[string]*TableInfo),
	}
}

// ExtractFromSQL 从 SQL 文件提取表结构
func (e *StructureExtractor) ExtractFromSQL(sqlFile string) error {
	file, err := os.Open(sqlFile)
	if err != nil {
		return fmt.Errorf("打开 SQL 文件失败: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("读取 SQL 文件失败: %w", err)
	}

	// 使用正则表达式提取 CREATE TABLE 语句
	tablePattern := regexp.MustCompile(`(?i)CREATE TABLE (?:IF NOT EXISTS )?` + "`([^`]+)`" + `\s*\((.*?)\)\s*ENGINE`)
	matches := tablePattern.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		tableName := match[1]
		tableBody := match[2]

		// 移除 ttpos_ 前缀
		cleanTableName := strings.TrimPrefix(tableName, "ttpos_")

		// 提取字段
		fields := e.extractFieldsFromTableBody(tableBody)

		e.tables[cleanTableName] = &TableInfo{
			Fields: fields,
			Source: "shop_01.sql",
		}
	}

	return nil
}

// extractFieldsFromTableBody 从表体中提取字段定义
func (e *StructureExtractor) extractFieldsFromTableBody(tableBody string) []FieldInfo {
	var fields []FieldInfo

	// 使用正则表达式直接匹配字段定义
	fieldPattern := regexp.MustCompile("`([^`]+)`\\s+(\\w+(?:\\([^)]+\\))?)")
	matches := fieldPattern.FindAllStringSubmatch(tableBody, -1)

	fieldMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) >= 3 {
			fieldName := match[1]
			fieldType := match[2]
			fieldType = simplifyFieldType(fieldType)

			// 避免重复
			if !fieldMap[fieldName] {
				fields = append(fields, FieldInfo{
					Name: fieldName,
					Type: fieldType,
				})
				fieldMap[fieldName] = true
			}
		}
	}

	return fields
}

// ExtractFromMigrations 从迁移文件提取结构变更
func (e *StructureExtractor) ExtractFromMigrations(migrationsDir string) (int, error) {
	// 获取所有大于 20250722091421 的迁移文件
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.php"))
	if err != nil {
		return 0, err
	}

	var migrationFiles []string
	for _, file := range files {
		fileName := filepath.Base(file)
		// 检查文件名时间戳
		if len(fileName) >= 14 {
			timestamp := fileName[:14]
			if timestamp > "20250722091421" || strings.HasPrefix(timestamp, "202508") ||
				strings.HasPrefix(timestamp, "202509") || strings.HasPrefix(timestamp, "202510") ||
				strings.HasPrefix(timestamp, "202511") {
				migrationFiles = append(migrationFiles, file)
			}
		}
	}

	sort.Strings(migrationFiles)

	count := 0
	for _, file := range migrationFiles {
		if err := e.extractFromMigrationFile(file); err != nil {
			// 只记录警告，不中断
			continue
		}
		count++
	}

	return len(migrationFiles), nil
}

// extractFromMigrationFile 从单个迁移文件提取结构变更
func (e *StructureExtractor) extractFromMigrationFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// 提取表名
	tablePattern := regexp.MustCompile(`->table\(['"]([^'"]+)['"]`)
	tableMatches := tablePattern.FindAllStringSubmatch(contentStr, -1)

	// 提取 addColumn 的字段
	addColumnPattern := regexp.MustCompile(`->addColumn\(['"]([^'"]+)['"],\s*['"]([^'"]+)['"]`)
	columnMatches := addColumnPattern.FindAllStringSubmatch(contentStr, -1)

	// 提取 createTable 的表
	createTablePattern := regexp.MustCompile(`(?:createTable|hasTable)\(['"]([^'"]+)['"]`)
	createTableMatches := createTablePattern.FindAllStringSubmatch(contentStr, -1)

	// 合并所有表名
	tables := make(map[string]bool)
	for _, match := range tableMatches {
		if len(match) > 1 {
			tables[match[1]] = true
		}
	}
	for _, match := range createTableMatches {
		if len(match) > 1 {
			tables[match[1]] = true
		}
	}

	// 处理每个表
	for tableName := range tables {
		cleanTableName := strings.TrimPrefix(tableName, "ttpos_")

		// 如果表不存在，创建它
		if _, exists := e.tables[cleanTableName]; !exists {
			e.tables[cleanTableName] = &TableInfo{
				Fields: []FieldInfo{},
				Source: fmt.Sprintf("migration: %s", filepath.Base(filePath)),
			}
		}

		// 添加字段
		fieldMap := make(map[string]bool)
		for _, f := range e.tables[cleanTableName].Fields {
			fieldMap[f.Name] = true
		}

		for _, match := range columnMatches {
			if len(match) >= 3 {
				fieldName := match[1]
				fieldType := match[2]
				fieldType = simplifyFieldType(fieldType)

				// 检查字段是否已存在
				if !fieldMap[fieldName] {
					e.tables[cleanTableName].Fields = append(e.tables[cleanTableName].Fields, FieldInfo{
						Name: fieldName,
						Type: fieldType,
					})
					fieldMap[fieldName] = true
				}
			}
		}
	}

	return nil
}

// GenerateSignature 生成数据库结构签名
func (e *StructureExtractor) GenerateSignature() *DatabaseSignature {
	// 转换为 TableStructure 列表
	structure := make([]TableStructure, 0, len(e.tables))
	for tableName, tableInfo := range e.tables {
		// 按字段名排序
		sortedFields := make([]string, 0, len(tableInfo.Fields))
		for _, field := range tableInfo.Fields {
			sortedFields = append(sortedFields, fmt.Sprintf("%s:%s", field.Name, field.Type))
		}
		sort.Strings(sortedFields)

		structure = append(structure, TableStructure{
			Table:      tableName,
			Fields:     sortedFields,
			FieldCount: len(sortedFields),
		})
	}

	// 按表名排序
	sort.Slice(structure, func(i, j int) bool {
		return structure[i].Table < structure[j].Table
	})

	// 生成签名
	signatureJSON, err := json.Marshal(structure)
	if err != nil {
		return nil
	}

	hash := sha256.Sum256(signatureJSON)
	signatureStr := fmt.Sprintf("%x", hash)

	totalFields := 0
	for _, s := range structure {
		totalFields += s.FieldCount
	}

	return &DatabaseSignature{
		Signature:   signatureStr,
		TableCount:  len(structure),
		TotalFields: totalFields,
		Structure:   structure,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}
}

// saveSignatureToFile 保存签名到文件
func saveSignatureToFile(sig *DatabaseSignature, filePath string) error {
	data, err := json.MarshalIndent(sig, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化签名失败: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// saveSignature 保存签名到文件
func saveSignature(sig *DatabaseSignature, filename string) {
	data, err := json.MarshalIndent(sig, "", "  ")
	if err != nil {
		fmt.Printf("%s 错误: 序列化签名失败: %v %s\n", redColor, err, resetColor)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("%s 错误: 保存签名文件失败: %v %s\n", redColor, err, resetColor)
		return
	}

	fmt.Printf("%s 签名已保存到: %s %s\n", greenColor, filename, resetColor)
}

// compareSignatures 对比两个签名
// reference: shop_01.sql 执行后的数据库结构
// current: 迁移文件执行后的数据库结构
func compareSignatures(reference, current *DatabaseSignature, _ uint64) {
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println("数据库结构签名对比")
	fmt.Println("参考签名: shop_01.sql 执行后的数据库结构")
	fmt.Println("当前签名: 迁移文件执行后的数据库结构")
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println()

	if reference.Signature == current.Signature {
		fmt.Printf("%s✅ 签名完全一致！数据库结构相同。%s\n", greenColor, resetColor)
		fmt.Printf("\n表数量: %d\n", current.TableCount)
		fmt.Printf("字段总数: %d\n", current.TotalFields)
		return
	}

	fmt.Printf("%s❌ 签名不一致！数据库结构存在差异。%s\n", redColor, resetColor)
	fmt.Printf("\n参考签名 (shop_01.sql): %s...\n", reference.Signature[:32])
	fmt.Printf("当前签名 (迁移后): %s...\n", current.Signature[:32])

	// 详细对比
	fmt.Println("\n详细对比:")
	fmt.Println("-" + strings.Repeat("-", 79))

	// 对比表数量
	if reference.TableCount != current.TableCount {
		diff := current.TableCount - reference.TableCount
		if diff > 0 {
			fmt.Printf("表总数不一致: %d (shop_01.sql) vs %d (迁移后) - 迁移文件新增了 %d 个表\n",
				reference.TableCount, current.TableCount, diff)
		} else {
			fmt.Printf("表总数不一致: %d (shop_01.sql) vs %d (迁移后) - 迁移文件减少了 %d 个表\n",
				reference.TableCount, current.TableCount, -diff)
		}
	} else {
		fmt.Printf("表总数一致: %d 个表\n", reference.TableCount)
	}

	// 对比字段总数
	if reference.TotalFields != current.TotalFields {
		diff := current.TotalFields - reference.TotalFields
		if diff > 0 {
			fmt.Printf("字段总数不一致: %d (shop_01.sql) vs %d (迁移后) - 迁移文件新增了 %d 个字段\n",
				reference.TotalFields, current.TotalFields, diff)
		} else {
			fmt.Printf("字段总数不一致: %d (shop_01.sql) vs %d (迁移后) - 迁移文件减少了 %d 个字段\n",
				reference.TotalFields, current.TotalFields, -diff)
		}
	} else {
		fmt.Printf("字段总数一致: %d 个字段\n", reference.TotalFields)
	}

	// 构建表映射
	refTables := make(map[string]TableStructure)
	for _, t := range reference.Structure {
		refTables[t.Table] = t
	}

	curTables := make(map[string]TableStructure)
	for _, t := range current.Structure {
		curTables[t.Table] = t
	}

	// 找出差异
	var onlyInRef, onlyInCur []string
	var commonTables []string

	for table := range refTables {
		if _, exists := curTables[table]; !exists {
			onlyInRef = append(onlyInRef, table)
		} else {
			commonTables = append(commonTables, table)
		}
	}

	for table := range curTables {
		if _, exists := refTables[table]; !exists {
			onlyInCur = append(onlyInCur, table)
		}
	}

	if len(onlyInRef) > 0 {
		fmt.Printf("\n只在 shop_01.sql 执行后存在的表 (%d 个):\n", len(onlyInRef))
		sort.Strings(onlyInRef)
		for _, table := range onlyInRef {
			fmt.Printf("  - %s\n", table)
		}
	}

	if len(onlyInCur) > 0 {
		fmt.Printf("\n只在迁移文件执行后存在的表 (%d 个):\n", len(onlyInCur))
		sort.Strings(onlyInCur)
		for _, table := range onlyInCur {
			fmt.Printf("  - %s\n", table)
		}
	}

	// 对比共同表的字段（包括类型、默认值、注释检查）
	type FieldDiff struct {
		MissingInSQL    []string // shop_01.sql 执行后缺失的字段（迁移文件执行后新增的）
		ExtraInDB       []string // 当前数据库中多出的字段
		TypeMismatch    []string // 字段类型不一致（格式：字段名 (参考类型 -> 当前类型)）
		DefaultMismatch []string // 默认值不一致
		CommentMismatch []string // 注释不一致
	}

	fieldDiffs := make(map[string]*FieldDiff)

	for _, table := range commonTables {
		// 构建字段映射：字段名 -> 完整签名（类型:默认值:注释）
		refFieldMap := make(map[string]string) // 字段名 -> 完整信息
		for _, f := range refTables[table].Fields {
			// 解析字段签名：字段名:类型:默认值:注释
			parts := strings.SplitN(f, ":", 2)
			if len(parts) >= 2 {
				refFieldMap[parts[0]] = parts[1] // 保存类型:默认值:注释
			}
		}

		curFieldMap := make(map[string]string)
		for _, f := range curTables[table].Fields {
			parts := strings.SplitN(f, ":", 2)
			if len(parts) >= 2 {
				curFieldMap[parts[0]] = parts[1]
			}
		}

		diff := &FieldDiff{}

		// 检查参考签名中的字段（迁移文件中的字段）
		for fieldName, refFullInfo := range refFieldMap {
			curFullInfo, exists := curFieldMap[fieldName]
			if !exists {
				// 字段在参考签名中存在，但在当前数据库中不存在
				refParts := strings.SplitN(refFullInfo, ":", 3)
				refType := refParts[0]
				diff.MissingInSQL = append(diff.MissingInSQL, fmt.Sprintf("%s (%s)", fieldName, refType))
			} else if refFullInfo != curFullInfo {
				// 字段存在但信息不一致，详细对比
				refParts := strings.SplitN(refFullInfo, ":", 3)
				curParts := strings.SplitN(curFullInfo, ":", 3)

				if len(refParts) >= 1 && len(curParts) >= 1 {
					refType := refParts[0]
					curType := curParts[0]
					if refType != curType {
						diff.TypeMismatch = append(diff.TypeMismatch,
							fmt.Sprintf("%s (%s -> %s)", fieldName, refType, curType))
					}
				}

				if len(refParts) >= 2 && len(curParts) >= 2 {
					refDefault := refParts[1]
					curDefault := curParts[1]
					if refDefault != curDefault {
						diff.DefaultMismatch = append(diff.DefaultMismatch,
							fmt.Sprintf("%s (%s -> %s)", fieldName, refDefault, curDefault))
					}
				}

				if len(refParts) >= 3 && len(curParts) >= 3 {
					refComment := refParts[2]
					curComment := curParts[2]
					if refComment != curComment {
						diff.CommentMismatch = append(diff.CommentMismatch,
							fmt.Sprintf("%s (%s -> %s)", fieldName, refComment, curComment))
					}
				}
			}
		}

		// 检查当前数据库中的额外字段
		for fieldName := range curFieldMap {
			if _, exists := refFieldMap[fieldName]; !exists {
				curParts := strings.SplitN(curFieldMap[fieldName], ":", 3)
				curType := curParts[0]
				diff.ExtraInDB = append(diff.ExtraInDB, fmt.Sprintf("%s (%s)", fieldName, curType))
			}
		}

		// 如果有差异，记录
		if len(diff.MissingInSQL) > 0 || len(diff.ExtraInDB) > 0 || len(diff.TypeMismatch) > 0 ||
			len(diff.DefaultMismatch) > 0 || len(diff.CommentMismatch) > 0 {
			fieldDiffs[table] = diff
		}
	}

	if len(fieldDiffs) > 0 {
		fmt.Printf("\n字段差异详情 (%d 个表):\n", len(fieldDiffs))
		fmt.Println("-" + strings.Repeat("-", 79))

		count := 0
		for table, diff := range fieldDiffs {
			if count >= 20 {
				fmt.Printf("\n  ... 还有 %d 个表存在字段差异\n", len(fieldDiffs)-20)
				break
			}

			hasAnyDiff := false
			fmt.Printf("\n表: %s\n", table)

			// 显示 shop_01.sql 执行后缺失的字段（迁移文件执行后新增的）
			if len(diff.MissingInSQL) > 0 {
				hasAnyDiff = true
				sort.Strings(diff.MissingInSQL)
				fmt.Printf("  ⚠️  迁移文件新增的字段（shop_01.sql 执行后不存在）:\n")
				for _, field := range diff.MissingInSQL {
					fmt.Printf("     - %s\n", field)
				}
			}

			// 显示字段类型不一致
			if len(diff.TypeMismatch) > 0 {
				hasAnyDiff = true
				sort.Strings(diff.TypeMismatch)
				fmt.Printf("  ⚠️  字段类型不一致:\n")
				for _, field := range diff.TypeMismatch {
					fmt.Printf("     - %s\n", field)
				}
			}

			// 显示默认值不一致
			if len(diff.DefaultMismatch) > 0 {
				hasAnyDiff = true
				sort.Strings(diff.DefaultMismatch)
				fmt.Printf("  ⚠️  默认值不一致:\n")
				for _, field := range diff.DefaultMismatch {
					fmt.Printf("     - %s\n", field)
				}
			}

			// 显示注释不一致
			if len(diff.CommentMismatch) > 0 {
				hasAnyDiff = true
				sort.Strings(diff.CommentMismatch)
				fmt.Printf("  ⚠️  注释不一致:\n")
				for _, field := range diff.CommentMismatch {
					fmt.Printf("     - %s\n", field)
				}
			}

			// 显示当前数据库中多出的字段
			if len(diff.ExtraInDB) > 0 {
				hasAnyDiff = true
				sort.Strings(diff.ExtraInDB)
				fmt.Printf("  ℹ️  当前数据库中多出的字段:\n")
				for _, field := range diff.ExtraInDB {
					fmt.Printf("     - %s\n", field)
				}
			}

			if !hasAnyDiff {
				fmt.Printf("  (无差异)\n")
			}

			count++
		}
	}

	// 保存当前签名供后续对比
	// saveSignature(current, fmt.Sprintf("database_structure_shop%d.json", companyUuid))
}

// createTestCompany 在 saas 库中创建测试商户
func createTestCompany() (uint64, error) {
	fmt.Printf("%s 正在创建测试商户... %s\n", blueColor, resetColor)

	// 连接 saas 数据库
	saasDB, err := database.NewMySQLConnection(config.Database, "saas")
	if err != nil {
		return 0, fmt.Errorf("连接 saas 数据库失败: %w", err)
	}

	// 生成测试商户 UUID
	testUuid, err := utils.GetID()
	if err != nil {
		return 0, fmt.Errorf("生成商户UUID失败: %w", err)
	}

	// 创建测试商户记录
	company := &model.Company{
		BaseModel: model.BaseModel{
			Uuid:       testUuid,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:          "测试商户_验证数据库结构",
		Logo:          "",
		ExpireTime:    0, // 永不过期
		AuthDay:       0,
		Status:        1, // 启用
		AuthStartTime: time.Now().Unix(),
		OldCompanyId:  0,
		IsEnableErp:   0,
		LastSyncTime:  0,
	}

	if err := saasDB.Create(company).Error; err != nil {
		return 0, fmt.Errorf("创建商户记录失败: %w", err)
	}

	fmt.Printf("%s  ✓ 测试商户创建成功，UUID: %d %s\n", greenColor, testUuid, resetColor)
	return testUuid, nil
}

// createTestDatabase 创建测试数据库
func createTestDatabase(dbName string) error {
	fmt.Printf("%s 正在创建测试数据库: %s... %s\n", blueColor, dbName, resetColor)

	// 创建数据库连接（不指定数据库名）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	// 创建数据库
	fmt.Printf("  创建数据库...\n")
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", dbName))
	if err != nil {
		return fmt.Errorf("创建数据库失败: %w", err)
	}
	fmt.Printf("%s  ✓ 数据库创建成功 %s\n", greenColor, resetColor)

	return nil
}

// executeShop01SQL 执行 shop_01.sql
func executeShop01SQL(dbName string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	workDir, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("获取项目根目录失败: %w", err)
	}

	// 执行 shop_01.sql
	sqlFile := filepath.Join(workDir, "admin", "database", "seeds", "shop_01.sql")
	fmt.Printf("  执行 shop_01.sql...\n")
	if err := executeSQLFile(db, sqlFile); err != nil {
		return fmt.Errorf("执行 shop_01.sql 失败: %w", err)
	}
	fmt.Printf("%s  ✓ shop_01.sql 执行成功 %s\n", greenColor, resetColor)

	// 执行 shop_init_data.sql（如果存在）
	initDataFile := filepath.Join(workDir, "admin", "database", "seeds", "shop_init_data.sql")
	if _, err := os.Stat(initDataFile); err == nil {
		fmt.Printf("  执行 shop_init_data.sql...\n")
		if err := executeSQLFile(db, initDataFile); err != nil {
			return fmt.Errorf("执行 shop_init_data.sql 失败: %w", err)
		}
		fmt.Printf("%s  ✓ shop_init_data.sql 执行成功 %s\n", greenColor, resetColor)
	}

	return nil
}

// executeSQLFile 执行 SQL 文件
func executeSQLFile(db *sql.DB, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取 SQL 文件失败: %w", err)
	}

	// 分割 SQL 语句（按分号分割，但要考虑字符串中的分号）
	sqlStatements := splitSQLStatements(string(content))

	for _, stmt := range sqlStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		_, err := db.Exec(stmt)
		if err != nil {
			// 忽略一些常见错误（如表已存在等）
			if !strings.Contains(err.Error(), "already exists") &&
				!strings.Contains(err.Error(), "Duplicate") {
				return fmt.Errorf("执行 SQL 失败: %w\nSQL: %s", err, stmt[:min(100, len(stmt))])
			}
		}
	}

	return nil
}

// splitSQLStatements 分割 SQL 语句
// 正确处理字符串、注释和反引号
func splitSQLStatements(content string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	inBacktick := false
	escapeNext := false
	inSingleLineComment := false
	inMultiLineComment := false

	runes := []rune(content)
	for i := 0; i < len(runes); i++ {
		char := runes[i]

		// 处理转义字符
		if escapeNext {
			if !inSingleLineComment && !inMultiLineComment {
				current.WriteRune(char)
			}
			escapeNext = false
			continue
		}

		// 检查转义字符
		if char == '\\' && (inString || inBacktick) {
			escapeNext = true
			if !inSingleLineComment && !inMultiLineComment {
				current.WriteRune(char)
			}
			continue
		}

		// 处理多行注释 /* */
		if !inString && !inBacktick && !inSingleLineComment {
			if char == '/' && i+1 < len(runes) && runes[i+1] == '*' {
				inMultiLineComment = true
				i++ // 跳过 '*'
				continue
			}
			if inMultiLineComment {
				if char == '*' && i+1 < len(runes) && runes[i+1] == '/' {
					inMultiLineComment = false
					i++ // 跳过 '/'
					continue
				}
				continue // 忽略注释内容
			}
		}

		// 处理单行注释 --
		if !inString && !inBacktick && !inMultiLineComment {
			if char == '-' && i+1 < len(runes) && runes[i+1] == '-' {
				// 检查后面是否是空格、制表符或换行符（MySQL 注释要求）
				if i+2 >= len(runes) || runes[i+2] == ' ' || runes[i+2] == '\t' || runes[i+2] == '\n' || runes[i+2] == '\r' {
					// 移除刚才写入的 '-'
					currentStr := current.String()
					if len(currentStr) > 0 {
						runesInCurrent := []rune(currentStr)
						if len(runesInCurrent) > 0 && runesInCurrent[len(runesInCurrent)-1] == '-' {
							current.Reset()
							current.WriteString(string(runesInCurrent[:len(runesInCurrent)-1]))
						}
					}
					inSingleLineComment = true
					i++ // 跳过第二个 '-'
					continue
				}
			}
			if inSingleLineComment {
				if char == '\n' || char == '\r' {
					inSingleLineComment = false
				}
				continue // 忽略注释内容
			}
		}

		// 处理字符串（单引号或双引号）
		if !inBacktick && !inSingleLineComment && !inMultiLineComment {
			if char == '\'' || char == '"' {
				inString = !inString
			}
		}

		// 处理反引号（MySQL 标识符）
		if !inString && !inSingleLineComment && !inMultiLineComment {
			if char == '`' {
				inBacktick = !inBacktick
			}
		}

		// 写入字符（不在注释中）
		if !inSingleLineComment && !inMultiLineComment {
			current.WriteRune(char)
		}

		// 只有在非字符串、非注释、非反引号内时才检查分号
		if !inString && !inBacktick && !inSingleLineComment && !inMultiLineComment {
			if char == ';' {
				stmt := strings.TrimSpace(current.String())
				if stmt != "" {
					statements = append(statements, stmt)
				}
				current.Reset()
			}
		}
	}

	// 添加最后一个语句（如果没有分号结尾）
	if current.Len() > 0 {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}

// runMigrations 执行迁移文件
func runMigrations(_ string) error {
	workDir, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("获取项目根目录失败: %w", err)
	}
	adminDir := filepath.Join(workDir, "admin")

	// 检查 admin 目录是否存在
	if _, err := os.Stat(adminDir); os.IsNotExist(err) {
		return fmt.Errorf("admin 目录不存在: %s", adminDir)
	}

	// 检查 think 文件是否存在
	thinkFile := filepath.Join(adminDir, "think")
	if _, err := os.Stat(thinkFile); os.IsNotExist(err) {
		return fmt.Errorf("think 文件不存在: %s", thinkFile)
	}

	// 执行迁移命令: make php-migrate (需要在根目录执行，因为 php-migrate 定义在根目录的 Makefile 中)
	fmt.Printf("  开始执行迁移文件...\n")
	startTime := time.Now()

	cmd := exec.Command("make", "php-migrate")
	cmd.Dir = workDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	elapsed := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("迁移命令执行失败: %w\n输出: %s", err, string(output))
	}

	fmt.Printf("%s  ✓ 迁移文件执行成功，耗时: %s %s\n", greenColor, elapsed.Round(time.Millisecond), resetColor)

	return nil
}

// cleanupTestCompany 清理测试商户和数据库
func cleanupTestCompany(companyUuid uint64) {
	fmt.Printf("\n%s 正在清理测试数据... %s\n", blueColor, resetColor)

	// 删除数据库
	dbName := fmt.Sprintf("%s%d", constant.DBNamePrefix, companyUuid)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port)

	db, err := sql.Open("mysql", dsn)
	if err == nil {
		_, _ = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
		db.Close()
		fmt.Printf("%s  ✓ 测试数据库已删除 %s\n", greenColor, resetColor)
	}

	// 删除 saas 库中的商户记录
	saasDB, err := database.NewMySQLConnection(config.Database, "saas")
	if err == nil {
		_ = saasDB.Where("uuid = ?", companyUuid).Delete(&model.Company{})
		fmt.Printf("%s  ✓ 测试商户记录已删除 %s\n", greenColor, resetColor)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
