# Nginx Grab Staging 多播配置 任务分解

> 本文档定义 Nginx Grab Staging 多播配置的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 11  
**已完成**: 11  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: 配置文件创建（预计 0.5 天）

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 创建配置文件并定义 upstream

  - File: `docker/nginx/conf.d/grab_staging.conf`
  - Purpose: 创建新配置文件，定义 3 个上游服务器
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: 现有配置: `docker/nginx/conf.d/ttpos-bmp.conf`，设计文档: `design.md`
  - Prompt:
    ```
    Role: DevOps Engineer with Nginx expertise
    
    Task: 创建 grab_staging.conf 配置文件，定义 ups1、ups2、upstest 三个 upstream
    
    Context:
    - File: docker/nginx/conf.d/grab_staging.conf
    - Leverage: docker/nginx/conf.d/ttpos-bmp.conf（参考格式）
    - Requirements: 1.1-1.4（定义多个上游服务器）
    - Design: design.md（完整配置示例）
    
    Configuration:
    - upstream ups1: 14031--main--ttpos-go--weifashi.coder.hitosea.com:443
    - upstream ups2: 14031--main--rikugun--rikugun.coder.hitosea.com:443
    - upstream upstest: ttpos-test1.ttpos.com:443
    
    Restrictions:
    - 使用 HTTPS 协议（端口 443）
    - 配置格式符合 Nginx 语法规范
    
    Success Criteria:
    - 配置文件创建成功
    - 3 个 upstream 定义正确
    - 配置语法验证通过（nginx -t）
    ```

- [x] 1.2 配置主 location 和 mirror 指令

  - File: `docker/nginx/conf.d/grab_staging.conf`
  - Purpose: 配置主请求 location，启用 mirror 模块
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: Task 1.1 的配置文件，设计文档: `design.md`
  - Prompt:
    ```
    Role: Nginx Configuration Expert
    
    Task: 在 grab_staging.conf 中配置主 location /grab_staging，启用 mirror 模块
    
    Context:
    - File: docker/nginx/conf.d/grab_staging.conf
    - Requirements: 2.1-2.5（请求多播机制）
    - Design: design.md（mirror 配置示例）
    
    Configuration:
    - mirror /grab_staging_mirror_ups1（镜像到 ups1）
    - mirror /grab_staging_mirror_ups2（镜像到 ups2）
    - mirror_request_body on（启用请求体复制）
    - rewrite ^/grab_staging(.*)$ $1 break（移除前缀）
    - proxy_pass https://upstest（主请求到 upstest）
    
    Restrictions:
    - 镜像请求异步处理，不阻塞主请求
    - 使用 proxy_http_version 1.1
    - 配置 HTTPS 代理参数
    
    Success Criteria:
    - mirror 指令配置正确
    - rewrite 规则正确
    - proxy_pass 配置正确
    - 配置语法验证通过
    ```

- [x] 1.3 配置镜像 location（ups1）

  - File: `docker/nginx/conf.d/grab_staging.conf`
  - Purpose: 配置内部 location，处理 ups1 的镜像请求
  - Requirements: 2.1, 2.3, 3.1, 3.2, 3.3, 3.4, 4.1-4.6
  - Leverage: Task 1.2 的主 location 配置
  - Prompt:
    ```
    Role: Nginx Configuration Expert
    
    Task: 配置内部 location = /grab_staging_mirror_ups1，处理到 ups1 的镜像请求
    
    Context:
    - File: docker/nginx/conf.d/grab_staging.conf
    - Requirements: 2.1, 2.3, 3.*, 4.*（镜像请求、路径重写、HTTPS 代理）
    - Design: design.md（镜像 location 配置示例）
    
    Configuration:
    - internal（只允许内部访问）
    - rewrite ^/grab_staging(.*)$ $1 break（移除前缀）
    - proxy_pass https://ups1$request_uri（转发到 ups1）
    - 配置 HTTPS 代理参数（proxy_ssl_*, proxy_set_header）
    
    Restrictions:
    - 必须使用 internal 指令
    - 必须配置 TLS 1.2/1.3
    - 必须设置正确的请求头
    
    Success Criteria:
    - internal 指令配置正确
    - rewrite 规则正确
    - proxy_pass 配置正确
    - HTTPS 代理参数完整
    ```

