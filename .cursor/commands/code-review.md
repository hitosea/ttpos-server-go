---
name: code-review
description: 代码审查 - 事务错误处理和安全审计检查
---

# /code-review - 代码审查命令

## 使用场景

对指定文件或目录进行代码审查，重点检查事务错误处理和安全漏洞相关问题。

> **注意**: 此命令包括事务错误处理检查和安全审计检查，涵盖 SQL 注入、命令注入、认证授权、敏感信息泄露、XSS、路径遍历、加密安全等方面。

## 使用方式

```bash
/code-review order_manage.go
/code-review main/app/service
/code-review main/app/service --focus transaction
/code-review main/app/service --focus error-handling
/code-review main/app/service --focus security
/code-review main/app/service --focus all
```

## 参数

- `target`: 必填，目标文件或目录
  - 文件路径: `order_manage.go`、`main/app/service/order_manage.go`
  - 目录路径: `main/app/service`、`main/app/service/order`
- `--focus`: 可选，检查重点
  - `transaction`: 事务相关检查（默认）
  - `error-handling`: 错误处理检查
  - `security`: 安全审计检查
  - `all`: 全面检查（事务 + 错误处理 + 安全审计）
- `--fix`: 可选，自动修复发现的问题（默认: false，仅报告）
- `--report`: 可选，生成详细报告文件路径（默认: 不生成）

## 功能特点

### 事务和错误处理
- ✅ **事务使用检查**：检查事务中使用外部 db 而非 tx
- ✅ **错误处理检查**：检查未处理的错误
- ✅ **上下文传递检查**：检查事务上下文传递
- ✅ **ctx.GetDB() 检查**：检查事务中使用 ctx.GetDB() 但未设置事务上下文
- ✅ **Goroutine 检查**：检查 goroutine 中的错误处理
- ✅ **JSON 序列化检查**：检查 JSON 序列化/反序列化错误处理

### 安全审计
- ✅ **SQL 注入检查**：检查 JSON_EXTRACT 路径注入、fmt.Sprintf SQL 拼接、CREATE/DROP DATABASE 注入
- ✅ **命令注入检查**：检查 Shell 命令执行漏洞
- ✅ **认证授权检查**：检查 JWT 签名算法验证、Token 过期时间、MD5 Token 验证
- ✅ **敏感信息泄露检查**：检查硬编码凭据、API 响应返回密码、日志泄露密码
- ✅ **XSS 和输入验证检查**：检查 CORS 配置、HTTP 安全响应头、输入参数验证
- ✅ **文件上传和路径遍历检查**：检查 Delete 函数路径遍历、文件写入路径遍历、文件扩展名验证
- ✅ **加密和密码安全检查**：检查 MD5 密码加密、硬编码 AES 密钥

### 通用功能
- ✅ **优先级分类**：按严重程度分类问题（严重/高/中/低）
- ✅ **修复建议**：提供具体的修复建议和代码示例

## 检查项

### 1. 事务中使用外部 db 而非 tx（严重）

检查事务回调函数中是否使用外部 `db` 而非事务 `tx`。

**检查模式**：
```go
// 搜索模式
repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    // 检查内部是否使用 db 而非 tx
})
```

### 2. 事务调用未检查返回错误（高）

检查事务调用是否检查返回错误。

**检查模式**：
```go
// 搜索模式
db.Transaction(...)  // 未检查错误
repository.CommonRepo.Transaction(...)  // 未检查错误
```

### 3. 调用方法未检查返回错误（高）

检查方法调用是否检查返回错误。

**检查模式**：
```go
// 搜索模式
s.xxxSrv.Method(...)  // 未检查错误
repository.NewXxxRepo(...).Method(...)  // 未检查错误
```

### 4. 调用服务方法前未设置事务上下文（中）

检查事务中调用服务方法前是否设置事务上下文。

