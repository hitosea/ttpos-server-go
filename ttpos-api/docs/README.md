# TTPOS API 文档中心

欢迎来到 TTPOS API 文档中心！本目录包含了 `ttpos-api` 项目的所有详细文档。

## 📁 目录结构

```
doc/
├── README.md                # 本文档（文档导航）
├── CHANGELOG.md             # 版本更新记录
├── guide/                   # 📘 使用指南
│   ├── USAGE.md            # 详细使用指南
│   └── INTEGRATION.md      # 集成指南
└── reference/               # 📚 参考文档
    ├── MODULES.md          # 模块结构说明
    ├── ERRORS.md           # 错误定义规范
    └── WEBSOCKET.md        # WebSocket 详细文档
```

## 📖 文档导航

### 🚀 快速开始

如果您是第一次使用 `ttpos-api`，建议按以下顺序阅读：

1. **[模块说明 (reference/MODULES.md)](reference/MODULES.md)** ⭐ 推荐首先阅读
   - 了解项目的模块结构
   - 理解各模块的职责
   - 掌握目录组织方式

2. **[使用指南 (guide/USAGE.md)](guide/USAGE.md)**
   - 基本使用方法
   - 代码示例
   - 最佳实践

3. **[集成指南 (guide/INTEGRATION.md)](guide/INTEGRATION.md)**
   - 如何在您的项目中集成 `ttpos-api`
   - 各服务的集成步骤
   - 常见问题解决

### 📚 文档分类

#### 📘 使用指南 (guide/)

实用的操作指南和教程：

| 文档 | 说明 | 适用场景 |
|------|------|----------|
| [USAGE.md](guide/USAGE.md) | 详细使用指南 | 日常开发参考 |
| [INTEGRATION.md](guide/INTEGRATION.md) | 服务集成指南 | 集成到新项目 |

#### 📚 参考文档 (reference/)

详细的技术参考资料：

| 文档 | 说明 | 适用场景 |
|------|------|----------|
| [MODULES.md](reference/MODULES.md) | 模块结构和职责划分 | 了解项目架构 |
| [ERRORS.md](reference/ERRORS.md) | 错误定义和使用规范 | 错误处理和定义 |
| [WEBSOCKET.md](reference/WEBSOCKET.md) | WebSocket 消息详解 | WebSocket 开发 |

#### 📝 其他文档

| 文档 | 说明 | 适用场景 |
|------|------|----------|
| [CHANGELOG.md](CHANGELOG.md) | 版本更新记录 | 了解版本变化 |

### 🎯 按场景查找

#### 我想了解项目结构

→ 阅读 **[reference/MODULES.md](reference/MODULES.md)**

这份文档会告诉您：
- 项目有哪些模块
- 每个模块的作用
- 文件如何组织
- 如何找到您需要的代码

#### 我想使用消息结构体

→ 阅读 **[guide/USAGE.md](guide/USAGE.md)**

这份文档包含：
- 如何创建消息
- 如何验证消息
- 如何序列化/反序列化
- 完整的代码示例

#### 我想处理错误

→ 阅读 **[reference/ERRORS.md](reference/ERRORS.md)**

这份文档说明：
- 错误如何组织
- 公共错误 vs 模块错误
- 如何定义新错误
- 如何使用错误

#### 我想集成到我的服务

→ 阅读 **[guide/INTEGRATION.md](guide/INTEGRATION.md)**

这份文档指导：
- 如何添加依赖
- 如何配置 go.mod
- 如何导入和使用
- 常见集成问题

#### 我想开发 WebSocket 功能

→ 阅读 **[reference/WEBSOCKET.md](reference/WEBSOCKET.md)**

这份文档详细介绍：
- WebSocket 消息类型
- 每种消息的用途
- 消息字段说明
- 使用示例

#### 我想查看版本更新

→ 阅读 **[CHANGELOG.md](CHANGELOG.md)**

这份文档记录：
- 每个版本的变更
- 新增功能
- Bug 修复
- 破坏性变更

## 📝 文档规范

### 文档命名

- 使用大写字母 + 下划线命名（如 `MODULES.md`）
- 使用描述性的名称
- 保持简洁明了

### 文档结构

每份文档应包含：
1. **标题**：清晰的文档标题
2. **目录**：方便快速定位（可选）
3. **正文**：详细的内容说明
4. **示例**：代码示例（如适用）
5. **相关链接**：指向其他相关文档

### 文档更新

- 添加新功能时，同步更新相关文档
- 修改现有功能时，更新对应文档
- 定期审查文档的准确性
- 在 `CHANGELOG.md` 中记录重要变更

## 🔍 快速查找

### 按关键词查找

- **消息结构体** → guide/USAGE.md, reference/MODULES.md
- **错误处理** → reference/ERRORS.md
- **WebSocket** → reference/WEBSOCKET.md
- **集成** → guide/INTEGRATION.md
- **模块** → reference/MODULES.md
- **示例** → 各模块的 `examples/` 目录

### 按文件类型查找

- **`.go` 文件** → 源代码在各模块目录下
- **示例代码** → `ttpos-websocket/examples/`, `ttpos-message/examples/`
- **文档** → `doc/` 目录（当前目录）

## 💡 贡献文档

如果您想改进文档，请：

1. Fork 项目
2. 创建文档分支
3. 修改或添加文档
4. 提交 Pull Request

### 文档贡献指南

- 使用清晰的中文
- 提供代码示例
- 添加必要的图表（如适用）
- 保持格式一致
- 检查拼写和语法

## 📞 获取帮助

如果文档中没有找到您需要的信息：

1. 查看 [示例代码](../ttpos-websocket/examples/)
2. 查看源代码注释
3. 联系项目维护者
4. 提交 Issue

## 🔗 相关资源

- [项目主页](../README.md)
- [ttpos-websocket 示例](../ttpos-websocket/examples/README.md)
- [Go 官方文档](https://golang.org/doc/)
- [GoFrame 文档](https://goframe.org)

---

**最后更新**: 2025-11-15

**维护者**: TTPOS 开发团队