- [x] 1.4 配置镜像 location（ups2）

  - File: `docker/nginx/conf.d/grab_staging.conf`
  - Purpose: 配置内部 location，处理 ups2 的镜像请求
  - Requirements: 2.1, 2.3, 3.1, 3.2, 3.3, 3.4, 4.1-4.6
  - Leverage: Task 1.3 的镜像 location 配置
  - Prompt:
    ```
    Role: Nginx Configuration Expert
    
    Task: 配置内部 location = /grab_staging_mirror_ups2，处理到 ups2 的镜像请求
    
    Context:
    - File: docker/nginx/conf.d/grab_staging.conf
    - Requirements: 2.1, 2.3, 3.*, 4.*（镜像请求、路径重写、HTTPS 代理）
    - Leverage: Task 1.3 的配置（结构相同，只改上游）
    
    Configuration:
    - internal（只允许内部访问）
    - rewrite ^/grab_staging(.*)$ $1 break（移除前缀）
    - proxy_pass https://ups2$request_uri（转发到 ups2）
    - 配置 HTTPS 代理参数（proxy_ssl_*, proxy_set_header）
    
    Restrictions:
    - 必须使用 internal 指令
    - 必须配置 TLS 1.2/1.3
    - 必须设置正确的请求头
    
    Success Criteria:
    - internal 指令配置正确
    - rewrite 规则正确
    - proxy_pass 配置正确
    - HTTPS 代理参数完整
    ```

- [x] 1.5 验证配置语法

  - File: -
  - Purpose: 确保配置文件语法正确
  - Requirements: 所有功能需求
  - Leverage: Task 1.1-1.4 的配置文件
  - Command:
    ```bash
    # 进入 Nginx 容器
    docker exec -it nginx bash
    
    # 验证配置语法
    nginx -t
    
    # 如果验证通过，输出：
    # nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
    # nginx: configuration file /etc/nginx/nginx.conf test is successful
    ```
  - Success: 配置语法验证通过，无错误信息

- [x] 1.6 重载 Nginx 配置

  - File: -
  - Purpose: 使新配置生效
  - Requirements: 所有功能需求
  - Leverage: Task 1.5 验证通过的配置
  - Command:
    ```bash
    # 重载 Nginx 配置（热重载，不中断服务）
    docker exec nginx nginx -s reload
    
    # 检查 Nginx 状态
    docker exec nginx ps aux | grep nginx
    ```
  - Success: 配置重载成功，Nginx 进程正常运行

---

## Phase 2: 测试验证（预计 0.3 天）

- [x] 2.1 功能测试 - GET 请求

  - File: -
  - Purpose: 验证 GET 请求的请求多播功能
  - Requirements: 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 3.4
  - Leverage: Task 1.6 的配置
  - Test Script:
    ```bash
    # 发送测试请求
    curl -X GET "http://localhost/grab_staging/api/health" \
      -H "Authorization: Bearer test_token" \
      -v
    
    # 查看日志（应该看到 3 个请求）
    docker logs nginx 2>&1 | grep "grab_staging" | tail -n 5
    ```
  - Expected:
    - 返回 200 状态码
    - 响应来自 upstest
    - 日志中记录 3 个请求（主请求 + 2 个镜像请求）
    - 响应时间 < 200ms
  - Success: 测试通过，日志显示 3 个请求

- [x] 2.2 功能测试 - 路径重写

  - File: -
  - Purpose: 验证路径重写功能是否正确
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: Task 2.1 的测试脚本
  - Test Script:
    ```bash
    # 测试路径重写
    curl -X GET "http://localhost/grab_staging/api/order/list?page=1" \
      -H "Authorization: Bearer test_token" \
      -v
    
    # 检查日志中的请求路径（应该是 /api/order/list?page=1）
    docker logs nginx 2>&1 | grep "GET /api/order/list" | tail -n 3
    ```
  - Expected:
    - `/grab_staging` 前缀已移除
    - 查询参数保留
    - 上游接收到正确的路径
  - Success: 测试通过，路径重写正确