**检查模式**：
```go
// 搜索模式
repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    s.xxxSrv.Method(ctx, ...)  // 检查是否设置了 ctx.SetDB(tx)
})
```

### 5. Goroutine 中调用方法未检查返回错误（中）

检查 goroutine 中的方法调用是否检查错误。

**检查模式**：
```go
// 搜索模式
utils.Go(func() {
    // 检查内部方法调用是否检查错误
})
```

### 6. JSON 序列化/反序列化未检查错误（低）

检查 JSON 序列化/反序列化是否检查错误。

**检查模式**：
```go
// 搜索模式
json.Marshal(...)  // 未检查错误
json.Unmarshal(...)  // 未检查错误
```

### 7. 事务中使用 ctx.GetDB() 但未设置事务上下文（严重）

检查事务回调函数中使用 `ctx.GetDB()` 获取数据库连接，但未先调用 `ctx.SetDB(tx)` 设置事务上下文的情况。

**问题描述**：
在事务回调函数中，如果使用 `ctx.GetDB()` 获取数据库连接，但没有先调用 `ctx.SetDB(tx)` 设置事务上下文，那么获取到的仍然是普通的 `db` 连接而不是事务 `tx`，导致操作不在事务中执行。

**检查模式**：
```go
// 搜索模式 1: 事务回调函数中直接使用 ctx.GetDB()
repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    db := ctx.GetDB()  // ❌ 错误：获取的是普通 db，不是 tx
    // ... 使用 db 进行操作 ...
})

// 搜索模式 2: 事务回调函数中调用服务方法，服务方法内部使用 ctx.GetDB()
repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    // ... 其他操作 ...
    info, err := s.GetOrderCartInfo(ctx, saleBillUuid)  // ❌ 错误：方法内部使用 ctx.GetDB()
    // ... 使用 info ...
})
```

**正确模式**：
```go
// ✅ 正确示例 1: 先设置事务上下文，再使用 ctx.GetDB()
repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    ctx.SetDB(tx)  // ✅ 先设置事务上下文
    db := ctx.GetDB()  // ✅ 现在获取的是 tx
    // ... 使用 db 进行操作 ...
})

// ✅ 正确示例 2: 先设置事务上下文，再调用服务方法
repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    ctx.SetDB(tx)  // ✅ 先设置事务上下文
    // ... 其他操作 ...
    info, err := s.GetOrderCartInfo(ctx, saleBillUuid)  // ✅ 现在会使用 tx
    // ... 使用 info ...
})

// ✅ 正确示例 3: 事务提交后调用服务方法（使用普通 db 连接）
repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    // ... 事务操作 ...
    return nil
}); err != nil {
    return errors.WithMessage(err)
}
// ✅ 事务已提交，使用普通 db 连接是正确的
info, err := s.GetOrderCartInfo(ctx, saleBillUuid)
```

**检查策略**：
1. 搜索所有事务回调函数：`grep -r "Transaction.*func.*gorm.DB"`
2. 检查事务回调函数内部是否调用 `ctx.GetDB()`
3. 检查事务回调函数内部是否调用服务方法（如 `s.xxx()`）
4. 检查这些服务方法内部是否使用 `ctx.GetDB()`
5. 确认调用前是否设置了 `ctx.SetDB(tx)`

**常见问题模式**：
- ❌ 在事务回调函数中使用 `ctx.GetDB()` 但没有先设置 `ctx.SetDB(tx)`
- ❌ 在事务回调函数中调用服务方法，但服务方法内部使用 `ctx.GetDB()`，且调用前未设置 `ctx.SetDB(tx)`
- ✅ 事务提交后调用服务方法是正确的，因为事务已提交，使用普通 `db` 连接是合理的

---

## 安全审计检查项

### 8. SQL 注入漏洞（高危）

#### 8.1 JSON_EXTRACT 路径注入

**检查模式**：
```go
// 搜索模式
"JSON_EXTRACT(.*, '$."+language+"')"
"JSON_UNQUOTE(JSON_EXTRACT(.*, '$."+.*+"'))"
```

