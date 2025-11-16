# 明天的工作计划

> 补充后端文档体系缺失部分

**创建日期**: 2025-11-16  
**预计开始**: 2025-11-17  
**预计完成**: 2025-11-19 (3天完成 P0)

---

## 📋 快速概览

**总缺失**: 约 51+ 个文档  
**P0 (必须)**: 20 个文档  
**P1 (建议)**: 20+ 个文档  
**P2 (可选)**: 10+ 个文档

---

## 🎯 Day 1: Agent 核心 (2025-11-17)

### 上午: Phase 1 - Agent 核心模板 (4个)

```bash
# 创建文件
cd /Users/ben/projects/ttpos-server-go
touch docs/agent/templates/requirements-template.md
touch docs/agent/templates/design-template.md
touch docs/agent/templates/tasks-template.md
touch docs/agent/templates/proposal-template.md
```

**参考源文件**:
- `/Users/ben/projects/ttpos-flutter/docs/agent/templates/requirements-template.md`
- `/Users/ben/projects/ttpos-flutter/docs/agent/templates/design-template.md`
- `/Users/ben/projects/ttpos-flutter/docs/agent/templates/tasks-template.md`
- `/Users/ben/projects/ttpos-flutter/docs/agent/templates/proposal-template.md`

**适配要点**:
- 将 Flutter/Dart 代码示例改为 Go/PHP/Vue
- 添加数据库设计部分
- 添加 API 设计部分 (REST + gRPC)
- 添加微服务设计部分

**预计时间**: 3-4 小时

---

### 下午: Phase 2 - 测试流程工作流 (3个)

```bash
# 创建文件
touch docs/agent/workflows/test-submission.md
touch docs/agent/workflows/test-verification.md
touch docs/agent/workflows/defect-management.md
```

**参考源文件**:
- `/Users/ben/projects/ttpos-flutter/docs/agent/workflows/test-submission.md`
- `/Users/ben/projects/ttpos-flutter/docs/agent/workflows/test-verification.md`
- `/Users/ben/projects/ttpos-flutter/docs/agent/workflows/defect-management.md`

**适配要点**:
- 测试环境配置 (Docker, MySQL, Redis, Nacos)
- API 测试工具 (Postman, curl, grpcurl)
- 测试覆盖率要求 (Go: go test -cover, PHP: PHPUnit)
- 日志查看 (SkyWalking, Go log, PHP log)

**预计时间**: 2-3 小时

---

## 🎯 Day 2: 人类核心 (2025-11-18)

### 上午: Phase 3 - 核心架构文档 (3个)

```bash
# 创建文件
touch docs/human/architecture/overview.md
touch docs/human/architecture/modules.md
touch docs/human/architecture/code-style-guide.md
```

**参考源文件**:
- `/Users/ben/projects/ttpos-flutter/docs/human/architecture/overview.md`
- `/Users/ben/projects/ttpos-flutter/docs/human/architecture/modules.md`
- `/Users/ben/projects/ttpos-flutter/docs/human/architecture/code-style-guide.md` (632行)

**适配要点**:
- overview: 三语言架构图 (Go/PHP/Vue)
- modules: main、admin、ttpos-bmp 关系说明
- code-style-guide: 参考 .cursor/rules/golang.mdc 和 php.mdc 扩展

**预计时间**: 4-5 小时

---

### 下午: Phase 4 - 测试标准体系 (6个)

```bash
# 创建目录和文件
mkdir -p docs/human/testing/standards
mkdir -p docs/human/testing/examples

touch docs/human/testing/README.md
touch docs/human/testing/standards/README.md
touch docs/human/testing/standards/api-testing.md
touch docs/human/testing/standards/service-testing.md
touch docs/human/testing/standards/controller-testing.md
touch docs/human/testing/standards/model-testing.md
```

**参考源文件**:
- `/Users/ben/projects/ttpos-flutter/docs/human/testing/`

**适配要点**:
- api-testing: REST API 和 gRPC 测试
- service-testing: Go Service 层测试 (Mock、Table-Driven)
- controller-testing: Go Gin 和 PHP ThinkPHP 控制器测试
- model-testing: Go Struct 和 PHP Model 测试

