# 功能开发工作流（后端版）

> Agent 执行清单：从需求到上线的完整流程

---

## 流程概览

```
历史搜索 → SP评估 → 创建Spec → 选择技术栈 → 代码实现 → 测试 → 提测 → 验收 → 上线
```

**时间**: SP1-3: 1-2天, SP5: 2-3天

---

## 前置条件

- [ ] Go 1.21+ 已安装
- [ ] PHP 8.1+ 已安装
- [ ] Node.js 18+ 已安装（如需前端开发）
- [ ] Docker 环境已配置
- [ ] 已了解项目架构和规范

参考: `README.md`, `.cursor/rules/golang.mdc`, `.cursor/rules/php.mdc`

---

## Step 1: 搜索历史经验 (5-10min)

### 检索路径
```yaml
1. Graphiti: query="{功能} implementation backend"
2. Codebase: 相似实现
3. Specs: docs/shared/specs/ 中的类似功能
```

### 输出
- [ ] 可复用代码
- [ ] 常见陷阱
- [ ] 实现建议

---

## Step 2: 评估 Story Point (10-30min)

### 评估维度
- Go 开发工作量
- PHP 开发工作量（如需要）
- 数据库设计工作量
- 第三方集成工作量
- 测试复杂度

参考: `.cursor/rules/scrum_story_point.mdc`

### 决策
```
IF SP > 5 THEN 拆分需求 → 重新评估
ELSE 继续 Step 3
```

---

## Step 3: 创建 Spec (20-60min)

### 命名
```bash
格式: story-{module}-{feature}
示例: story-order-quick-payment
```

### 创建文件
```bash
docs/shared/specs/story-{module}-{feature}/
├── requirements.md  # 需求+AC
├── design.md        # 设计+架构
└── tasks.md         # 任务清单 ⭐
```

### tasks.md 结构 ⭐
```markdown
## Phase 1: 核心实现

- [ ] 1.1 {任务标题}
  - File: {文件路径}
  - Purpose: {任务目的}
  - Requirements: {关联需求编号}
  - Leverage: {可复用代码}
  - Language: Go / PHP / Vue  ← 新增
```

**任务颗粒度**: 1-4小时/任务

---

## Step 4: 代码实现 (主要时间)

### 创建分支
```bash
git checkout -b feature/story-{module}-{feature}
```

### 执行任务循环 ⭐

**FOR EACH 未完成任务 IN tasks.md:**

1. **读取任务信息**
   - 任务编号 (1.1, 1.2...)
   - Requirements 关联
   - Leverage 可复用代码
   - Language 标识 ← 新增

2. **确定技术栈**
   ```yaml
   IF Language == "Go (main/)" THEN
     参考: .cursor/rules/golang.mdc
     目录: main/app/
   
   ELSE IF Language == "Go (ttpos-bmp)" THEN
     参考: ttpos-bmp/.cursor/rules/go-rules.mdc
     目录: ttpos-bmp/app/
   
   ELSE IF Language == "PHP" THEN
     参考: .cursor/rules/php.mdc
     目录: admin/app/
   
   ELSE IF Language == "Vue" THEN
     参考: .cursor/rules/vue.mdc
     目录: admin/views/
   ```

3. **实现代码（Go 示例）**
   - 遵循规范 (golang.mdc):
     * [ ] 接口以 I 开头
     * [ ] 实现以 Impl 结尾
     * [ ] 不使用 panic，返回 error
     * [ ] URL 使用 snake_case
     * [ ] try-catch 捕获 error 和 stackTrace
     * [ ] 所有注释使用中文

4. **数据库迁移（如需要）**
   ```bash
   # 创建迁移文件
   cd admin
   php think make:migration AddOrderTable
   
   # 编写迁移代码（遵循 php.mdc 规范）
   # - 时间字段: int 类型, _time 结尾
   # - 金额字段: decimal(20,8)
   # - UUID字段: biginteger
   
   # 更新 Go model
   # 更新 seeds 文件
   
   # 执行迁移
   php think migrate:run
   ```

5. **验证**
   ```bash
   # Go 代码检查
   cd main
   go vet ./...
   gofmt -w .
   
   # PHP 代码检查
   cd admin
   composer lint
   
   # Vue 代码检查
   npm run lint
   ```

6. **标记完成**
   ```markdown
   - [x] 1.1 {任务标题}
   ```