- [x] 2.3 功能测试 - HTTPS 代理

  - File: -
  - Purpose: 验证 HTTPS 代理配置是否正确
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6
  - Leverage: Task 2.1-2.2 的测试脚本
  - Test Script:
    ```bash
    # 测试 HTTPS 连接
    curl -X GET "http://localhost/grab_staging/api/health" \
      -H "Authorization: Bearer test_token" \
      -v 2>&1 | grep "SSL\|TLS"
    
    # 检查上游连接
    docker exec nginx curl -I https://ttpos-test1.ttpos.com
    docker exec nginx curl -I https://14031--main--ttpos-go--weifashi.coder.hitosea.com
    docker exec nginx curl -I https://14031--main--rikugun--rikugun.coder.hitosea.com
    ```
  - Expected:
    - HTTPS 连接成功
    - TLS 版本为 1.2 或 1.3
    - 证书验证通过
  - Success: 测试通过，HTTPS 代理正常

- [x] 2.4 功能测试 - 日志记录

  - File: -
  - Purpose: 验证日志记录功能
  - Requirements: 5.1, 5.2, 5.3, 5.4
  - Leverage: Task 2.1-2.3 的测试结果
  - Test Script:
    ```bash
    # 发送多个测试请求
    for i in {1..10}; do
      curl -X GET "http://localhost/grab_staging/api/health" \
        -H "Authorization: Bearer test_token" \
        > /dev/null 2>&1
    done
    
    # 查看日志
    docker logs nginx 2>&1 | grep "grab_staging" | tail -n 30
    
    # 统计各上游的请求数
    echo "主请求数:"
    docker logs nginx 2>&1 | grep "GET /grab_staging" | grep -v "mirror" | wc -l
    echo "ups1 镜像请求数:"
    docker logs nginx 2>&1 | grep "grab_staging_mirror_ups1" | wc -l
    echo "ups2 镜像请求数:"
    docker logs nginx 2>&1 | grep "grab_staging_mirror_ups2" | wc -l
    ```
  - Expected:
    - 日志记录所有请求
    - 可以区分主请求和镜像请求
    - 记录请求时间、状态码、响应大小
  - Success: 测试通过，日志记录完整

- [x] 2.5 性能测试 - 并发请求

  - File: -
  - Purpose: 验证系统在并发请求下的性能
  - Requirements: 性能要求
  - Leverage: Task 2.1-2.4 的测试结果
  - Test Script:
    ```bash
    # 使用 wrk 进行压力测试
    wrk -t4 -c100 -d30s -H "Authorization: Bearer test_token" \
      http://localhost/grab_staging/api/health
    
    # 或使用 ab
    ab -n 1000 -c 100 -H "Authorization: Bearer test_token" \
      http://localhost/grab_staging/api/health
    
    # 监控 Nginx 资源使用
    docker stats nginx --no-stream
    ```
  - Expected:
    - 支持 100 QPS 并发请求
    - 平均响应时间 < 200ms
    - Nginx CPU 使用率增加 < 20%
    - Nginx 内存使用率增加 < 50MB
  - Success: 性能测试通过，指标达标

- [x] 2.6 故障测试 - 单个上游故障

  - File: -
  - Purpose: 验证单个上游故障时的容错性
  - Requirements: 2.4（任一上游不可用不影响其他上游）
  - Leverage: Task 2.1-2.5 的测试环境
  - Test Script:
    ```bash
    # 模拟 ups1 故障（修改配置或防火墙规则）
    # 这里假设无法直接停止 ups1，可以修改配置指向不存在的地址
    
    # 发送测试请求
    curl -X GET "http://localhost/grab_staging/api/health" \
      -H "Authorization: Bearer test_token" \
      -v
    
    # 查看日志
    docker logs nginx 2>&1 | grep "grab_staging" | tail -n 10
    ```
  - Expected:
    - 主请求正常返回（upstest）
    - ups2 镜像请求正常
    - ups1 镜像请求失败但不影响主请求
    - 日志中记录 ups1 的错误
  - Success: 测试通过，故障隔离正确

