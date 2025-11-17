# MessageTemplate 实体模型说明

## 基本信息

- **实体名称**: MessageTemplate
- **表名**: message_template
- **所属模块**: ttpos-message
- **描述**: 消息模板实体，用于管理邮件和短信模板

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | uint64 | 模板ID | 主键 |
| Uuid | string | 模板UUID | 唯一标识 |
| TemplateName | string | 模板名称 | |
| TemplateType | string | 模板类型 | email/sms |
| TemplateSubject | string | 模板主题 | 邮件用 |
| TemplateContent | string | 模板内容 | 支持变量 |
| TemplateArgs | string | 模板参数定义 | JSON格式 |
| Status | int | 状态 | 0-禁用 1-启用 |
| Remark | string | 备注 | |
| CreatedAt | int | 创建时间 | 时间戳 |
| UpdatedAt | int | 更新时间 | 时间戳 |
| DeletedAt | int | 删除时间 | 时间戳，软删除 |

## 关联关系

### 关联实体
- **Uuid** → MessageRecord.TemplateUuid（消息记录关联模板）

## 数据流分析

### 数据来源
- 消息模板配置信息
- 通过管理后台创建和配置

### 数据流向
1. **模板创建流程**:
   - 在系统中创建消息模板（邮件或短信）
   - 配置模板内容，支持变量占位符
   - 定义模板参数（TemplateArgs）

2. **消息发送流程**:
   - 根据业务需求选择对应的模板（通过 TemplateUuid）
   - 使用模板参数渲染模板内容
   - 创建 MessageRecord 记录
   - 发送消息（邮件或短信）

### 业务场景
- 邮件模板管理
- 短信模板管理
- 模板变量替换
- 模板启用/禁用控制

## 索引建议

- 主键索引: Id
- 唯一索引: Uuid
- 普通索引: TemplateType（类型查询）
- 普通索引: Status（状态查询）

## 注意事项

1. TemplateContent 支持变量占位符，需要与 TemplateArgs 配合使用
2. TemplateSubject 仅用于邮件模板
3. 使用软删除机制（DeletedAt）
4. Status 字段控制模板可用性

