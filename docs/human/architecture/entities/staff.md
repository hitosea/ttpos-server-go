# Staff 实体模型说明

## 基本信息

- **实体名称**: Staff
- **表名**: ttpos_staff
- **所属模块**: model
- **描述**: 员工管理实体，用于管理系统员工信息，包含登录认证、角色权限、班次管理等完整功能

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| Uuid | uint64 | UUID | 继承自BaseModel |
| CreateTime | int64 | 创建时间 | 继承自BaseModel |
| UpdateTime | int64 | 更新时间 | 继承自BaseModel |
| DeleteTime | int64 | 删除时间 | 继承自BaseModel |
| CompanyUuid | uint64 | 集团ID | 多租户标识 |
| Username | string | 用户名 | 登录用户名 |
| Password | string | 登录密码 | 加密存储 |
| Phone | string | 手机号 | 联系方式 |
| PasswordChangeCount | int | 修改密码次数 | 密码安全统计 |
| PasswordChangeTime | int64 | 修改密码时间 | 时间戳 |
| RealName | string | 姓名 | 员工真实姓名 |
| IsSuper | int | 是否超级管理员 | 0-否 1-是 |
| UserType | int | 账号类型 | 0-总台 1-门店 |
| IsDisable | int | 是否禁用 | 1-禁用 0-未禁用 |
| BindKey | string | 绑定的设备key | 设备绑定 |
| CashierOnline | int | 收银员当班状态 | 0-不在线 1-在线 |
| CashierLoginTime | int64 | 收银员当班登录时间 | 时间戳 |
| DutyNo | string | 当班编号 | 班次标识 |

## 关联关系

### 关联实体
- **CompanyUuid** → Company 实体（通过 CompanyUuid 关联）
- **BindKey** → Device 实体（通过 BindKey 关联）

### 关联的子实体
- **StaffRole** → 员工角色关系（一对多关系）
- **StaffShiftLog** → 员工交班记录（一对多关系）
- **StaffLoginLog** → 登录日志（一对多关系）
- **StaffOperationLog** → 操作日志（一对多关系）

### 多对多关系
- **Role** → 角色（通过 staff_role 关联表）

### 被关联关系
- **H5Order** → 通过 StaffUuid 关联（接单拒单操作）
- **SaleOrder** → 通过 OperatorUuid 关联
- **SaleOrderOperationRecord** → 通过 OperatorUuid 关联

## 数据流分析

### 数据来源
- 管理员创建员工账号
- 员工自助注册（如适用）
- 第三方系统同步
- 批量导入员工数据

### 数据流向
1. **员工创建流程**:
   - 管理员录入员工基本信息
   - 设置登录用户名和初始密码
   - 分配相应的角色权限
   - 创建员工账号记录

2. **登录认证流程**:
   - 员工输入用户名密码
   - 系统验证身份信息
   - 记录登录日志和IP地址
   - 更新最后登录时间

3. **角色权限管理**:
   - 为员工分配一个或多个角色
   - 根据角色定义赋予相应权限
   - 支持动态权限调整
   - 记录权限变更历史

4. **班次管理流程**:
   - 员工上班时登录收银系统
   - 系统记录当班开始时间
   - 工作期间记录各项操作
   - 下班时进行交班结算

### 业务场景
- 员工账号管理
- 登录认证和授权
- 角色权限控制
- 班次和交接班管理
- 操作日志审计
- 设备绑定管理
- 密码安全管理
- 员工状态管理

## 索引建议

- 主键索引: ID
- 唯一索引: Uuid
- 唯一索引: Username（用户名唯一）
- 普通索引: CompanyUuid（公司查询）
- 普通索引: Phone（手机号查询）
- 普通索引: IsDisable（状态筛选）
- 普通索引: UserType（类型查询）
- 普通索引: IsSuper（超级管理员查询）
- 普通索引: CashierOnline（当班状态查询）

## 业务方法

### GetUserName() string
- **功能**: 获取用户名
- **返回**: 用户名
- **逻辑**: 优先返回RealName，如果为空则返回Username
- **用途**: 统一用户名获取逻辑

## 扩展实体

### StaffRole 员工角色关系表
#### 基本信息
- **表名**: ttpos_staff_role
- **描述**: 管理员工与角色的多对多关系

#### 字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| ID | uint | 主键ID |
| StaffUuid | int64 | 员工UUID |
| RoleUuid | int64 | 角色UUID |

### StaffShiftLog 员工交班记录表
#### 基本信息
- **表名**: ttpos_staff_shift_log
- **描述**: 记录员工班次交接的完整信息

#### 字段说明
| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| ID | uint | 主键ID | 继承自BaseModel |
| StaffUuid | uint64 | 员工ID | |
| ShiftNo | string | 交班编号 | |
| Status | int | 状态 | 0-未交班 1-已交班 |
| PreviousShiftCash | float64 | 上一班遗留备用金 | |
| CurrentCashTotal | float64 | 当前钱箱现金总计 | |
| Incomes | string | 收入详情 | JSON格式 |
| TotalIncome | float64 | 总收入 | |
| CashTakenOut | float64 | 本班取出现金 | |
| CashLeft | float64 | 本班遗留备用金 | |
| CashIncome | float64 | 本班收入现金 | |
| TotalBusiness | float64 | 本班营业总额 | |
| IsPrinted | int | 是否打印 | 0-未打印 1-已打印 |
| Remark | string | 备注 | |
| WithdrawCash | float64 | 中途取出现金 | |
| DepositCash | float64 | 中途存入现金 | |
| ExceptionRemark | string | 异常报备 | |
| Abnormal | string | 异常信息 | JSON字符串 |
| ShiftStartTime | int64 | 当班开始时间 | |
| ShiftEndTime | int64 | 当班结束时间 | |
| ErpnextOpenPosEntryName | string | erpnext开账名称 | |
| ErpnextClosePosEntryName | string | erpnext结账名称 | |
| ErpnextAsyncRecordId | string | erpnext异步记录ID | |

