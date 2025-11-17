# gRPC 服务模板

> 🤖 Agent 视角模板：适用于 `ttpos-bmp/app/ttpos-*/` 中创建 Protobuf、服务实现、Nacos 注册及 Graphiti/活动日志提醒。

---

## 元信息

| 字段       | 内容                                                              |
| ---------- | ----------------------------------------------------------------- |
| 服务名称   | `ttpos-{module}-{feature}`                                        |
| Proto 文件 | `ttpos-bmp/app/ttpos-{service}/manifest/protobuf/{feature}.proto` |
| 负责人     | `@`                                                               |
| 关联 Spec  | `task-all-{feature}` / `story-{module}-{feature}`                 |
| 版本       | `v1.0.0`                                                          |

---

## 1. 背景与目标

- 业务场景、调用方、性能/可靠性要求。
- 与 HTTP API 的边界与依赖。

---

## 2. Protobuf 设计

```proto
syntax = "proto3";
package ttpos.{service};

service {Feature}Service {
  rpc DoAction(DoActionRequest) returns (DoActionResponse);
}

message DoActionRequest {
  uint64 order_uuid = 1;
}

message DoActionResponse {
  uint32 code = 1;
  string message = 2;
}
```

- **命名规范**：小写包名、驼峰 Service；字段 snake_case + 类型。
- **文件结构**：`manifest/protobuf/`, `manifest/sql/`, `manifest/config/`.

---

## 3. 代码生成

```bash
cd ttpos-bmp/app/ttpos-{service}
make proto
# 或
protoc --go_out=. --go-grpc_out=. manifest/protobuf/{feature}.proto
```

- 确认生成的 `internal/controller/rpc/`, `internal/logic/`, `internal/svc/` 文件结构就绪。

---

## 4. 服务实现

```go
func (s *DoActionLogic) DoAction(ctx context.Context, req *pb.DoActionRequest) (*pb.DoActionResponse, error) {
    // 参数校验
    // 调用 Repository / Logic
    // 返回响应
}
```

- 依赖注入通过 `svc.ServiceContext`。
- 错误统一由 `github.com/pkg/errors` 包装。

---

## 5. 配置与注册

- `manifest/config/config.yaml`
  ```yaml
  grpc:
    name: ttpos-{service}
    listen: 0.0.0.0:9001
  registries:
    nacos:
      endpoint: ...
  ```
- 注册到 Nacos / Consul，记录服务名与命名空间。

---

## 6. 客户端调用

```go
conn, _ := grpc.DialContext(ctx, "ttpos-{service}", grpc.WithInsecure())
client := pb.New{Feature}ServiceClient(conn)
resp, err := client.DoAction(ctx, &pb.DoActionRequest{OrderUuid: 1})
```

- 若 main/ 需要调用，在 `main/app/service/*` 中封装 Client。

---

## 7. 测试

- 单元测试：`internal/logic/*_test.go`。
- 集成测试：模拟 gRPC 调用，验证路由/配置。
- 性能测试：QPS 指标、超时设置。

---

## 8. 文档与依赖

- 在 Spec `design.md` 中记录 Proto、gRPC 调用流程。
- 若暴露 HTTP Gateway，更新 `docs/shared/api/{module}_api.md`。
- 列出依赖的数据库/队列/缓存。

---

## 9. Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 如有 gRPC 遇到的问题，使用 `docs/agent/templates/graphiti-episode.md`。

---

**最后更新**：2025-11-17