#### 8.2 fmt.Sprintf SQL 拼接

**检查模式**：
```go
// 搜索模式
fmt.Sprintf(".*SELECT.*%s.*", .*)
fmt.Sprintf(".*WHERE.*%s.*", .*)
fmt.Sprintf(".*INSERT.*%s.*", .*)
fmt.Sprintf(".*UPDATE.*%s.*", .*)
fmt.Sprintf(".*DELETE.*%s.*", .*)
```

#### 8.3 CREATE/DROP DATABASE 注入

**检查模式**：
```go
// 搜索模式
fmt.Sprintf("CREATE DATABASE %s", .*)
fmt.Sprintf("DROP DATABASE %s", .*)
db.Exec("CREATE DATABASE " + .*)
db.Exec("DROP DATABASE " + .*)
```

### 9. 命令注入漏洞（低危）

#### 9.1 Shell 命令执行

**检查模式**：
```go
// 搜索模式
exec.Command("sh", "-c", .*)
exec.Command("bash", "-c", .*)
exec.Command("cmd", "/c", .*)
```

### 10. 认证和授权漏洞

#### 10.1 JWT 签名算法未验证（高危）

**检查模式**：
```go
// 搜索模式
jwt.ParseWithClaims(.*, func(token *jwt.Token) (interface{}, error) {
    return []byte(secret), nil  // 未验证签名算法
})
```

#### 10.2 Token 过期时间过长（高危）

**检查模式**：
```go
// 搜索模式
GenerateToken(.*, .*, [0-9]{10,}, .*)  // 过期时间超过 7 天（604800 秒）
```

#### 10.3 MD5 用于 Token 完整性验证（中危）

**检查模式**：
```go
// 搜索模式
md5.New()
md5.Sum(.*)
```

### 11. 敏感信息泄露（高危）

#### 11.1 硬编码凭据

**检查模式**：
```go
// 搜索模式
const.*=.*"password|secret|key|token".*"
.*=.*\[\]byte\(".*"\)
```

#### 11.2 API 响应返回密码

**检查模式**：
```go
// 搜索模式
type.*Resp struct {
    .*Password.*string.*`json:"password"`
    .*password.*string.*`json:"password"`
    .*Pwd.*string.*`json:"pwd"`
}
```

#### 11.3 日志泄露密码

**检查模式**：
```go
// 搜索模式
Log.*password|Password|pwd
zap.Any\(".*password.*", .*\)
zap.String\(".*password.*", .*\)
```

### 12. XSS 和输入验证

#### 12.1 CORS 配置过于开放（高危）

**检查模式**：
```go
// 搜索模式
Set\("Access-Control-Allow-Origin", "\*"\)
Set\("Access-Control-Allow-Credentials", "true"\)
```

#### 12.2 缺少 HTTP 安全响应头（中危）

**检查模式**：
```go
// 搜索模式 - 检查是否缺少以下响应头
// X-Frame-Options
// X-Content-Type-Options
// Strict-Transport-Security
// X-XSS-Protection
// Content-Security-Policy
```

#### 12.3 输入参数缺少验证（中危）

**检查模式**：
```go
// 搜索模式 - 检查请求结构体是否缺少 binding 标签
type.*Req struct {
    .*string.*`json:".*"`  // 缺少 binding 标签
}
```

### 13. 文件上传和路径遍历（高危）

#### 13.1 Delete 函数路径遍历

**检查模式**：
```go
// 搜索模式
func.*Delete.*string.*error {
    .*filepath.Join.*fileName.*
    // 未使用 filepath.Base()
    // 未验证路径
}
```

#### 13.2 文件写入路径遍历

**检查模式**：
```go
// 搜索模式
os.WriteFile.*filePath.*
os.Create.*filePath.*
ioutil.WriteFile.*filePath.*
// 未验证路径
```