#### 业务方法

##### IsHandedOver() bool
- **功能**: 判断是否已交班
- **返回**: 是否已交班
- **用途**: 交班状态检查

##### IsReported() bool
- **功能**: 判断是否已经报备
- **返回**: 是否已报备
- **用途**: 异常情况检查

### StaffLoginLog 管理员登录记录表
#### 基本信息
- **表名**: ttpos_staff_login_log
- **描述**: 记录员工登录历史和安全信息

#### 字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| ID | uint | 主键ID |
| StaffUuid | uint64 | 员工UUID |
| Username | string | 用户名 |
| Ip | string | 登录IP |
| Result | string | 登录结果 |

### StaffOperationLog 员工操作日志表
#### 基本信息
- **表名**: ttpos_staff_operation_log
- **描述**: 详细记录员工系统操作行为

#### 字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| ID | uint | 主键ID |
| StaffUuid | uint64 | 员工ID |
| Title | string | 标题 |
| Url | string | 操作URL |
| RequestData | string | 请求数据 |
| ResponseData | string | 响应数据 |
| Type | string | 操作类型 |
| Ip | string | 操作IP |
| Source | string | 操作来源 |
| Agent | string | 操作用户代理 |

### StaffShiftSnapshot 员工交班快照表
#### 基本信息
- **表名**: ttpos_staff_shift_snapshot
- **描述**: 存储交班时的详细数据快照

#### 字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| ID | uint | 主键ID |
| ShiftLogUuid | uint64 | 交班记录ID |
| Content | string | 快照JSON |

## 交班快照详细结构

### StaffShiftSnapshotContent 交班快照内容
包含完整的班次数据统计，包括：
- 基础信息（员工ID、交班编号、状态等）
- 现金管理（收入、支出、余额等）
- 异常统计（退款、改价、赠菜等）
- 订单统计（销售额、订单量、客单价等）
- 高峰期分析
- 税率统计
- 支付方式统计
- 员工信息快照

### 业务规则

1. **账号管理规则**:
   - 用户名在系统内必须唯一
   - 超级管理员拥有所有权限
   - 禁用账号无法登录系统
   - 密码修改需要记录次数和时间

2. **登录安全规则**:
   - 记录每次登录的IP地址
   - 记录登录成功或失败结果
   - 支持密码修改次数限制
   - 设备绑定增强安全性

3. **角色权限规则**:
   - 员工可以拥有多个角色
   - 角色权限可以动态调整
   - 权限变更立即生效
   - 记录权限变更历史

4. **班次管理规则**:
   - 每次当班都有唯一当班编号
   - 交班时必须进行现金对账
   - 支持异常情况报备
   - 快照数据不可篡改

5. **操作审计规则**:
   - 记录所有重要操作行为
   - 保存操作请求和响应数据
   - 记录操作来源和IP信息
   - 支持操作历史查询

## 注意事项

1. **数据安全**: 密码需要加密存储，定期要求修改
2. **权限控制**: 严格按照角色权限控制操作范围
3. **日志记录**: 重要操作必须记录日志，支持审计
4. **设备管理**: 支持员工账号与设备绑定
5. **多租户**: 通过CompanyUuid实现数据隔离
6. **状态管理**: 员工状态变更需要及时同步
7. **班次连续性**: 确保班次交接数据的完整性
8. **异常处理**: 异常情况需要及时报备和处理

## 扩展功能

### 智能权限管理
- 基于操作频率的权限推荐
- 权限使用情况分析
- 异常权限访问告警

### 班次优化
- 基于历史数据的班次安排
- 工作量分析和优化建议
- 交接效率提升

### 安全管理
- 登录异常检测
- 密码强度检查
- 设备访问控制

### 绩效分析
- 员工操作效率统计
- 销售业绩分析
- 工作质量评估

## 性能优化

### 查询优化
- 角色权限关联查询优化
- 登录日志分页查询
- 操作日志按时间索引

### 数据归档
- 历史登录日志定期归档
- 操作日志按策略清理
- 交班快照数据压缩

## 统计分析

### 登录分析
- **登录频率统计**: 按员工、时间统计登录次数
- **登录地点分析**: IP地址分布和异常登录检测
- **登录成功率**: 登录失败原因分析和安全评估

### 权限分析
- **角色使用统计**: 各角色使用频率和权限利用率
- **权限变更分析**: 权限调整趋势和影响评估
- **权限合规检查**: 权限分配的合规性审计

### 班次分析
- **班次效率分析**: 员工工作效率和业绩统计
- **现金管理分析**: 现金收支准确性和异常统计
- **交接质量分析**: 交接班完整性和问题统计

### 操作分析
- **操作频率统计**: 员工操作类型和频率分析
- **异常操作监控**: 异常操作行为识别和告警
- **操作效率评估**: 操作流程优化建议