---
description: Go BMP 模块编码规范 — 编写或修改 ttpos-bmp/ 目录下的 Go 代码时自动加载
globs:
  - "ttpos-bmp/**/*.go"
---

# Go BMP 编码规范

## 关键约束

- **禁止修改自动生成文件**：`dao/`, `model/entity/`, `model/do/`, `service/`
- 使用 `gerror` 处理错误（不用标准库 errors）
- 业务逻辑写在 `internal/logic/` 目录
- 使用 `dao` 层操作数据库
- 事务使用 `g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error { ... })`

## 子模块代码生成

在 `ttpos-bmp/app/ttpos-xxx/` 目录下：

```bash
make dao      # 生成 DAO/DO/Entity
make ctrl     # 生成控制器/SDK
make service  # 生成服务接口
make pb       # 生成 protobuf Go 代码
```