#### 13.3 文件扩展名验证不足（中危）

**检查模式**：
```go
// 搜索模式
filepath.Ext.*fileName.*
// 只基于扩展名验证，未检查文件内容
```

### 14. 加密和密码安全（高危）

#### 14.1 MD5 用于密码加密

**检查模式**：
```go
// 搜索模式
func.*EncryptPassword.*string.*string {
    md5.New\(\)
    md5.Sum\(\)
}
```

#### 14.2 硬编码 AES 密钥

**检查模式**：
```go
// 搜索模式
.*=.*\[\]byte\(".*SECRET.*KEY.*"\)
.*=.*\[\]byte\(".*AES.*"\)
```

## 输出格式

### 控制台输出

```
代码审查报告
============

目标: main/app/service/order_manage.go
检查时间: 2025-12-19 14:30:00

问题统计:
- 严重: 1 个
- 高: 2 个
- 中: 1 个
- 低: 0 个

问题列表:
========

[严重] order_manage.go:1978-1985
事务中使用外部 db 而非 tx
- 位置: 第 1978 行
- 问题: HandleMemberBalance 调用未检查返回错误
- 建议: 添加错误检查

[高] order_manage.go:874
事务中使用外部 db 而非 tx
- 位置: 第 874 行
- 问题: returnInventory 调用前未设置事务上下文
- 建议: 添加 ctx.SetDB(tx)
...
```

### 报告文件格式（--report）

如果指定 `--report` 参数，会生成 Markdown 格式的详细报告：

```markdown
# 代码审查报告

**目标**: main/app/service/order_manage.go  
**检查时间**: 2025-12-19 14:30:00  
**检查范围**: 事务错误处理

## 问题统计

- 发现问题数: 4 个
- 严重: 1 个
- 高: 2 个
- 中: 1 个
- 低: 0 个

## 问题列表

### [严重] 事务中使用外部 db 而非 tx - 行号 1978-1985

**问题描述**:  
HandleMemberBalance 调用未检查返回错误。

**代码片段**:
```go
s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
    // ...
})
```

**建议修复**:
```go
if err := s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
    // ...
}); err != nil {
    return errors.WithMessage(err)
}
```

...
```

## 执行流程

### Step 1: 解析参数

```yaml
解析 target: 文件或目录路径
解析 --focus: 检查重点（transaction/error-handling/all）
解析 --fix: 是否自动修复
解析 --report: 报告文件路径
```

### Step 2: 识别目标文件

```yaml
IF target 是文件 THEN
    文件列表 = [target]
ELSE IF target 是目录 THEN
    文件列表 = 扫描目录下所有 .go 文件
END IF
```

### Step 3: 执行检查

```yaml
FOR EACH 文件 IN 文件列表:
    1. 读取文件内容
    2. 根据 --focus 执行相应检查
    3. 记录发现的问题
END FOR
```

### Step 4: 生成报告

```yaml
IF --report 指定 THEN
    生成 Markdown 报告文件
ELSE
    输出到控制台
END IF
```

### Step 5: 自动修复（可选）

```yaml
IF --fix 指定 THEN
    FOR EACH 问题 IN 问题列表:
        IF 问题可以自动修复 THEN
            执行修复
            记录修复日志
        END IF
    END FOR
END IF
```

## 检查策略

### 事务检查策略

1. **搜索事务块**：
   ```bash
   grep -r "Transaction\|db\.Begin" --include="*.go"
   ```

2. **检查事务参数使用**：
   - 检查事务回调函数内部是否使用外部 `db`
   - 检查所有 Repository 调用是否使用 `tx`

3. **检查 ctx.GetDB() 使用**：
   - 检查事务回调函数内部是否调用 `ctx.GetDB()`
   - 检查事务回调函数内部是否调用服务方法（如 `s.xxx()`）
   - 检查这些服务方法内部是否使用 `ctx.GetDB()`
   - 确认调用前是否设置了 `ctx.SetDB(tx)`

