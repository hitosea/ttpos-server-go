# Bug-251226-001 修复任务清单

> **当前状态**: 🟡 规划中
> **开始时间**: 2025-12-26
> **预计完成**: 2025-12-28

---

## 📋 任务列表

### 1. 代码修复

#### 1.1 创建时区转换工具函数

- [ ] **添加 `TimezoneToMySQLOffset` 函数** `main/pkg/utils/time.go`
  - 需求: 将时区名称（如 `Asia/Shanghai`）转换为 MySQL 偏移量（如 `+08:00`）
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 需要处理常见时区和边界情况

#### 1.2 修改 Repository 层方法签名

- [ ] **修改 `CountBusinessSummaryReq` 结构体** `main/app/repository/statistics.go`
  - 需求: 添加 `Timezone string` 字段
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **修改 `CountBusinessPaymentMethodReq` 结构体** `main/app/repository/statistics.go`
  - 需求: 添加 `Timezone string` 字段
  - 预计时间: 0.5小时
  - 负责人: 

#### 1.3 修复 `CountBusinessSummary` 方法

- [ ] **修复 `CountBusinessSummary` SQL 查询** `main/app/repository/statistics.go`
  - 需求: 将 `FROM_UNIXTIME(sb.finish_time, '%s')` 替换为时区转换版本
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 需要修改两处 SQL（totalQuery 和 query）

#### 1.4 修复 `CountBusinessPaymentMethod` 方法

- [ ] **修复 `CountBusinessPaymentMethod` SQL 查询** `main/app/repository/statistics.go`
  - 需求: 将 `FROM_UNIXTIME(po.create_time, '%s')` 替换为时区转换版本
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 需要修改两处 SQL（countQuery 和 dataQuery）

#### 1.5 修复其他使用 `FROM_UNIXTIME` 的方法

- [ ] **修复 `CountSaleDays` 方法** `main/app/repository/statistics.go`
  - 需求: 修复 `FROM_UNIXTIME(complete_time, '%Y-%m-%d')` 的使用
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **修复 `CountPaymentDays` 方法** `main/app/repository/statistics.go`
  - 需求: 修复 `FROM_UNIXTIME(sp.complete_time, '%Y-%m-%d')` 的使用
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **修复 `CountAreaDays` 方法** `main/app/repository/statistics.go`
  - 需求: 修复 `FROM_UNIXTIME(ss.complete_time, '%Y-%m-%d')` 的使用
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **修复 `Count7Days` 方法** `main/app/repository/statistics.go`
  - 需求: 修复 `FROM_UNIXTIME(complete_time, '%Y-%m-%d')` 的使用
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **修复 `CountMemberNumDays` 方法** `main/app/repository/statistics.go`
  - 需求: 修复 `FROM_UNIXTIME(create_time, '%Y-%m-%d')` 的使用
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **修复 `CountMemberDays` 方法** `main/app/repository/statistics.go`
  - 需求: 修复 `FROM_UNIXTIME(complete_time, '%Y-%m-%d')` 的使用
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **修复 `CountMemberPaymentDays` 方法** `main/app/repository/statistics.go`
  - 需求: 修复 `FROM_UNIXTIME(smp.complete_time, '%Y-%m-%d')` 的使用
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **修复 `CountFreePaymentDays` 方法** `main/app/repository/statistics.go`
  - 需求: 修复 `FROM_UNIXTIME(complete_time, '%Y-%m-%d')` 的使用
  - 预计时间: 0.5小时
  - 负责人: 

#### 1.6 修改 Service 层调用

- [ ] **修改 `CountBusinessSummary` Service 方法** `main/app/service/statistics.go`
  - 需求: 在调用 Repository 时传递时区参数
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **修改 `CountBusinessPaymentMethod` Service 方法** `main/app/service/statistics.go`
  - 需求: 在调用 Repository 时传递时区参数
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **修改其他统计方法的 Service 层调用**
  - 需求: 确保所有调用 Repository 的地方都传递时区参数
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 需要检查所有使用上述 Repository 方法的地方