7. **提交**
   ```bash
   git commit -m "feat(order): 完成任务 1.1 - {标题}
   
   Task: 1.1
   Requirements: 1.1, 2.3
   Language: Go"
   ```

**REPEAT 直到所有任务 [x]**

### 核心规范（Go）
- 接口以 `I` 开头，实现以 `Impl` 结尾
- 不使用 panic，返回 error
- URL 使用 snake_case (不用 kebab-case)
- data 不能是 null 或数组
- 所有注释使用中文

参考: `.cursor/rules/golang.mdc`, `.cursor/rules/php.mdc`

---

## Step 5: 编写测试 (必须)

### 覆盖率要求

| 模块 | 要求 |
|------|------|
| main/service | 70%+ |
| main/repository | 80%+ |
| ttpos-bmp/logic | 70%+ |
| 支付/订单 | 100% |

### 执行（Go）
```bash
# 运行测试
cd main
go test ./...

# 检查覆盖率
go test -cover ./...
```

### 执行（PHP）
```bash
cd admin
php think test
```

参考: `.cursor/rules/testing.mdc`

---

## Step 6: 更新文档

### 检查清单
- [ ] API 变更 → `docs/shared/api/`
- [ ] 新模块 → `docs/human/architecture/`
- [ ] 数据库迁移 → admin/database/migrations/
- [ ] CHANGELOG.md 更新

---

## Step 7: 提交代码

### 自测
```yaml
必测:
  - 核心功能 (按AC)
  - 正常流程 (≥3次)
  - 异常场景
  - 边界条件
```

### 提交 PR
```bash
git push origin feature/story-{module}-{feature}
```

### PR 检查清单
- [ ] Commit 消息符合规范
- [ ] 关联了对应的 Issue
- [ ] 所有测试通过
- [ ] API 响应格式正确
- [ ] 数据库迁移文件已创建
- [ ] Go model 已同步更新
- [ ] 代码审查通过

---

## 检查清单 (Checklist)

### 任务分析
- [ ] 历史经验已搜索
- [ ] SP 评估 ≤ 5
- [ ] Spec 已创建

### 代码实现
- [ ] 技术栈已确定
- [ ] 分支已创建
- [ ] 代码符合规范
- [ ] 所有任务 [x]
- [ ] API 响应格式正确

### 数据库（如有）
- [ ] 迁移文件已创建
- [ ] Go model 已更新
- [ ] seeds 已更新
- [ ] 迁移已执行

### 测试
- [ ] 单元测试已编写
- [ ] 测试覆盖率达标
- [ ] 所有测试通过

### 提交
- [ ] PR 已创建
- [ ] Code Review 通过
- [ ] PR 已合并

---

## 常见问题

### Q: 如何选择 Go 服务？

**A**: 
- **main/**: 主要业务逻辑（订单、支付、会员等）
- **ttpos-bmp/**: 微服务模块（ERP、管理功能）

判断标准：
1. 如果是核心业务 → main/
2. 如果是管理功能 → ttpos-bmp/
3. 不确定时 → 查看 `docs/human/architecture/`

### Q: 数据库迁移必须在 PHP 中吗？

**A**: 是的。所有数据库迁移文件都在 `admin/database/migrations/` 中管理（PHP Phinx），但需要同步更新 Go model。

### Q: API 响应格式有什么要求？

**A**: 
- data 不能是 null 或数组
- 必须是 `{code, message, data{list, meta}}`
- 分页信息放在 meta 中

详见: `.cursor/rules/golang.mdc`

---

## 相关资源

### 规范文件
- `.cursor/rules/AGENT_QUICK_REF.mdc` - Agent速查表
- `.cursor/rules/golang.mdc` - Go开发规范 ⭐⭐⭐
- `.cursor/rules/php.mdc` - PHP开发规范 ⭐⭐⭐
- `.cursor/rules/vue.mdc` - Vue开发规范
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - GoFrame规范

### 工作流
- [数据库迁移工作流](./database-migration.md) - 数据库操作
- [微服务集成工作流](./microservice-integration.md) - gRPC服务

### 模板
- `docs/agent/templates/requirements-template.md` - 需求模板
- `docs/agent/templates/design-template.md` - 设计模板
- `docs/agent/templates/tasks-template.md` - 任务模板

---

**最后更新**: 2025-11-16
**维护者**: 后端开发组

