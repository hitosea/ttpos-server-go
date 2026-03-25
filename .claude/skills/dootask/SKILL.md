---
name: dootask
description: 读取 DooTask 任务详情。当用户使用 /dootask 命令或需要查看任务信息时触发。
---

# /dootask - 读取任务详情

## 使用方式

```bash
/dootask 36917                    # 读取任务 ID 为 36917 的详情
/dootask 22222                    # 读取任务 ID 为 22222 的详情
```

## 参数

- `task_id`: 必填，DooTask 任务编号（纯数字）

## 执行流程

### 1. 参数验证

检查用户传入的 task_id 是否为有效数字。

### 2. 调用 MCP 服务

使用 DooTask MCP 服务获取任务详情：

```yaml
工具: mcp__DooTask__get_task
参数:
  task_id: {用户传入的任务ID}
```

### 3. 格式化输出

成功获取任务后，按以下格式展示：

```markdown
## 📋 任务详情 #{task_id}

**标题**: {name}
**项目**: {project_name}
**看板列**: {column_name}

### 描述
{description}

### 内容
{content}

### 👤 负责人
- {owner_name} (user_id: {owner_userid})

### 🤝 协助人员
- {assist_name} (user_id: {assist_userid})
...

### ⏰ 时间安排
- 开始时间: {start_at}
- 结束时间: {end_at}
- 完成时间: {complete_at}

### 📊 状态信息
- 状态: {flow_item_name}
- 优先级: {p_name}
```

### 4. 错误处理

任务不存在时输出：

```markdown
## ❌ 错误：任务不存在

任务 ID #{task_id} 不存在，请检查：
1. 任务编号是否正确
2. 是否有权限访问该任务
3. 任务是否已被删除

💡 提示：可以使用 mcp__DooTask__list_tasks 查看可访问的任务列表
```

## 相关命令

- `/dev-task` - 开发指挥部
- `/commit` - 生成提交消息