---

## Phase 3: 文档编写（预计 0.2 天）

- [x] 3.1 编写使用文档

  - File: 在 `design.md` 中补充使用说明
  - Purpose: 提供清晰的使用指南
  - Requirements: 所有功能需求
  - Leverage: Task 2.1-2.6 的测试经验
  - Content:
    - 功能介绍
    - 使用场景
    - 请求示例
    - 响应示例
    - 日志查看方法
    - 性能指标
  - Success: 使用文档完整，开发者可以快速上手

- [x] 3.2 编写注意事项

  - File: 在 `design.md` 中补充注意事项
  - Purpose: 提醒使用者避免常见问题
  - Requirements: 风险识别
  - Leverage: requirements.md 中的风险分析
  - Content:
    - ⚠️ 避免数据写入操作
    - ⚠️ 性能影响说明
    - ⚠️ 启用/禁用控制
    - ⚠️ 上游服务器准备
  - Success: 注意事项完整，风险提示清晰

- [x] 3.3 编写故障排查指南

  - File: 在 `design.md` 中补充故障排查章节
  - Purpose: 提供问题排查方法
  - Requirements: 错误处理
  - Leverage: Task 2.6 的故障测试经验
  - Content:
    - 常见问题及解决方法
    - 日志分析方法
    - 配置验证命令
    - 网络连接测试
  - Success: 故障排查指南完整，问题可以快速定位

---

## 提交清单

完成所有任务后，请检查：

### 配置文件质量

- [ ] 所有任务标记为 `[x]`
- [ ] 配置语法验证通过（`nginx -t`）
- [ ] 配置已重载生效
- [ ] 配置文件格式规范（缩进、注释）

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - 配置文件创建完成
  - 上游定义正确
  - 请求多播功能正常
  - 路径重写正确
  - HTTPS 代理正常
  - 日志记录完整

### 测试完成

- [ ] 功能测试通过
  - GET 请求测试通过
  - 路径重写测试通过
  - HTTPS 代理测试通过
  - 日志记录测试通过
- [ ] 性能测试通过
  - 并发 100 QPS 测试通过
  - 响应时间 < 200ms
  - 资源使用在可接受范围内
- [ ] 故障测试通过
  - 单个上游故障不影响主请求
  - 错误日志记录正确

### 文档同步

- [ ] design.md 已补充使用说明
- [ ] design.md 已补充注意事项
- [ ] design.md 已补充故障排查指南
- [ ] requirements.md 验收标准已核对

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/task-infra-grab-staging/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/task-infra-grab-staging/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/task-infra-grab-staging/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/task-infra-grab-staging/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/task-infra-grab-staging/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的设计方案
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成配置
5. **实现配置**: 按照规范实现配置
6. **验证配置**: 运行 `nginx -t` 验证语法
7. **测试功能**: 运行测试脚本验证功能
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Nginx 配置开发

```
Role: DevOps Engineer with Nginx expertise

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: docker/nginx/conf.d/grab_staging.conf
- Leverage config: docker/nginx/conf.d/ttpos-bmp.conf
- Requirements: {需求编号和内容}
- Design: design.md

Restrictions:
- 配置语法符合 Nginx 规范
- 使用 HTTPS 协议
- 使用 TLS 1.2/1.3
- 配置 mirror 模块
- 使用 rewrite 移除前缀
- 配置 proxy_* 指令

Success Criteria:
- {成功标准1}
- 配置语法验证通过（nginx -t）
- 配置重载成功（nginx -s reload）
```

### 测试工程师

```
Role: QA Engineer with Nginx testing expertise

Task: {测试任务描述}

Context:
- Target: Nginx grab_staging 配置
- Test environment: Docker 容器
- Requirements: {需求编号}

Test Cases Required:
- 功能测试
- 性能测试
- 故障测试
- 日志验证

Restrictions:
- 使用 curl / wrk / ab 等工具
- 检查日志输出
- 验证响应时间
- 验证错误处理

Success Criteria:
- 所有测试通过
- 性能指标达标
- 错误处理正确
- 日志记录完整
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-04.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-04  
**维护者**: rikugun

