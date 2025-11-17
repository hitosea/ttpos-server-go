# Graphiti Episode 模板

> 复制本文件后填写，命名遵循 `{类型}-{主题}-{YYYY-MM}`，完成后交由责任人通过 Graphiti MCP 入库。

## 元信息

- **Episode 名称**：``
- **Episode 类型**：`concept | qa | experience | relation | evolution`
- **Group ID**：`ttpos-main | ttpos-php | ttpos-database | ttpos-security | ttpos-troubleshooting | ttpos-integration`
- **涉及技术栈**：Go (main/) / PHP (admin/) / GoFrame (ttpos-bmp/) / Vue / Database
- **适用迭代或版本**：`Sprint xx` / `v2.x`
- **状态**：`draft | ready | published`
- **Owner**：`@姓名`
- **协作者 / 审核人**：`@姓名`
- **Source 链接**：
  - ``
- **Related Docs**（相对路径）：
  - `docs/shared/troubleshooting/...`
  - `docs/shared/specs/...`
- **Related Tickets/Specs**：
  - `story-...`
  - `task-...`

## 背景

> 描述问题出现的业务场景、触发条件、受影响模块（如 `main/app/service/payment`、`admin/database/migrations/`）。

## 关键结论

> 用列表记录本次沉淀得到的事实、约束或注意事项。
- 
- 

## 操作步骤 / 诊断流程

> 记录可复用的排查流程或操作命令，保持步骤化。
1. 
2. 
3. 

## 解决方案与代码参考

> 需要时附上关键代码/命令片段，标注文件路径。

```bash
# 示例：回滚最近一次迁移
cd path/ttpos-server-go/admin && php think migrate:rollback --step=1
```

```go
// 示例：main/app/service/payment/payment_srv.go
```

## 预防与后续行动

- 

## 版本记录

| 日期       | 修改人 | 说明       |
| ---------- | ------ | ---------- |
| YYYY-MM-DD | @姓名  | 初始创建   |
| YYYY-MM-DD | @姓名  | 更新内容… |

> Episode 入库后请在关联的 troubleshooting 文档或 Spec 中补充 `Related Episode: {名称}`，并在活动日志中自动记录。***

