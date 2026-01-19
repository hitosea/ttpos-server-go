> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# Nginx Grab Staging 多播配置 需求文档

> 本文档定义 Nginx Grab Staging 多播配置的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/nginx-grab-staging-config.md](../../../../team/proposals/2026-01/nginx-grab-staging-config.md) |
| **创建日期**      | 2026-01-04                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint 当前                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/) [x] Nginx [x] Docker                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | ✅ 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2026-01-04             |
| **审核意见** | 技术任务，配置简单明确，风险可控，批准进入设计阶段         |

---

## 📋 概述

新增 `grab_staging.conf` Nginx 配置文件，实现 Grab 外卖模块的请求多播功能。通过 Nginx mirror 模块，将同一请求同时转发到 3 个不同的上游环境（ups1、ups2、upstest），支持多环境对比测试、版本一致性验证和问题排查。

**核心价值**：
- 一次请求同时验证多个环境，提升测试效率
- 通过对比不同环境响应，快速定位问题根源
- 支持新版本上线前的一致性验证，降低发布风险

## 🎯 产品对齐

该功能支持以下产品目标：
1. **提升开发效率**：减少重复测试操作，加快开发迭代
2. **降低发布风险**：通过多环境对比，提前发现版本差异
3. **增强质量保障**：支持并发压力测试，提高测试覆盖度

## 📝 用户故事

**作为** 后端开发工程师  
**我想** 将同一个 API 请求同时转发到 3 个不同的上游环境  
**以便于** 对比测试不同环境的响应差异，快速验证版本一致性和定位问题

---

## 功能需求

### Requirement 1: 定义多个上游服务器

**用户故事**: 作为运维工程师，我想在 Nginx 配置中定义 3 个上游服务器，以便于支持多环境转发

#### 验收标准

1. **WHEN** 配置文件加载 **THEN** 系统 **SHALL** 正确解析 ups1、ups2、upstest 三个 upstream 定义
2. **WHEN** 上游服务器地址变更 **THEN** 系统 **SHALL** 支持通过修改配置文件更新地址
3. **WHEN** Nginx 重载配置 **THEN** 系统 **SHALL** 使用新的上游服务器地址

#### 具体要求

- [x] 1.1 定义 upstream ups1，指向 `https://14031--main--ttpos-go--weifashi.coder.hitosea.com:443`
- [x] 1.2 定义 upstream ups2，指向 `https://14031--main--rikugun--rikugun.coder.hitosea.com:443`
- [x] 1.3 定义 upstream upstest，指向 `https://ttpos-test1.ttpos.com:443`
- [x] 1.4 所有上游使用 HTTPS 协议（端口 443）

---

### Requirement 2: 实现请求多播机制

**用户故事**: 作为开发工程师，我想将请求同时发送到 3 个上游，以便于对比不同环境的响应

#### 验收标准

1. **WHEN** 客户端请求 `/grab_staging/xxx` **THEN** 系统 **SHALL** 同时转发到 ups1、ups2、upstest 三个上游
2. **WHEN** 主请求（upstest）返回响应 **THEN** 系统 **SHALL** 立即返回给客户端，不等待镜像请求完成
3. **WHEN** 镜像请求（ups1、ups2）发送后 **THEN** 系统 **SHALL** 异步处理，不阻塞主请求
4. **WHEN** 任一上游服务器不可用 **THEN** 系统 **SHALL** 不影响其他上游的正常转发

#### 具体要求

- [x] 2.1 使用 Nginx mirror 模块实现请求复制
- [x] 2.2 主请求发送到 upstest，返回其响应给客户端
- [x] 2.3 镜像请求发送到 ups1 和 ups2，异步处理
- [x] 2.4 启用 `mirror_request_body on`，复制请求体
- [x] 2.5 镜像请求的响应被丢弃，不返回给客户端

---

### Requirement 3: 路径重写规则

**用户故事**: 作为开发工程师，我想自动移除 `/grab_staging` 前缀，以便于上游服务器接收正确的路径

#### 验收标准

