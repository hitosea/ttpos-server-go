# Nginx Grab Staging 配置需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2026-01-04   |
| **目标版本** | v2.11.0 |
| **状态**   | ✅ 已批准   |
| **关联任务** | - |
| **关联 Spec** | [task-infra-grab-staging](../../shared/specs/active/task-infra-grab-staging/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

在 Grab 外卖模块的开发和测试过程中，需要同时在多个环境（开发环境、测试环境、生产环境）验证同一请求的行为，以便进行对比测试、版本验证和问题排查。

当前 `ttpos-bmp.conf` 中的 `/api/v1/grab` 配置只支持单一的固定上游（`http://ttpos-takeout:14031`），无法满足以下场景：

1. **多环境对比测试**：需要同时观察同一请求在不同环境的响应差异
2. **版本一致性验证**：新版本上线前，需要验证新旧版本的行为一致性
3. **并发压力测试**：需要同时对多个环境进行压力测试
4. **问题定位**：通过对比不同环境的响应，快速定位问题根源

### 业务价值

- 提升测试效率：一次请求同时验证多个环境，无需重复操作
- 降低版本风险：新版本上线前，通过对比测试提前发现差异
- 提高问题排查效率：对比不同环境响应，快速定位问题根源
- 增强测试覆盖：支持多环境并发测试，提高测试置信度

### 目标用户

- [x] 后端开发工程师
- [x] 测试工程师
- [x] 运维工程师
- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客

---

## 💡 解决方案概述

### 方案描述

新增一个 `grab_staging.conf` 配置文件，专门用于 Grab 外卖模块的多环境反向代理配置。

核心设计：
1. **定义多个上游服务器**：配置 3 个上游服务器（ups1、ups2、upstest）
2. **请求多播机制**：使用 Nginx mirror 模块，将请求同时转发到所有 3 个上游
3. **主请求处理**：以 upstest 为主请求，返回其响应给客户端
4. **镜像请求**：同时将请求复制到 ups1 和 ups2，但不等待其响应
5. **路径重写**：移除 `/grab_staging` 前缀，转发到上游服务

### 核心功能点

1. **定义 3 个上游服务器**
   - `ups1`: `https://14031--main--ttpos-go--weifashi.coder.hitosea.com`
   - `ups2`: `https://14031--main--rikugun--rikugun.coder.hitosea.com`
   - `upstest`: `https://ttpos-test1.ttpos.com`

2. **请求多播机制**
   - 主请求发送到 `upstest`，返回其响应
   - 镜像请求同时发送到 `ups1` 和 `ups2`
   - 客户端只接收主请求的响应，不等待镜像请求完成

3. **路径重写规则**
   - 请求路径：`/grab_staging/xxx`
   - 转发路径：`/xxx`（移除 `/grab_staging` 前缀）

4. **应用场景**
   - 多环境对比测试：同一请求在不同环境的行为对比
   - 新版本验证：同时验证新旧版本的响应一致性
   - 压力测试：验证多个环境的稳定性和性能

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端
- [x] 开发/测试环境

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [ ] 数据模型
- [ ] 业务逻辑
- [ ] 第三方集成
- [x] Nginx 反向代理配置
- [x] Docker 容器配置

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯配置调整，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1 天
- **预估 SP**: 2（待技术评审确认）
- **说明**：使用 mirror 模块实现请求多播，配置相对复杂，需要测试验证

### 风险识别

**潜在风险**：
1. **数据污染**：镜像请求可能导致多个环境同时写入数据，造成数据污染
2. **性能影响**：同时转发到 3 个上游可能增加 Nginx 负载
3. **日志膨胀**：3 倍的请求量可能导致日志快速增长
4. **资源消耗**：上游服务器需要处理额外的镜像请求

**缓解措施**：
1. **只读操作**：建议只对查询类 API 使用此功能，避免数据写入
2. **性能监控**：监控 Nginx 和上游服务器的资源使用情况
3. **日志管理**：配置日志轮转和清理策略
4. **按需启用**：仅在需要对比测试时使用，非测试场景建议禁用

---

## 🔗 相关资源

### 参考需求

- 类似功能: `ttpos-bmp.conf` 中的 `/api/v1/grab` 配置
- 参考文档: Nginx 官方文档 - Upstream 配置

### 相关文档

- Nginx 配置文件: `docker/nginx/conf.d/ttpos-bmp.conf`
- Docker Compose 配置: `docker-compose.yml`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | -      |           |
| 技术负责人   | -      |           |
| 开发代表     | rikugun |           |
| 测试代表     | -      |           |
| UI/UX 设计师 | -      |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] 创建 Spec：`task-infra-grab-staging`
- [x] 分配负责人：rikugun
- [ ] 产品审核通过
- [ ] 进入技术设计阶段（`/spec-design`）