4. **检查错误处理**：
   - 检查事务调用是否检查返回错误
   - 检查事务内部方法调用是否检查错误

### 错误处理检查策略

1. **搜索方法调用**：
   ```bash
   grep -r "\.\w+\(.*\)$" --include="*.go"
   ```

2. **检查错误检查**：
   - 检查方法签名是否返回 `error`
   - 检查调用是否检查错误

3. **检查 JSON 序列化**：
   ```bash
   grep -r "json\.Marshal\|json\.Unmarshal" --include="*.go"
   ```

### 安全审计检查策略

1. **SQL 注入检查**：
   ```bash
   # JSON_EXTRACT 路径注入
   grep -r "JSON_EXTRACT.*\$\.\"+.*+\"" --include="*.go"
   # fmt.Sprintf SQL 拼接
   grep -r "fmt\.Sprintf.*SELECT\|UPDATE\|DELETE\|INSERT" --include="*.go"
   # CREATE/DROP DATABASE
   grep -r "CREATE DATABASE\|DROP DATABASE" --include="*.go"
   ```

2. **命令注入检查**：
   ```bash
   grep -r "exec\.Command.*sh.*-c\|exec\.Command.*bash.*-c" --include="*.go"
   ```

3. **认证授权检查**：
   ```bash
   # JWT 签名算法验证
   grep -r "jwt\.ParseWithClaims" --include="*.go"
   # Token 过期时间
   grep -r "GenerateToken.*[0-9]\{10,\}" --include="*.go"
   # MD5 Token 验证
   grep -r "md5\.New\|md5\.Sum" --include="*.go"
   ```

4. **敏感信息泄露检查**：
   ```bash
   # 硬编码凭据
   grep -r "const.*=.*password\|secret\|key\|token" --include="*.go" -i
   # API 响应返回密码
   grep -r "Password.*json:\"password\"" --include="*.go"
   # 日志泄露密码
   grep -r "Log.*password\|zap.*password" --include="*.go" -i
   ```

5. **XSS 和输入验证检查**：
   ```bash
   # CORS 配置
   grep -r "Access-Control-Allow-Origin.*\*" --include="*.go"
   # 输入参数验证
   grep -r "type.*Req struct" --include="*.go" -A 10
   ```

6. **文件上传和路径遍历检查**：
   ```bash
   # Delete 函数
   grep -r "func.*Delete.*string.*error" --include="*.go" -A 10
   # 文件写入
   grep -r "os\.WriteFile\|os\.Create\|ioutil\.WriteFile" --include="*.go"
   ```

7. **加密和密码安全检查**：
   ```bash
   # MD5 密码加密
   grep -r "func.*EncryptPassword\|func.*HashPassword" --include="*.go" -A 10
   # 硬编码 AES 密钥
   grep -r "\[\]byte.*SECRET.*KEY\|\[\]byte.*AES" --include="*.go" -i
   ```

## 优先级分类

### 事务和错误处理

| 优先级 | 问题类型 | 修复优先级 |
|--------|---------|-----------|
| **严重** | 事务中使用外部 db 而非 tx | P0 |
| **严重** | 事务中使用 ctx.GetDB() 但未设置事务上下文 | P0 |
| **高** | 事务调用未检查返回错误 | P0 |
| **高** | 调用方法未检查返回错误 | P0 |
| **中** | 调用服务方法前未设置事务上下文 | P1 |
| **中** | Goroutine 中调用方法未检查返回错误 | P1 |
| **低** | JSON 序列化/反序列化未检查错误 | P2 |

### 安全审计