1. **WHEN** 请求路径为 `/grab_staging/api/order/list` **THEN** 系统 **SHALL** 转发 `/api/order/list` 到上游
2. **WHEN** 请求路径为 `/grab_staging/health` **THEN** 系统 **SHALL** 转发 `/health` 到上游
3. **WHEN** 请求包含查询参数 **THEN** 系统 **SHALL** 保留查询参数并转发

#### 具体要求

- [x] 3.1 使用 `rewrite ^/grab_staging(.*)$ $1 break;` 移除前缀
- [x] 3.2 保留原始请求方法（GET/POST/PUT/DELETE）
- [x] 3.3 保留请求头和请求体
- [x] 3.4 保留查询参数

---

### Requirement 4: HTTPS 代理配置

**用户故事**: 作为运维工程师，我想正确配置 HTTPS 代理，以便于与上游 HTTPS 服务器通信

#### 验收标准

1. **WHEN** 转发到 HTTPS 上游 **THEN** 系统 **SHALL** 使用 TLS 1.2 或 TLS 1.3 协议
2. **WHEN** 建立 SSL 连接 **THEN** 系统 **SHALL** 正确验证服务器证书
3. **WHEN** 设置 Host 头 **THEN** 系统 **SHALL** 使用上游服务器的域名

#### 具体要求

- [x] 4.1 启用 `proxy_ssl_server_name on`
- [x] 4.2 配置 `proxy_ssl_protocols TLSv1.2 TLSv1.3`
- [x] 4.3 设置 `proxy_set_header Host $proxy_host`
- [x] 4.4 设置 `proxy_set_header X-Real-IP $remote_addr`
- [x] 4.5 设置 `proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for`
- [x] 4.6 设置 `proxy_set_header X-Forwarded-Proto $scheme`

---

### Requirement 5: 日志记录

**用户故事**: 作为开发工程师，我想查看所有上游的请求日志，以便于对比不同环境的响应时间和状态码

#### 验收标准

1. **WHEN** 请求发送到 3 个上游 **THEN** 系统 **SHALL** 在访问日志中记录所有 3 个请求
2. **WHEN** 查看日志 **THEN** 系统 **SHALL** 区分主请求和镜像请求
3. **WHEN** 上游返回错误 **THEN** 系统 **SHALL** 记录错误状态码和响应时间

#### 具体要求

- [x] 5.1 使用 Nginx 默认访问日志格式
- [x] 5.2 记录请求时间、响应时间、状态码
- [x] 5.3 记录上游服务器地址
- [x] 5.4 支持通过日志分析工具对比不同上游的性能

---

## 非功能需求

### 配置管理

- **配置文件位置**: `docker/nginx/conf.d/grab_staging.conf`
- **配置格式**: 标准 Nginx 配置语法
- **配置验证**: 使用 `nginx -t` 验证配置正确性
- **配置重载**: 使用 `nginx -s reload` 热重载配置

### 性能要求

- [ ] 镜像请求不影响主请求响应时间
- [ ] Nginx CPU 使用率增加 < 20%
- [ ] Nginx 内存使用率增加 < 50MB
- [ ] 支持 100 QPS 并发请求

### 可靠性要求

- [ ] 任一上游服务器故障不影响其他上游
- [ ] 镜像请求失败不影响主请求响应
- [ ] 配置错误不导致 Nginx 启动失败
- [ ] 支持 Nginx 热重载，不中断服务

### 安全要求

- [ ] 只允许内部网络访问 `/grab_staging` 路径（可选）
- [ ] HTTPS 连接使用 TLS 1.2+ 协议
- [ ] 不在日志中记录敏感信息（如 Token、密码）

---

## 验收标准

### 功能验收

1. **配置文件创建**: `grab_staging.conf` 文件创建完成，语法正确
2. **上游定义**: 3 个 upstream 定义正确，地址可访问
3. **请求多播**: 请求同时转发到 3 个上游，主请求返回响应
4. **路径重写**: `/grab_staging` 前缀正确移除
5. **HTTPS 代理**: HTTPS 连接正常，证书验证通过
6. **日志记录**: 访问日志记录所有 3 个请求