### 2. 测试验证

#### 2.1 单元测试

- [ ] **编写 `TimezoneToMySQLOffset` 单元测试** `main/pkg/utils/time_test.go`
  - 需求: 测试常见时区转换、边界情况、错误处理
  - 预计时间: 1小时
  - 负责人: 

- [ ] **编写 `CountBusinessSummary` 单元测试** `main/app/repository/statistics_test.go`
  - 需求: 测试不同时区下的日期分组正确性
  - 预计时间: 1.5小时
  - 负责人: 

- [ ] **编写 `CountBusinessPaymentMethod` 单元测试** `main/app/repository/statistics_test.go`
  - 需求: 测试不同时区下的日期分组正确性
  - 预计时间: 1.5小时
  - 负责人: 

#### 2.2 集成测试

- [ ] **端到端测试：跨日期边界统计**
  - 需求: 在跨日期边界时间创建订单，验证统计准确性
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 测试场景：23:00-01:00 的订单统计

- [ ] **多时区测试**
  - 需求: 测试不同时区商户的统计准确性
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 测试时区：`Asia/Shanghai`, `Asia/Tokyo`, `Asia/Bangkok`, `Europe/Istanbul`

#### 2.3 手动验证

- [ ] **功能测试：统计报表准确性**
  - 需求: 在测试环境创建测试数据，验证统计报表的日期分组正确性
  - 预计时间: 1小时
  - 负责人: 

- [ ] **回归测试：现有功能不受影响**
  - 需求: 验证现有统计功能正常工作
  - 预计时间: 1小时
  - 负责人: 

### 3. 文档更新

- [ ] **更新故障排查指南** `docs/shared/troubleshooting/mysql-fromunixtime-timezone.md`
  - 需求: 记录问题原因、解决方案、预防措施
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **更新开发文档：时区处理规范**
  - 需求: 在开发文档中明确时区处理规范
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 禁止直接使用 `FROM_UNIXTIME`，必须使用时区转换

### 4. 代码审查与优化

- [ ] **代码审查**
  - 需求: 通过 Code Review，确保代码质量和规范
  - 预计时间: 1小时
  - 负责人: 

- [ ] **性能测试**
  - 需求: 验证时区转换对查询性能的影响
  - 预计时间: 0.5小时
  - 负责人: 

### 5. 部署上线

- [ ] **发布到测试环境**
  - 需求: 部署并验证功能正常
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **发布到生产环境**
  - 需求: 生产发布并监控
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 需要监控统计查询的响应时间和错误率

---

## 📊 任务统计

- **总任务数**: 25
- **已完成**: 0
- **进行中**: 0
- **未开始**: 25
- **完成率**: 0%

### 工作量估算

- **代码修复**: 约 8 小时
- **测试验证**: 约 7 小时
- **文档更新**: 约 1 小时
- **代码审查**: 约 1.5 小时
- **部署上线**: 约 1.5 小时
- **总计**: 约 19 小时（2-3 个工作日）

---

## 🔗 相关链接

- Bug 报告: `bug.md`
- 修复方案: `solution.md`
- 相关文件: `main/app/repository/statistics.go`
- 相关文件: `main/app/service/statistics.go`
- 排查文档: `docs/shared/troubleshooting/mysql-fromunixtime-timezone.md`

---

## 📝 注意事项

1. **时区转换准确性**: 确保时区名称到偏移量的转换准确，特别是处理夏令时的情况
2. **性能影响**: 虽然 `CONVERT_TZ` 在数据库层执行，但仍需关注查询性能
3. **全面检查**: 需要检查所有使用 `FROM_UNIXTIME` 的地方，确保都修复
4. **测试覆盖**: 确保测试覆盖不同时区和跨日期边界场景
5. **向后兼容**: 确保修改后的代码向后兼容，不影响现有功能

