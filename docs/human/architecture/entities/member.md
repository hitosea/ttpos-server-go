# Member 实体模型说明

## 基本信息

- **实体名称**: Member
- **表名**: ttpos_member
- **所属模块**: main
- **描述**: 会员信息表，管理会员的基础信息、积分、余额等

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Uuid | uint64 | UUID | 主键标识 |
| MemberNo | string | 会员编号 | NOT NULL |
| Nickname | string | 昵称 | NOT NULL |
| Gender | int | 性别 | 0-女 1-男 2-未知 |
| Phone | string | 电话号码 | NOT NULL |
| IsVisitor | bool | 是否游客 | 0-否 1-是 |
| DeviceId | string | 设备ID | 用于标识游客 |
| Password | string | 密码 | NOT NULL |
| Birthday | int64 | 生日 | 时间戳 |
| Point | float64 | 积分 | NOT NULL |
| FrozenPoint | float64 | 冻结积分 | 不能使用，前端显示为已扣除或已增加 |
| AccumulatedGetPoint | float64 | 累计获取积分 | NOT NULL |
| AccumulatedConsumptionGetPoint | float64 | 累计消费获取积分 | NOT NULL |
| AccumulatedConsumptionAmount | float64 | 累计消费金额 | NOT NULL |
| ConsumptionCount | int | 消费次数 | NOT NULL |
| Balance | float64 | 余额 | NOT NULL |
| FrozenBalance | float64 | 冻结余额 | 不能使用，前端显示为已扣除或已增加 |
| GiftBalance | float64 | 赠送账户余额 | NOT NULL |
| FrozenGiftBalance | float64 | 冻结赠送账户余额 | 不能使用，前端显示为已扣除或已增加 |
| AccumulatedRechargeAmount | float64 | 累计充值金额 | NOT NULL |
| MemberLevelUuid | uint64 | 会员等级ID | NOT NULL |
| MemberCardUuid | uint64 | 会员卡片ID | NOT NULL |
| MemberCardNo | string | 会员卡号 | NOT NULL |
| ReferrerUuid | uint64 | 推荐人ID | NOT NULL |
| ActivityUuid | uint64 | 活动ID | NOT NULL |

## 关联关系

### 关联实体
- **MemberLevel** → 会员等级（通过 MemberLevelUuid 关联）
- **MemberCard** → 会员卡（通过 MemberCardUuid 关联）
- **MemberBalanceLog** → 会员余额变动记录（通过 MemberUuid 关联）
- **MarketingActivity** → 营销活动（通过 ActivityUuid 关联）

## 业务方法

### 积分管理
- **GetPoints()**: 获取会员积分余额（积分+冻结积分）
- **ChangePoint(points)**: 变动积分（正数增加，负数扣减）
- **UpdatePoint(changePoints, changePointsConsumption)**: 更新积分并清零冻结积分

### 余额管理
- **GetBalance()**: 获取会员余额（余额+冻结余额）
- **GetGiftBalance()**: 获取赠送余额（赠送余额+冻结赠送余额）
- **GetBalanceAll()**: 获取总余额（余额+赠送余额）
- **SetFrozenBalance(balanceAmount, deductRatioMain, deductRatioGift)**: 设置冻结余额
- **UpdateBalance(changeBalance, changeBalanceGift)**: 更新余额并清零冻结余额

### 折扣管理
- **GetMemberDiscountRate()**: 获取会员等级折扣率
- **GetMemberCardDiscountRate()**: 获取会员卡折扣率

### 其他方法
- **IsExistActivityAndReferrer()**: 是否存在推荐人和活动ID
- **GetMemberCardName()**: 获取会员卡名称
- **GetMemberLevelName()**: 获取会员等级名称
- **HasPassword()**: 是否有密码
- **GetRechargeMoney()**: 获取累计充值金额

## 数据流分析

### 数据来源
- 会员注册信息
- 会员消费记录
- 会员充值记录
- 会员积分变动记录

### 数据流向
1. **会员注册流程**:
   - 创建会员基础信息
   - 分配会员编号和UUID
   - 设置初始会员等级

2. **消费流程**:
   - 订单支付时使用会员余额
   - 冻结相应余额
   - 订单完成后更新余额和积分
   - 记录消费历史

3. **充值流程**:
   - 创建充值订单
   - 支付完成后增加余额
   - 可能赠送积分和余额
   - 记录充值历史

4. **积分管理流程**:
   - 消费获得积分
   - 使用积分抵扣
   - 积分过期处理
   - 积分变动记录

### 业务场景
- 会员注册和管理
- 会员等级体系
- 会员积分系统
- 会员余额管理
- 会员卡系统
- 推荐人系统
- 营销活动参与

## 索引建议

- 主键索引: Uuid
- 唯一索引: MemberNo（会员编号唯一）
- 唯一索引: Phone（手机号唯一）
- 普通索引: MemberLevelUuid（等级查询）
- 普通索引: MemberCardUuid（会员卡查询）
- 普通索引: ReferrerUuid（推荐人查询）
- 普通索引: ActivityUuid（活动查询）

## 注意事项

1. 使用冻结机制处理并发操作
2. 积分和余额都支持主账户和赠送账户分离
3. 支持游客模式（IsVisitor）
4. 会员等级和会员卡都可以提供折扣
5. 所有金额字段使用decimal类型保证精度
6. 支持推荐人系统和营销活动关联