| 优先级 | 问题类型 | 修复优先级 |
|--------|---------|-----------|
| **严重** | SQL 注入漏洞（JSON_EXTRACT 路径注入、fmt.Sprintf SQL 拼接） | P0 |
| **严重** | 路径遍历漏洞（Delete 函数、文件写入） | P0 |
| **严重** | JWT 签名算法未验证 | P0 |
| **严重** | Token 过期时间过长（>7天） | P0 |
| **严重** | 硬编码凭据（密码、密钥、API Secret） | P0 |
| **严重** | API 响应返回密码 | P0 |
| **严重** | 日志泄露密码 | P0 |
| **严重** | CORS 配置过于开放（* + Credentials） | P0 |
| **严重** | MD5 用于密码加密 | P0 |
| **严重** | 硬编码 AES 密钥 | P0 |
| **中** | CREATE/DROP DATABASE 注入 | P1 |
| **中** | MD5 用于 Token 完整性验证 | P1 |
| **中** | 缺少 HTTP 安全响应头 | P1 |
| **中** | 输入参数缺少验证 | P1 |
| **中** | 文件扩展名验证不足 | P1 |
| **低** | Shell 命令执行 | P2 |

## 自动修复能力

### 可自动修复的问题

- ✅ 事务调用未检查返回错误
- ✅ 调用方法未检查返回错误（部分）
- ✅ JSON 序列化/反序列化未检查错误

### 需要人工修复的问题

- ⚠️ 事务中使用外部 db 而非 tx（需要上下文分析）
- ⚠️ 事务中使用 ctx.GetDB() 但未设置事务上下文（需要上下文分析）
- ⚠️ 调用服务方法前未设置事务上下文（需要上下文分析）
- ⚠️ Goroutine 中调用方法未检查返回错误（需要上下文分析）

## 示例

### 示例 1: 检查单个文件

```bash
/code-review main/app/service/order_manage.go
```

输出：
```
代码审查报告
============

目标: main/app/service/order_manage.go
检查时间: 2025-12-19 14:30:00

问题统计:
- 严重: 1 个
- 高: 2 个

[严重] order_manage.go:1978-1985
事务中使用外部 db 而非 tx
...
```

### 示例 2: 检查目录并生成报告

```bash
/code-review main/app/service --report docs/code-review-report.md
```

### 示例 3: 检查并自动修复

```bash
/code-review main/app/service --fix
```

### 示例 4: 安全审计检查

```bash
/code-review main/app/service --focus security
```

输出：
```
代码审查报告
============

目标: main/app/service
检查时间: 2025-12-29 14:30:00
检查范围: 安全审计

问题统计:
- 严重: 5 个
- 高: 3 个
- 中: 2 个
- 低: 1 个

[严重] local.go:55-66
路径遍历漏洞 - Delete 函数
- 位置: 第 55 行
- 问题: fileName 参数未验证，可能包含 ../ 导致路径遍历
- 建议: 使用 filepath.Base() 提取文件名，并验证路径在上传目录内

[严重] jwt.go:38-49
JWT 签名算法未验证
- 位置: 第 38 行
- 问题: jwt.ParseWithClaims 未验证签名算法
- 建议: 添加签名算法验证，只接受 HS256

...
```

### 示例 5: 全面检查（事务 + 错误处理 + 安全审计）

```bash
/code-review main/app/service --focus all --report docs/code-review-report.md
```

## 相关文档

- 代码审查规则: `.cursor/rules/code-review.mdc`
- 事务错误处理扫描报告: `docs/shared/troubleshooting/transaction-error-handling-scan-report.md`
- Go Main 核心约束: `.cursor/rules/go-main.mdc`
- 安全开发规范: `.cursor/rules/security.mdc`

## 错误处理

| 错误类型 | 处理方式 |
|---------|---------|
| 文件不存在 | 提示文件路径错误 |
| 目录不存在 | 提示目录路径错误 |
| 无 Go 文件 | 提示未找到 Go 文件 |
| 解析失败 | 记录错误并继续检查其他文件 |

---

**版本**: v2.0.0  
**创建日期**: 2025-12-19  
**最后更新**: 2025-12-29  
**维护者**: 代码审查组  
**状态**: ✅ MVP + 安全审计