### 测试验收

1. **单元测试**: 配置文件语法验证通过
2. **集成测试**: 端到端请求测试通过
3. **性能测试**: 并发 100 QPS 测试通过
4. **故障测试**: 单个上游故障不影响其他上游

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **配置文档**: 配置示例和说明完整
3. **使用文档**: 使用注意事项和最佳实践完整

---

## 约束条件

### 技术约束

#### Nginx 配置

- 必须使用 Nginx mirror 模块
- 配置文件必须符合 Nginx 语法规范
- 必须支持热重载，不中断服务
- 必须兼容 Docker 容器环境

#### Docker 环境

- 配置文件挂载到容器内 `/etc/nginx/conf.d/`
- 支持通过 `docker exec` 重载配置
- 日志输出到标准输出（stdout）

### 业务约束

- **只读操作优先**: 建议只对查询类 API 使用此功能
- **避免数据写入**: POST/PUT/DELETE 操作会在 3 个环境同时执行，可能导致数据污染
- **按需启用**: 非测试场景建议禁用镜像功能

### 资源约束

- 开发时间: 1 天
- Story Point: 2 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `Nginx 1.18+` - 支持 mirror 模块
- `Docker` - 容器化部署
- `Docker Compose` - 容器编排

### 服务依赖

- **Nginx → ups1**: HTTPS 连接
- **Nginx → ups2**: HTTPS 连接
- **Nginx → upstest**: HTTPS 连接

### 业务依赖

- 3 个上游服务器必须可访问
- 上游服务器必须支持 HTTPS
- 上游服务器必须能够处理额外的镜像请求

---

## 风险和缓解

### 风险 1: 数据污染

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 建议只对查询类 API（GET 请求）使用此功能
- 在配置文件中添加注释，提醒使用者避免写入操作
- 提供按需启用/禁用的机制（注释 mirror 指令）

### 风险 2: 性能影响

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 镜像请求异步处理，不阻塞主请求
- 监控 Nginx CPU 和内存使用情况
- 建议在低峰期或测试环境使用
- 提供性能测试报告

### 风险 3: 日志膨胀

**影响**: 低  
**概率**: 高  
**缓解措施**:

- 配置日志轮转策略（logrotate）
- 定期清理旧日志
- 监控磁盘使用情况
- 可选：使用 ELK 等日志分析工具集中管理

### 风险 4: 上游服务器负载

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 确认上游服务器能够承受额外流量
- 在测试环境先验证配置
- 监控上游服务器资源使用情况
- 提供快速禁用镜像的方法

---

## 时间表

- **Phase 1 - 配置文件创建**: 0.5 天
  - 创建 `grab_staging.conf`
  - 定义 3 个 upstream
  - 配置主 location 和镜像 location
  
- **Phase 2 - 测试验证**: 0.3 天
  - 配置语法验证
  - 功能测试
  - 性能测试
  
- **Phase 3 - 文档编写**: 0.2 天
  - 编写使用文档
  - 编写注意事项
  - 编写故障排查指南
  
- **总计**: 1 天（SP = 2）

---

## 参考资料

### 核心规范

- `docker/nginx/conf.d/ttpos-bmp.conf` - 现有 Nginx 配置参考
- Nginx 官方文档 - mirror 模块
- Nginx 官方文档 - upstream 配置
- Nginx 官方文档 - proxy 模块

### 架构文档

- `docs/human/architecture/infrastructure.md` - 基础设施架构（如有）
- `docker-compose.yml` - Docker 容器配置

### 开发指南

- Nginx 配置最佳实践
- Docker 容器化部署指南

### 外部参考

- [Nginx mirror 模块官方文档](http://nginx.org/en/docs/http/ngx_http_mirror_module.html)
- [Nginx upstream 模块官方文档](http://nginx.org/en/docs/http/ngx_http_upstream_module.html)
- [Nginx proxy 模块官方文档](http://nginx.org/en/docs/http/ngx_http_proxy_module.html)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-04.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-04  
**作者**: rikugun  
**审核者**: [待指定]

