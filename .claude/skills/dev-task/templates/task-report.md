# 任务编排报告

> 生成时间: {timestamp}

## 目标

{goal}

## 执行摘要

| ID | 任务 | Agent | 状态 | 耗时 | 摘要 |
|----|------|-------|------|------|------|
{task_rows}

**总计**: {total_tasks} 个任务, {completed} 完成, {failed} 失败

## Layer 执行详情

### Layer {n}

**并行任务**: {layer_task_names}
**策略**: {parallel|sequential}

#### {task_id}: {task_name}

**Agent**: {agent_type}
**隔离**: {worktree|无}
**输出文件**: {output_file}
**摘要**: {summary}

{如果有关键发现，列出 2-3 条}

---

## 综合分析

{基于所有子任务产出的综合分析，2-3 段落}

## 产出文件索引

| 文件 | 来源任务 | 说明 |
|------|---------|------|
{file_index_rows}

## 后续建议

{基于执行结果的建议，如果有失败任务则优先列出修复建议}
