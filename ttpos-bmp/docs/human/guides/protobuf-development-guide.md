# Protobuf 开发规范

> gRPC 服务接口定义和开发规范

## 📋 概述

本指南介绍在 TTPOS 业务中台项目中使用 Protocol Buffers (Protobuf) 定义 gRPC 服务接口的规范和最佳实践。

## 📁 文件组织

### 目录结构
```
manifest/protobuf/
├── svc/                    # 服务定义
│   ├── user.proto         # 用户服务
│   ├── order.proto        # 订单服务
│   └── ...
├── entity/                 # 实体定义
│   ├── user.proto         # 用户实体
│   ├── order.proto        # 订单实体
│   └── ...
└── common/                 # 公共定义
    ├── common.proto       # 通用消息
    └── error.proto        # 错误定义
```

### 文件命名
- 服务文件: `{service}.proto`
- 实体文件: `{entity}.proto`
- 使用小写字母和下划线分隔

## 🔧 基本语法

### 文件结构
```protobuf
syntax = "proto3";

package ttpos.bmp.[module].v1;
option go_package = "ttpos-bmp/api/ttpos-[module]/v1;v1";

// 导入依赖
import "common/common.proto";

// 服务定义
service [ServiceName] {
  // RPC 方法
}

// 消息定义
message [MessageName] {
  // 字段定义
}
```

### 包命名
```protobuf
// 包命名规范
package ttpos.bmp.manager.v1;    // 管理模块 v1
package ttpos.bmp.shop.v1;       // 门店模块 v1
package ttpos.bmp.erp.v1;        // ERP模块 v1
package ttpos.bmp.takeout.v1;    // 外送模块 v1
package ttpos.bmp.message.v1;    // 消息模块 v1
```

## 📝 消息定义

### 字段类型
```protobuf
message User {
  // 基本类型
  string username = 1;
  int64 user_id = 2;
  bool is_active = 3;
  double balance = 4;

  // 枚举类型
  UserStatus status = 5;

  // 嵌套消息
  Address address = 6;

  // 重复字段
  repeated string roles = 7;

  // 映射类型
  map<string, string> metadata = 8;
}
```

### 枚举定义
```protobuf
enum UserStatus {
  USER_STATUS_UNSPECIFIED = 0;  // 必须从 0 开始
  USER_STATUS_ACTIVE = 1;
  USER_STATUS_INACTIVE = 2;
  USER_STATUS_SUSPENDED = 3;
}
```

### 嵌套消息
```protobuf
message Address {
  string street = 1;
  string city = 2;
  string state = 3;
  string zip_code = 4;
  string country = 5;
}
```

## 🌐 服务定义

### RPC 方法
```protobuf
service UserService {
  // 简单 RPC
  rpc GetUser(GetUserReq) returns (GetUserRes);

  // 服务端流式 RPC
  rpc ListUsers(ListUsersReq) returns (stream ListUsersRes);

  // 客户端流式 RPC
  rpc CreateUsers(stream CreateUserReq) returns (CreateUsersRes);

  // 双向流式 RPC
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}
```

### 请求/响应消息
```protobuf
// 请求消息
message GetUserReq {
  int64 user_id = 1;
}

// 响应消息
message GetUserRes {
  User user = 1;
}

// 分页请求
message ListUsersReq {
  int32 page = 1;
  int32 page_size = 2;
  string keyword = 3;
}

// 分页响应
message ListUsersRes {
  repeated User users = 1;
  int32 total = 2;
  int32 page = 3;
  int32 page_size = 4;
}
```

## 🔄 公共定义

### 通用消息
```protobuf
// common/common.proto
message Empty {}

// 分页参数
message Pagination {
  int32 page = 1;
  int32 page_size = 2;
}

// 排序参数
message Sort {
  string field = 1;
  string order = 2;  // "asc" or "desc"
}

// 时间范围
message TimeRange {
  int64 start_time = 1;
  int64 end_time = 2;
}
```