---

## 📝 附录

### User Story（初稿）

**作为** 后端开发工程师  
**我想** 将同一个 API 请求同时转发到 3 个不同的上游环境  
**以便于** 对比测试不同环境的响应差异，快速验证版本一致性和定位问题

### AC 验收标准（初稿）

1. **WHEN** 访问 `/grab_staging/xxx` **THEN** 系统 **SHALL** 移除 `/grab_staging` 前缀并同时转发到 ups1、ups2、upstest 三个上游
2. **WHEN** 主请求（upstest）返回响应 **THEN** 系统 **SHALL** 立即返回给客户端，不等待镜像请求完成
3. **WHEN** 镜像请求（ups1、ups2）发送后 **THEN** 系统 **SHALL** 不阻塞主请求，异步处理
4. **WHEN** 查看 Nginx 访问日志 **THEN** 系统 **SHALL** 记录所有 3 个上游的请求日志
5. **WHEN** 任一上游服务器不可用 **THEN** 系统 **SHALL** 不影响其他上游的正常转发

### 配置示例

```nginx
# 定义上游服务器
upstream ups1 {
    server 14031--main--ttpos-go--weifashi.coder.hitosea.com:443;
}

upstream ups2 {
    server 14031--main--rikugun--rikugun.coder.hitosea.com:443;
}

upstream upstest {
    server ttpos-test1.ttpos.com:443;
}

# Grab Staging 主请求（返回响应给客户端）
location /grab_staging {
    # 镜像请求到 ups1
    mirror /grab_staging_mirror_ups1;
    # 镜像请求到 ups2
    mirror /grab_staging_mirror_ups2;
    # 镜像请求异步处理，不等待响应
    mirror_request_body on;
    
    # 移除 /grab_staging 前缀
    rewrite ^/grab_staging(.*)$ $1 break;
    
    # 代理配置（主请求到 upstest）
    proxy_http_version 1.1;
    proxy_ssl_server_name on;
    proxy_ssl_protocols TLSv1.2 TLSv1.3;
    proxy_set_header Host $proxy_host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # 主请求到 upstest
    proxy_pass https://upstest;
}

# 镜像请求到 ups1（内部 location）
location = /grab_staging_mirror_ups1 {
    internal;
    
    # 移除 /grab_staging 前缀
    rewrite ^/grab_staging(.*)$ $1 break;
    
    # 代理配置
    proxy_http_version 1.1;
    proxy_ssl_server_name on;
    proxy_ssl_protocols TLSv1.2 TLSv1.3;
    proxy_set_header Host $proxy_host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # 转发到 ups1
    proxy_pass https://ups1$request_uri;
}

# 镜像请求到 ups2（内部 location）
location = /grab_staging_mirror_ups2 {
    internal;
    
    # 移除 /grab_staging 前缀
    rewrite ^/grab_staging(.*)$ $1 break;
    
    # 代理配置
    proxy_http_version 1.1;
    proxy_ssl_server_name on;
    proxy_ssl_protocols TLSv1.2 TLSv1.3;
    proxy_set_header Host $proxy_host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # 转发到 ups2
    proxy_pass https://ups2$request_uri;
}
```

### 工作原理

1. **主请求流程**：
   - 客户端请求 `/grab_staging/api/xxx`
   - Nginx 移除 `/grab_staging` 前缀
   - 转发到 `upstest` 的 `/api/xxx`
   - 返回 `upstest` 的响应给客户端

2. **镜像请求流程**：
   - 同时复制请求到 `ups1` 和 `ups2`
   - 镜像请求异步处理，不阻塞主请求
   - 镜像请求的响应被丢弃，不返回给客户端

3. **日志记录**：
   - Nginx 访问日志会记录所有 3 个请求
   - 可以通过日志对比不同上游的响应时间和状态码

### 使用注意事项

⚠️ **重要提醒**：

1. **避免数据写入操作**
   - 镜像请求会导致同一操作在 3 个环境同时执行
   - 建议只用于 GET 请求（查询操作）
   - 避免用于 POST/PUT/DELETE（写入操作）

2. **性能影响**
   - 3 倍的请求量会增加 Nginx 和上游服务器负载
   - 建议在低峰期或测试环境使用
   - 监控 Nginx CPU 和内存使用情况

3. **启用/禁用控制**
   - 非测试场景建议注释掉 `mirror` 指令
   - 或者通过 `if` 条件控制是否启用镜像
   - 可以通过请求头标识控制是否触发镜像

4. **上游服务器准备**
   - 确保 ups1 和 ups2 能够承受额外流量
   - 确认上游服务器的防火墙和域名解析正确
   - 建议先在测试环境验证配置

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2026-01-04  
**维护者**: rikugun  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