**预计时间**: 3-4 小时

---

## 🎯 Day 3: 共享文档 (2025-11-19)

### Phase 5 - 共享文档核心 (4个)

```bash
# 创建文件
touch docs/shared/api/conventions.md
touch docs/shared/troubleshooting/common-issues.md
touch docs/human/business/glossary.md
touch docs/human/guides/installation.md
```

**参考源文件**:
- `/Users/ben/projects/ttpos-flutter/docs/shared/api/conventions.md`
- `/Users/ben/projects/ttpos-flutter/docs/shared/troubleshooting/common-issues.md`
- `/Users/ben/projects/ttpos-flutter/docs/human/business/glossary.md`
- `/Users/ben/projects/ttpos-flutter/docs/human/guides/installation.md`

**适配要点**:
- conventions.md: REST API 和 gRPC 规范，响应格式
- common-issues.md: 数据库、Redis、Nacos、gRPC 常见问题
- glossary.md: 后端术语 + 餐饮业务术语
- installation.md: Go、PHP、MySQL、Redis、Docker 环境配置

**预计时间**: 3-4 小时

---

## 📊 进度跟踪

### Day 1 (2025-11-17)
- [ ] Phase 1: Agent 核心模板 (4个)
- [ ] Phase 2: 测试流程工作流 (3个)

### Day 2 (2025-11-18)
- [ ] Phase 3: 核心架构文档 (3个)
- [ ] Phase 4: 测试标准体系 (6个)

### Day 3 (2025-11-19)
- [ ] Phase 5: 共享文档核心 (4个)

**总计**: 20 个 P0 文档

---

## 🔧 工作方式建议

### 方式 1: 从前端复制并适配
```bash
# 示例
cp /Users/ben/projects/ttpos-flutter/docs/agent/templates/requirements-template.md \
   /Users/ben/projects/ttpos-server-go/docs/agent/templates/requirements-template.md

# 然后用 Cursor 适配内容
```

### 方式 2: 让 Agent 帮助
```
告诉 Agent:
"参考前端的 requirements-template.md，为后端创建一个适配版本，
要求支持 Go/PHP/Vue，包含数据库设计、API 设计部分"
```

### 方式 3: 组合方式（推荐）
1. 先从前端复制
2. 让 Agent 自动适配
3. 人工审查和补充

---

## ✅ 完成标准

每个文档需要达到：
- ✅ 格式符合规范（受众标识 🤖/👤/📚）
- ✅ 代码示例适配后端技术栈
- ✅ 交叉引用正确
- ✅ 无死链接
- ✅ 无待填充占位符

---

## 📚 参考资源

### 前端文档位置
```
/Users/ben/projects/ttpos-flutter/docs/
```

### 后端规范文件
```
.cursor/rules/AGENT_QUICK_REF.mdc
.cursor/rules/golang.mdc
.cursor/rules/php.mdc
.cursor/rules/vue.mdc
```

### 详细缺失报告
```
docs/MISSING_DOCS_REPORT.md
```

---

## 🎯 成功指标

完成后应实现：
1. ✅ Agent 可以使用 `/create-spec` 创建 Spec
2. ✅ Agent 可以使用 `/propose` 创建提案
3. ✅ Agent 可以执行测试流程工作流
4. ✅ 新人可以通过文档快速上手
5. ✅ 开发者可以查阅完整的架构和测试文档

---

## 💡 小贴士

### 适配时注意
- 将 "Flutter" 改为 "Go/PHP/Vue"
- 将 "GetX" 改为 "Go Service/Repository"
- 将 "Dio" 改为 "Gin/ThinkPHP HTTP Client"
- 将 "Freezed" 改为 "Go Struct/PHP Model"
- 添加数据库相关内容
- 添加微服务相关内容

### 可直接复用
- 通用流程说明
- 文档结构
- 检查清单逻辑
- 命名规范

### 需要重写
- 代码示例
- 技术细节
- 工具链说明
- 环境配置

---

**开始日期**: 2025-11-17  
**预计完成**: 2025-11-19  
**加油！** 💪