### 错误定义
```protobuf
// common/error.proto
enum ErrorCode {
  ERROR_CODE_UNSPECIFIED = 0;
  ERROR_CODE_NOT_FOUND = 1;
  ERROR_CODE_INVALID_ARGUMENT = 2;
  ERROR_CODE_PERMISSION_DENIED = 3;
  ERROR_CODE_INTERNAL_ERROR = 4;
}

message Error {
  ErrorCode code = 1;
  string message = 2;
  map<string, string> details = 3;
}
```

## 📋 设计规范

### 字段编号
- 字段编号 1-15: 占用 1 个字节
- 字段编号 16-2047: 占用 2 个字节
- 为未来扩展预留字段编号

### 字段命名
- 使用小写字母和下划线: `user_name`
- 保持一致性
- 使用有意义的名称

### 消息大小
- 避免过大的消息
- 对于大数据使用流式传输
- 考虑网络传输效率

## 🛠️ 代码生成

### 生成命令
```bash
# 生成 Go 代码
gf gen pb

# 生成特定文件
gf gen pb manifest/protobuf/svc/user.proto
```

### 生成的文件
```
api/
├── ttpos-[module]/
│   └── v1/
│       ├── [service]_grpc.pb.go     # gRPC 服务代码
│       ├── [service].pb.go          # Protobuf 消息代码
│       └── [service]_grpc_mock.go   # Mock 测试代码
```

## 🔧 使用示例

### 服务实现
```go
// internal/controller/rpc/user.go
type cUser struct{}

func (c *cUser) GetUser(ctx context.Context, req *v1.GetUserReq) (*v1.GetUserRes, error) {
    user, err := logic.User().GetById(ctx, req.UserId)
    if err != nil {
        return nil, err
    }

    return &v1.GetUserRes{
        User: convert.ToProtoUser(user),
    }, nil
}
```

### 客户端调用
```go
// 创建 gRPC 客户端
conn, err := grpc.Dial(address, grpc.WithInsecure())
if err != nil {
    return err
}
defer conn.Close()

client := v1.NewUserServiceClient(conn)

// 调用服务
resp, err := client.GetUser(ctx, &v1.GetUserReq{
    UserId: userId,
})
```

## 📊 版本管理

### 版本控制
- 使用语义化版本: v1, v2, v3
- 向后兼容的变更保持同一版本
- 不兼容的变更使用新版本

### 版本迁移
```protobuf
// v1/user.proto
message UserV1 {
  string username = 1;
  string email = 2;
}

// v2/user.proto
message UserV2 {
  string username = 1;
  string email = 2;
  string phone = 3;      // 新增字段
}
```

## 🧪 测试

### 单元测试
```go
func TestUserService_GetUser(t *testing.T) {
    // 创建 mock 服务
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockLogic := mock_logic.NewMockIUserLogic(ctrl)
    mockLogic.EXPECT().GetById(gomock.Any(), int64(1)).Return(user, nil)

    // 测试服务调用
    service := &sUserService{logic: mockLogic}
    resp, err := service.GetUser(context.Background(), &v1.GetUserReq{UserId: 1})

    assert.Nil(t, err)
    assert.Equal(t, user.Id, resp.User.Id)
}
```

## 📚 最佳实践

### 设计原则
1. **单一职责**: 每个服务负责一个明确的业务领域
2. **向后兼容**: 新版本必须兼容旧版本
3. **性能考虑**: 合理设计消息大小和调用频率
4. **错误处理**: 明确的错误码和错误信息

### 常见模式
1. **CRUD 服务**: Create, Read, Update, Delete 操作
2. **查询服务**: 复杂的查询和筛选操作
3. **流式服务**: 大数据传输或实时通信
4. **聚合服务**: 组合多个服务的功能

## 🔗 参考资料

- [Protocol Buffers 官方文档](https://developers.google.com/protocol-buffers)
- [gRPC 官方文档](https://grpc.io/docs/)
- [GoFrame Protobuf 指南](https://goframe.org/pages/viewpage.action?pageId=1114369)

---

**最后更新:** 2025-11-17