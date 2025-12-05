# 总部-分店颗粒化同步（后端）任务分解

> 本文档将功能拆分为可执行的开发任务。

**关联文档**:
- 设计文档: `design.md`
- 关联任务: DooTask #37462

---

## 📋 任务总览

| Phase | 任务数 | 预估工时 | 状态 |
|-------|--------|----------|------|
| Phase 1: 数据库和模型 | 5 | 2.5h | ⏳ 待开始 |
| Phase 2: 常量定义 | 1 | 0.5h | ⏳ 待开始 |
| Phase 3: Repository 扩展 | 6 | 2h | ⏳ 待开始 |
| Phase 4: Service 层实现 | 12 | 9.5h | ⏳ 待开始 |
| Phase 5: API 层实现 | 3 | 2h | ⏳ 待开始 |
| Phase 6: 测试 | 6 | 5.5h | ⏳ 待开始 |
| Phase 7: 部署和联调 | 3 | 2h | ⏳ 待开始 |
| **总计** | **36** | **24h** | - |

---

## Phase 1: 数据库和模型

### Task 1.1: 创建数据库迁移文件

**目标**: 为5张表添加 `headquarter_uuid` 字段

**步骤**:

1. 创建迁移PHP文件：`admin/database/migrations/20251205175111_add_headquarter_uuid_to_sync_tables.php`
   
   ```php
   <?php
   
   use think\migration\Migrator;
   use think\migration\db\Column;
   
   class AddHeadquarterUuidToSyncTables extends Migrator
   {
       public function change()
       {
           // 1. 优惠券表
           $this->addHeadquarterUuidColumn('marketing_coupon');
           
           // 2. 满额减活动表
           $this->addHeadquarterUuidColumn('full_reduction_activity');
           
           // 3. 菜品标签表
           $this->addHeadquarterUuidColumn('product_label');
           
           // 4. 营销活动表
           $this->addHeadquarterUuidColumn('marketing_activity');
           
           // 5. 支付方式表
           $this->addHeadquarterUuidColumn('payment_method');
       }
       
       private function addHeadquarterUuidColumn($tableName)
       {
           $table = $this->table($tableName);
           
           if (!$table->hasColumn('headquarter_uuid')) {
               $table->addColumn(
                   'headquarter_uuid',
                   'biginteger',
                   [
                       'limit' => 20,
                       'signed' => false,
                       'null' => false,
                       'default' => 0,
                       'comment' => '总部uuid，0表示本店创建，>0表示从总部同步',
                       'after' => 'uuid'
                   ]
               )->update();
           }
           
           // 添加索引
           try {
               $table->addIndex('headquarter_uuid', ['name' => 'idx_headquarter_uuid'])->update();
           } catch (\Exception $e) {
               // 索引已存在，忽略
           }
       }
   }
   ```
   
   **说明**：
   - ✅ 使用ThinkPHP迁移系统
   - ✅ 支持幂等性（可重复执行）
   - ✅ 自动检查字段和索引是否存在
   - ✅ 所有表的字段都添加在 `uuid` 字段之后

**验收标准**:
- [ ] PHP迁移文件已创建
- [ ] 代码符合ThinkPHP迁移规范
- [ ] 包含所有5张表的字段添加逻辑
- [ ] 支持幂等性（可重复执行）

**预估工时**: 0.5h

---

### Task 1.2: 确认支付方式表结构和业务规则

**目标**: 确认支付方式表的实际名称、位置、结构和同步业务规则

**步骤**:

1. 查找支付方式相关代码（可能在旧管理端PHP或新管理端Go）
   ```bash
   # 在旧管理端PHP搜索
   grep -r "payment.*method\|payment_name" admin/app/
   
   # 在新管理端Go搜索
   grep -r "PaymentMethod\|payment_name" main/app/model/
   
   # 查找数据库表结构
   grep -r "ttpos_payment" admin/database/migrations/
   grep -r "ttpos_payment" admin/database/seeds/
   ```

2. 确认表名和关键字段：
   - `id`, `payment_name`, `code`, `source`（1=手动添加）
   - `logo_file_uuid`, `qrcode_file_uuid`
   - `fee_percent`, `status`, `sort`
   - `is_show_cashier`, `is_show_assistant`, `is_show_member_recharge`
   - `default_img`, `erpnext_payment`

3. **确认 code 生成规则**：
   - 手动添加（`source = 1`）时的 code 生成逻辑
   - 通常是 90000-99999 范围自增
   - 查找现有代码：`grep -r "source.*1\|generatePaymentCode" admin/app/`

4. **确认特殊 code 的含义**：
   - code=40 和 code=10：不同步（系统保留）
   - code=90111, 90222, 90333：特殊处理（只更新 headquarter_uuid）

**验收标准**:
- [ ] 确认支付方式表名和所有字段
- [ ] 确认 code 生成规则（手动添加时）
- [ ] 确认特殊 code 的业务含义
- [ ] 确认是在PHP还是Go中管理
- [ ] 记录所有默认值

**预估工时**: 1h

---

### Task 1.3: 执行数据库迁移

**目标**: 在主库和所有分店库执行迁移

**步骤**:

1. 在开发环境执行迁移
   ```bash
   cd admin
   php think migrate:run
   ```

2. 验证迁移结果
   ```bash
   # 查看迁移状态
   php think migrate:status
   
   # 连接数据库验证字段
   docker exec -it saas-mysql mysql -uroot -p{password} -e "USE ttpos_db; DESC ttpos_marketing_coupon;"
   ```

3. 检查字段是否已添加
   ```sql
   -- 应该能看到 headquarter_uuid 字段
   SHOW COLUMNS FROM ttpos_marketing_coupon LIKE 'headquarter_uuid';
   SHOW COLUMNS FROM ttpos_full_reduction_activity LIKE 'headquarter_uuid';
   SHOW COLUMNS FROM ttpos_product_label LIKE 'headquarter_uuid';
   SHOW COLUMNS FROM ttpos_marketing_activity LIKE 'headquarter_uuid';
   SHOW COLUMNS FROM ttpos_payment_method LIKE 'headquarter_uuid';
   ```

**验收标准**:
- [ ] 开发环境迁移成功
- [ ] 所有表都有 `headquarter_uuid` 字段
- [ ] 索引已创建
- [ ] 迁移脚本已记录（用于生产环境）

**预估工时**: 0.5h

---

### Task 1.4: 更新 Go Model

**目标**: 为相关 Model 添加 `HeadquarterUuid` 字段

**步骤**:

1. 更新 `main/app/model/marketing.go` (MarketingCoupon)
   ```go
   type MarketingCoupon struct {
       BaseModel
       HeadquarterUuid uint64  `gorm:"column:headquarter_uuid;default:0" json:"headquarter_uuid"` // 新增
       Name            string  `gorm:"column:name;type:varchar(50)" json:"name"`
       // ... 其他字段
   }
   ```

2. 更新 `main/app/model/full_reduction_activity.go`
   ```go
   type FullReductionActivity struct {
       BaseModel
       HeadquarterUuid       uint64 `gorm:"column:headquarter_uuid;default:0" json:"headquarter_uuid"` // 新增
       Name                  string `gorm:"column:name;type:varchar(1000);default:''" json:"name"`
       // ... 其他字段
   }
   ```

3. 更新 `main/app/model/product_label.go`
   ```go
   type ProductLabel struct {
       BaseModel
       HeadquarterUuid uint64 `gorm:"column:headquarter_uuid;default:0" json:"headquarter_uuid"` // 新增
       Name            string `gorm:"default:'';column:name" json:"name"`
       // ... 其他字段
   }
   ```

4. 更新 `main/app/model/marketing.go` (MarketingActivity)
   ```go
   type MarketingActivity struct {
       BaseModel
       HeadquarterUuid       uint64 `gorm:"column:headquarter_uuid;default:0" json:"headquarter_uuid"` // 新增
       Name                  string `gorm:"column:name;type:varchar(2500);default:''" json:"name"`
       // ... 其他字段
   }
   ```

5. 创建或更新 `main/app/model/payment_method.go`（如果不存在）
   ```go
   type PaymentMethod struct {
       BaseModel
       HeadquarterUuid uint64 `gorm:"column:headquarter_uuid;default:0" json:"headquarter_uuid"`
       Name            string `gorm:"column:name;type:varchar(100)" json:"name"`
       Code            string `gorm:"column:code;type:varchar(50)" json:"code"`
       // ... 其他字段（根据实际表结构）
   }
   
   func (*PaymentMethod) TableName() string {
       return "ttpos_payment_method"
   }
   ```

**验收标准**:
- [ ] 所有5个Model都添加了 `HeadquarterUuid` 字段
- [ ] 字段标签正确（gorm, json）
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 1.5: 定义 DTO

**目标**: 创建请求和响应 DTO

**步骤**:

1. 创建 `main/app/dto/req/sync_req.go` 或在现有文件中添加
   ```go
   // GetHeadquartersDataListReq 获取总部可同步数据列表请求
   type GetHeadquartersDataListReq struct {
       DataTypes []string `json:"data_types"` // 可选，指定查询的数据类型
   }
   
   // GranularSyncReq 颗粒化同步请求
   type GranularSyncReq struct {
       SyncData GranularSyncData `json:"sync_data" binding:"required"`
   }
   
   // GranularSyncData 要同步的数据（按种类分组）
   type GranularSyncData struct {
       ProductCategory   []uint64 `json:"product_category"`
       Unit              []uint64 `json:"unit"`
       Flavor            []uint64 `json:"flavor"`
       Attribute         []uint64 `json:"attribute"`
       Sauce             []uint64 `json:"sauce"`
       Product           []uint64 `json:"product"`
       MaterialCategory  []uint64 `json:"material_category"`
       Material          []uint64 `json:"material"`
       BomCard           []uint64 `json:"bom_card"`
       Supplier          []uint64 `json:"supplier"`
       Tax               []uint64 `json:"tax"`
       Coupon            []uint64 `json:"coupon"`
       FullReduction     []uint64 `json:"full_reduction"`
       ProductLabel      []uint64 `json:"product_label"`
       MarketingActivity []uint64 `json:"marketing_activity"`
       PaymentMethod     []uint64 `json:"payment_method"`
   }
   ```

2. 创建 `main/app/dto/resp/sync_resp.go` 或在现有文件中添加
   ```go
   // HeadquartersDataListResp 总部可同步数据列表响应
   type HeadquartersDataListResp struct {
       DataGroups []DataGroup `json:"data_groups"`
   }
   
   // DataGroup 数据分组
   type DataGroup struct {
       Type        string     `json:"type"`
       TypeName    string     `json:"type_name"`
       Items       []DataItem `json:"items"`
       SyncedUuids []uint64   `json:"synced_uuids"` // 分店已同步的总部数据uuid列表
   }
   
   // DataItem 数据项
   type DataItem struct {
       Uuid           uint64         `json:"uuid"`
       Name           string         `json:"name"`
       RelatedData    []RelatedData  `json:"related_data,omitempty"` // 关联数据（明确类型）
       AdditionalInfo map[string]any `json:"additional_info,omitempty"`
   }
   
   // RelatedData 关联数据
   type RelatedData struct {
       Type  string   `json:"type"`  // 关联数据的类型（如：product, category等）
       Uuids []uint64 `json:"uuids"` // 关联的uuid列表
   }
   
   // GranularSyncResp 颗粒化同步响应
   type GranularSyncResp struct {
       TaskUuid uint64 `json:"task_uuid"`
       Message  string `json:"message"`
   }
   ```

**验收标准**:
- [ ] Request DTO 已创建
- [ ] Response DTO 已创建
- [ ] 字段标签正确（json, binding）
- [ ] 编译通过

**预估工时**: 0.5h

---

## Phase 2: 常量定义

### Task 2.1: 定义新的同步数据类型常量

**目标**: 在 `constant` 包中定义新的数据类型常量

**步骤**:

1. 在 `main/app/constant/sync.go` (如果没有则创建) 中添加常量
   ```go
   const (
       // 现有常量
       SyncTaskTypeProductCategory  = "product_category"
       SyncTaskTypeUnit             = "unit"
       // ... 其他现有常量
       
       // 新增常量
       SyncDataTypeCoupon           = "coupon"
       SyncDataTypeFullReduction    = "full_reduction"
       SyncDataTypeProductLabel     = "product_label"
       SyncDataTypeMarketingActivity = "marketing_activity"
       SyncDataTypePaymentMethod    = "payment_method"
   )
   
   var SyncDataTypeNames = map[string]string{
       // 现有映射
       SyncTaskTypeProductCategory: "商品分类",
       SyncTaskTypeUnit:            "单位",
       // ... 其他现有映射
       
       // 新增映射
       SyncDataTypeCoupon:           "优惠券",
       SyncDataTypeFullReduction:    "满额减",
       SyncDataTypeProductLabel:     "菜品标签",
       SyncDataTypeMarketingActivity: "营销活动",
       SyncDataTypePaymentMethod:    "支付方式",
   }
   ```

**验收标准**:
- [ ] 常量已定义
- [ ] 常量名称已添加到映射表
- [ ] 编译通过

**预估工时**: 0.5h

---

## Phase 3: Repository 扩展

### Task 3.1: 扩展 Repository 查询选项（优惠券）

**目标**: 为优惠券 Repository 添加按 uuid 列表查询的选项

**步骤**:

1. 检查是否已有 `MarketingCouponRepo`，如果没有则创建
   ```bash
   ls main/app/repository/marketing_coupon_repo.go
   ```

2. 如果不存在，创建 `main/app/repository/i_marketing_coupon_repo.go`
   ```go
   type IMarketingCouponRepo interface {
       Create(coupon *model.MarketingCoupon) error
       GetByUuid(uuid uint64, options ...DBOption) (*model.MarketingCoupon, error)
       GetList(options ...DBOption) ([]*model.MarketingCoupon, error)
       
       // 选项方法
       WhereUuids(uuids []uint64) DBOption
       WhereHeadquarterUuid(headquarterUuid uint64) DBOption
   }
   ```

3. 创建 `main/app/repository/marketing_coupon_repo.go`
   ```go
   type MarketingCouponRepoImpl struct {
       db *gorm.DB
   }
   
   func NewMarketingCouponRepo(db *gorm.DB) IMarketingCouponRepo {
       return &MarketingCouponRepoImpl{db: db}
   }
   
   func (r *MarketingCouponRepoImpl) WhereUuids(uuids []uint64) DBOption {
       return func(db *gorm.DB) *gorm.DB {
           if len(uuids) > 0 {
               return db.Where("uuid IN (?)", uuids)
           }
           return db
       }
   }
   
   func (r *MarketingCouponRepoImpl) WhereHeadquarterUuid(headquarterUuid uint64) DBOption {
       return func(db *gorm.DB) *gorm.DB {
           return db.Where("headquarter_uuid = ?", headquarterUuid)
       }
   }
   
   // ... 其他方法
   ```

**验收标准**:
- [ ] Repository 接口已定义
- [ ] Repository 实现已创建
- [ ] 选项方法已实现
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 3.2: 扩展 Repository 查询选项（满额减）

**目标**: 为满额减 Repository 添加按 uuid 列表查询的选项

**步骤**: 参考 Task 3.1，为 `FullReductionActivity` 创建 Repository

**验收标准**:
- [ ] Repository 接口已定义
- [ ] Repository 实现已创建
- [ ] 选项方法已实现
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 3.3: 扩展 Repository 查询选项（菜品标签）

**目标**: 为菜品标签 Repository 添加按 uuid 列表查询的选项

**步骤**:

1. 检查是否已有 `ProductLabelRepo`
2. 如果存在，扩展查询选项；如果不存在，参考 Task 3.1 创建
3. 特别注意：需要 `Preload("ProductPackages")` 加载关联商品

**验收标准**:
- [ ] Repository 接口已定义
- [ ] Repository 实现已创建
- [ ] 支持 Preload 关联商品
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 3.4: 扩展 Repository 查询选项（营销活动）

**目标**: 为营销活动 Repository 添加按 uuid 列表查询的选项

**步骤**: 参考 Task 3.1，为 `MarketingActivity` 创建 Repository

**验收标准**:
- [ ] Repository 接口已定义
- [ ] Repository 实现已创建
- [ ] 选项方法已实现
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 3.5: 扩展 Repository 查询选项（支付方式）

**目标**: 为支付方式 Repository 添加按 uuid 列表查询的选项

**步骤**: 参考 Task 3.1，为 `PaymentMethod` 创建 Repository

**验收标准**:
- [ ] Repository 接口已定义
- [ ] Repository 实现已创建
- [ ] 选项方法已实现
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 3.6: 扩展现有 Service 的 Sync 方法

**目标**: 为现有 Service 添加按 uuid 列表同步的方法

**步骤**:

1. 检查 `ProductSrv` 的同步方法是否支持按 uuid 列表过滤
   ```go
   // main/app/service/i_product_srv.go
   type IProductSrv interface {
       // 现有方法
       SyncProduct(ctx context.Context) error
       
       // 新增方法
       SyncProductByUuids(ctx context.Context, uuids []uint64) error
       SyncProductCategoryByUuids(ctx context.Context, uuids []uint64) error
       // ... 其他
   }
   ```

2. 实现新方法（在 `main/app/service/product.go` 中）
   ```go
   func (s *productSrv) SyncProductByUuids(ctx context.Context, uuids []uint64) error {
       if len(uuids) == 0 {
           return nil
       }
       
       // 复用现有 SyncProduct 逻辑，但添加 uuid 过滤
       // ... 实现代码
   }
   ```

3. 同样为 `MaterialSrv`, `SupplierSrv` 等添加按 uuid 同步的方法

**验收标准**:
- [ ] ProductSrv 的同步方法支持 uuid 过滤
- [ ] MaterialSrv 的同步方法支持 uuid 过滤
- [ ] SupplierSrv 的同步方法支持 uuid 过滤
- [ ] 编译通过

**预估工时**: 1h

---

## Phase 4: Service 层实现

### Task 4.1: 实现 GetHeadquartersDataList 方法

**目标**: 实现获取总部可同步数据列表的核心逻辑

**步骤**:

1. 在 `main/app/service/i_sync_srv.go` 中添加接口
   ```go
   type ISyncSrv interface {
       // 现有方法
       Sync(ctx context.Context, syncReq req.SyncReq) (resp.SyncResp, error)
       // ...
       
       // 新增方法
       GetHeadquartersDataList(ctx context.Context, req req.GetHeadquartersDataListReq) (resp.HeadquartersDataListResp, error)
       GranularSync(ctx context.Context, req req.GranularSyncReq) (resp.GranularSyncResp, error)
   }
   ```

2. 在 `main/app/service/sync.go` 中实现方法
   ```go
   func (s *SyncSrv) GetHeadquartersDataList(ctx context.Context, req req.GetHeadquartersDataListReq) (resp.HeadquartersDataListResp, error) {
       companySetting := ctx.GetCompanySetting()
       
       // 检查是否为分店
       if !companySetting.IsSubShop() {
           return resp.HeadquartersDataListResp{}, errors.New("非分店账号无法查看总部数据")
       }
       
       headquarterUuid := companySetting.HeadquarterUuid
       subShopUuid := companySetting.CompanyUuid
       
       headquarterDB := s.dbm.GetDB(headquarterUuid)
       subShopDB := s.dbm.GetDB(subShopUuid)
       
       // 定义要查询的数据类型
       dataTypes := req.DataTypes
       if len(dataTypes) == 0 {
           dataTypes = []string{
               constant.SyncDataTypeProductCategory,
               constant.SyncDataTypeUnit,
               // ... 所有类型
           }
       }
       
       // 查询各类型数据
       var dataGroups []resp.DataGroup
       for _, dataType := range dataTypes {
           group, err := s.getDataGroupByType(ctx, dataType, headquarterDB, subShopDB, headquarterUuid)
           if err != nil {
               logger.Logger.Error("查询数据失败", zap.String("dataType", dataType), zap.Error(err))
               continue
           }
           dataGroups = append(dataGroups, group)
       }
       
       return resp.HeadquartersDataListResp{
           DataGroups: dataGroups,
       }, nil
   }
   ```

**验收标准**:
- [ ] 方法已实现
- [ ] 检查分店权限
- [ ] 调用 `getDataGroupByType` 查询数据
- [ ] 编译通过

**预估工时**: 1h

---

### Task 4.2: 实现 getDataGroupByType 方法

**目标**: 根据数据类型查询数据分组

**步骤**:

1. 在 `sync.go` 中实现方法
   ```go
   func (s *SyncSrv) getDataGroupByType(ctx context.Context, dataType string, headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
       switch dataType {
       case constant.SyncDataTypeProductCategory:
           return s.getProductCategoryGroup(headquarterDB, subShopDB, headquarterUuid)
       case constant.SyncDataTypeCoupon:
           return s.getCouponGroup(headquarterDB, subShopDB, headquarterUuid)
       case constant.SyncDataTypeProductLabel:
           return s.getProductLabelGroup(headquarterDB, subShopDB, headquarterUuid)
       case constant.SyncDataTypeFullReduction:
           return s.getFullReductionGroup(headquarterDB, subShopDB, headquarterUuid)
       case constant.SyncDataTypeMarketingActivity:
           return s.getMarketingActivityGroup(headquarterDB, subShopDB, headquarterUuid)
       case constant.SyncDataTypePaymentMethod:
           return s.getPaymentMethodGroup(headquarterDB, subShopDB, headquarterUuid)
       default:
           return resp.DataGroup{}, errors.New("不支持的数据类型")
       }
   }
   ```

**验收标准**:
- [ ] 方法已实现
- [ ] switch 分支完整
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 4.3: 实现 getCouponGroup 方法

**目标**: 查询优惠券数据分组

**步骤**:

1. 在 `sync.go` 中实现方法
   ```go
   func (s *SyncSrv) getCouponGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
       // 1. 查询总部优惠券（headquarter_uuid = 0）
       var hqCoupons []model.MarketingCoupon
       err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").
           Find(&hqCoupons).Error
       if err != nil {
           return resp.DataGroup{}, errors.WithMessage(err, "查询总部优惠券失败")
       }
       
       // 2. 查询分店已同步的优惠券uuid列表
       var syncedUuids []uint64
       err = subShopDB.Model(&model.MarketingCoupon{}).
           Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
           Pluck("uuid", &syncedUuids).Error
       if err != nil {
           return resp.DataGroup{}, errors.WithMessage(err, "查询分店已同步优惠券失败")
       }
       
       // 3. 组装数据项
       var items []resp.DataItem
       for _, coupon := range hqCoupons {
           items = append(items, resp.DataItem{
               Uuid:        coupon.Uuid,
               Name:        coupon.Name,
               RelatedData: []resp.RelatedData{}, // 优惠券无关联数据
               AdditionalInfo: map[string]any{
                   "amount": coupon.Amount,
                   "status": coupon.Status,
               },
           })
       }
       
       return resp.DataGroup{
           Type:        constant.SyncDataTypeCoupon,
           TypeName:    constant.SyncDataTypeNames[constant.SyncDataTypeCoupon],
           Items:       items,
           SyncedUuids: syncedUuids, // 返回已同步的uuid列表
       }, nil
   }
   ```

**验收标准**:
- [ ] 方法已实现
- [ ] 查询总部数据列表
- [ ] 查询分店已同步uuid列表（使用Pluck）
- [ ] 组装响应数据（items + synced_uuids）
- [ ] 编译通过

**预估工时**: 1h

---

### Task 4.4: 实现 getProductLabelGroup 方法

**目标**: 查询菜品标签数据分组（包含关联商品）

**步骤**:

1. 在 `sync.go` 中实现方法（参考 design.md 中的示例代码）
2. **重点**：使用 `Preload("ProductPackages")` 加载关联商品
3. 提取关联商品的 uuid 列表，构建 `RelatedData` 结构（**明确类型为 "product"**）
4. 支持多种关联数据类型（如商品可能关联：分类、单位、规格、属性、加料、成本卡等）

**验收标准**:
- [ ] 方法已实现
- [ ] 查询总部菜品标签列表（Preload关联商品）
- [ ] 查询分店已同步uuid列表（使用Pluck）
- [ ] 提取关联商品 uuid，使用 `RelatedData` 结构，明确关联类型为 "product"
- [ ] 组装响应数据（items + synced_uuids）
- [ ] 编译通过

**预估工时**: 1h

---

### Task 4.5: 实现 getFullReductionGroup 方法

**目标**: 查询满额减数据分组

**步骤**: 参考 Task 4.3，查询满额减活动

**验收标准**:
- [ ] 方法已实现
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 4.6: 实现 getMarketingActivityGroup 方法

**目标**: 查询营销活动数据分组

**步骤**: 参考 Task 4.3，查询营销活动

**验收标准**:
- [ ] 方法已实现
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 4.7: 实现 getPaymentMethodGroup 方法（⚠️ 名称匹配）

**目标**: 查询支付方式数据分组（通过名称匹配判断已同步状态）

**步骤**: 

1. 在 `sync.go` 中实现方法（参考 design.md 中的完整代码）
2. **关键逻辑**：
   - 查询总部支付方式（过滤 code=40, 10）
   - 查询分店中 `headquarter_uuid = 总部uuid` 的支付方式**名称列表**
   - 通过名称匹配，找出已同步的总部uuid
   - 返回 items（总部列表）+ synced_uuids（已同步的总部uuid）

3. **完整代码示例**见 `design.md` 第666-720行

**验收标准**:
- [ ] 方法已实现
- [ ] 正确过滤 code=40 和 code=10
- [ ] 查询分店中 `headquarter_uuid = 总部uuid` 的支付方式名称
- [ ] 通过 payment_name 匹配总部支付方式
- [ ] 返回的 synced_uuids 是总部uuid列表
- [ ] 编译通过

**预估工时**: 1h

---

### Task 4.8: 实现 GranularSync 方法

**目标**: 实现颗粒化同步的入口方法

**步骤**:

1. 在 `sync.go` 中实现方法
   ```go
   func (s *SyncSrv) GranularSync(ctx context.Context, req req.GranularSyncReq) (resp.GranularSyncResp, error) {
       companySetting := ctx.GetCompanySetting()
       
       // 检查是否为分店
       if !companySetting.IsSubShop() {
           return resp.GranularSyncResp{}, errors.New("非分店账号无法执行颗粒化同步")
       }
       
       companyUuid := companySetting.CompanyUuid
       
       // 检查是否已有同步任务在运行
       if !syncTaskManager.tryStartTask(companyUuid) {
           return resp.GranularSyncResp{}, errors.New("数据同步中，请稍后再试")
       }
       
       // 创建同步任务
       syncTaskRepo := repository.NewSyncTaskRepo(s.dbm.GetDB(companyUuid))
       syncTask := &model.SyncTask{
           Status:     constant.SyncTaskStatusRunning,
           StartTime:  time.Now().Unix(),
       }
       
       if err := syncTaskRepo.Create(syncTask); err != nil {
           syncTaskManager.finishTask(companyUuid)
           return resp.GranularSyncResp{}, errors.WithMessage(err, "创建同步任务失败")
       }
       
       // 异步执行
       utils.Go(func() {
           s.executeGranularSync(ctx, syncTask, req.SyncData)
       })
       
       return resp.GranularSyncResp{
           TaskUuid: syncTask.Uuid,
           Message:  "数据同步已启动，可在同步历史中查看进度",
       }, nil
   }
   ```

**验收标准**:
- [ ] 方法已实现
- [ ] 检查分店权限
- [ ] 检查同步任务是否在运行
- [ ] 创建同步任务记录
- [ ] 异步执行同步
- [ ] 编译通过

**预估工时**: 1h

---

### Task 4.9: 实现 executeGranularSync 方法

**目标**: 执行颗粒化同步的核心逻辑

**步骤**:

1. 在 `sync.go` 中实现方法（参考 design.md 中的示例代码）
2. **Step 1**: 调用 `deleteUncheckedHeadquartersData` 删除未勾选数据
3. **Step 2**: 按顺序同步勾选的数据
4. **Step 3**: 更新任务状态，推送 WebSocket 通知

**验收标准**:
- [ ] 方法已实现
- [ ] 删除未勾选数据
- [ ] 同步勾选数据
- [ ] 更新任务状态
- [ ] 推送 WebSocket
- [ ] panic 恢复处理
- [ ] 编译通过

**预估工时**: 1.5h

---

### Task 4.10: 实现 deleteUncheckedHeadquartersData 方法

**目标**: 删除分店中未勾选的总部数据

**步骤**:

1. 在 `sync.go` 中实现方法
   ```go
   func (s *SyncSrv) deleteUncheckedHeadquartersData(ctx context.Context, syncData req.GranularSyncData, headquarterUuid uint64) error {
       subShopDB := s.dbm.GetDB(ctx.GetCompanyUuid())
       
       // 定义删除任务
       deleteTasks := []struct {
           TableName  string
           Uuids      []uint64
           SkipDelete bool
       }{
           {"ttpos_marketing_coupon", syncData.Coupon, false},
           {"ttpos_full_reduction_activity", syncData.FullReduction, false},
           {"ttpos_product_label", syncData.ProductLabel, false},
           {"ttpos_marketing_activity", syncData.MarketingActivity, false},
           {"ttpos_payment_method", syncData.PaymentMethod, true}, // 支付方式不删除
           // ... 其他表
       }
       
       for _, task := range deleteTasks {
           if task.SkipDelete {
               continue
           }
           
           // 构建查询：总部来源且未勾选
           query := subShopDB.Table(task.TableName).
               Where("headquarter_uuid = ?", headquarterUuid)
           
           if len(task.Uuids) > 0 {
               query = query.Where("uuid NOT IN (?)", task.Uuids)
           }
           
           // 硬删除
           err := query.Unscoped().Delete(&map[string]any{}).Error
           if err != nil {
               logger.Logger.Error("删除未勾选数据失败", zap.String("table", task.TableName), zap.Error(err))
               return errors.WithMessage(err, fmt.Sprintf("删除%s失败", task.TableName))
           }
       }
       
       return nil
   }
   ```

**验收标准**:
- [ ] 方法已实现
- [ ] 支付方式不删除
- [ ] 硬删除（Unscoped）
- [ ] 错误处理
- [ ] 编译通过

**预估工时**: 1h

---

### Task 4.11: 实现 SyncXxxByUuids 方法（优惠券）

**目标**: 实现按 uuid 同步优惠券

**步骤**:

1. 在 `sync.go` 中实现方法
   ```go
   func (s *SyncSrv) SyncMarketingCouponByUuids(ctx context.Context, uuids []uint64) error {
       if len(uuids) == 0 {
           return nil
       }
       
       companySetting := ctx.GetCompanySetting()
       headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
       subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)
       
       // 查询总部优惠券
       var hqCoupons []model.MarketingCoupon
       err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0 AND uuid IN (?)", uuids).
           Find(&hqCoupons).Error
       if err != nil {
           return errors.WithMessage(err, "查询总部优惠券失败")
       }
       
       // 同步到分店（先删除再创建）
       for _, hqCoupon := range hqCoupons {
           subShopDB.Unscoped().Where("uuid = ?", hqCoupon.Uuid).Delete(&model.MarketingCoupon{})
           
           newCoupon := hqCoupon
           newCoupon.HeadquarterUuid = companySetting.HeadquarterUuid
           
           err = subShopDB.Create(&newCoupon).Error
           if err != nil {
               logger.Logger.Error("同步优惠券失败", zap.Uint64("uuid", hqCoupon.Uuid), zap.Error(err))
               continue
           }
       }
       
       return nil
   }
   ```

**验收标准**:
- [ ] 方法已实现
- [ ] 先删除后创建
- [ ] 标记 `headquarter_uuid`
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 4.12: 实现 SyncXxxByUuids 方法（其他4种）

**目标**: 实现按 uuid 同步满额减、菜品标签、营销活动、支付方式

**步骤**:

1. 参考 Task 4.11，实现以下方法：
   - `SyncFullReductionByUuids`
   - `SyncProductLabelByUuids`
   - `SyncMarketingActivityByUuids`
   - `SyncPaymentMethodByUuids`（特殊处理：同名跳过）

2. **支付方式特殊处理**:
   ```go
   func (s *SyncSrv) SyncPaymentMethodByUuids(ctx context.Context, uuids []uint64) error {
       // ... 查询总部支付方式
       
       for _, hqPayment := range hqPayments {
           // 检查分店是否已有同名支付方式
           var existPayment model.PaymentMethod
           err := subShopDB.Where("name = ? AND delete_time = 0", hqPayment.Name).First(&existPayment).Error
           if err == nil {
               logger.Logger.Info("支付方式已存在，跳过同步", zap.String("name", hqPayment.Name))
               continue
           }
           
           // 创建新支付方式
           // ...
       }
   }
   ```

**验收标准**:
- [ ] 4个方法已实现
- [ ] 支付方式同名跳过
- [ ] 编译通过

**预估工时**: 1.5h

---

## Phase 5: API 层实现

### Task 5.1: 创建 API Handler

**目标**: 创建同步相关的 API Handler

**步骤**:

1. 检查是否已有 `shop_sync.go`，如果没有则创建 `main/app/api/v1/shop/shop_sync.go`
   ```go
   type SyncHandler struct {
       syncSrv service.ISyncSrv
   }
   
   func NewSyncHandler(syncSrv service.ISyncSrv) *SyncHandler {
       return &SyncHandler{syncSrv: syncSrv}
   }
   ```

2. 或者在现有 `shop_setting.go` 中添加方法

**验收标准**:
- [ ] Handler 已创建
- [ ] 注入 `syncSrv` 依赖
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 5.2: 实现 GetHeadquartersDataList API

**目标**: 实现获取总部可同步数据列表的 API

**步骤**:

1. 在 Handler 中添加方法
   ```go
   // GetHeadquartersDataList 获取总部可同步数据列表
   // @Summary 获取总部可同步数据列表
   // @Description 获取总部可同步数据列表（按种类分组）
   // @Tags 商家端.数据同步
   // @Accept json
   // @Produce json
   // @Security JwtToken
   // @Param data body req.GetHeadquartersDataListReq true "请求参数"
   // @Success 200 {object} resp.HeadquartersDataListResp
   // @Router /shop/sync/headquarters_data_list [post]
   func (h *SyncHandler) GetHeadquartersDataList(c *gin.Context) {
       ctx := helper.GetContext(c)
       
       var req req.GetHeadquartersDataListReq
       if err := c.ShouldBindJSON(&req); err != nil {
           helper.ErrorWithDetail(c, constant.CodeInvalidParam, err)
           return
       }
       
       resp, err := h.syncSrv.GetHeadquartersDataList(ctx, req)
       if err != nil {
           helper.ErrorWithDetail(c, constant.CodeFail, err)
           return
       }
       
       helper.Success(c, gin.H{"data": resp})
   }
   ```

**验收标准**:
- [ ] API 方法已实现
- [ ] 参数绑定和验证
- [ ] Swagger 注释完整
- [ ] 编译通过

**预估工时**: 0.5h

---

### Task 5.3: 实现 GranularSync API

**目标**: 实现颗粒化同步的 API

**步骤**:

1. 在 Handler 中添加方法（参考 Task 5.2）
2. 注册路由
   ```go
   // main/router/shop_router.go
   shopSync := shopGroup.Group("/sync")
   {
       shopSync.POST("/headquarters_data_list", shopSyncHandler.GetHeadquartersDataList)
       shopSync.POST("/granular_sync", shopSyncHandler.GranularSync)
   }
   ```

**验收标准**:
- [ ] API 方法已实现
- [ ] 路由已注册
- [ ] Swagger 注释完整
- [ ] 编译通过

**预估工时**: 1h

---

## Phase 6: 测试

### Task 6.1: 单元测试（Service 层）

**目标**: 为 Service 层方法编写单元测试

**步骤**:

1. 创建测试文件 `main/app/service/sync_test.go`
2. 测试用例：
   - `TestGetHeadquartersDataList`: 测试查询总部数据列表
   - `TestGetCouponGroup`: 测试查询优惠券分组
   - `TestDeleteUncheckedHeadquartersData`: 测试删除未勾选数据
   - `TestSyncMarketingCouponByUuids`: 测试按uuid同步优惠券

3. 使用 Mock 数据库和依赖
   ```go
   func TestGetCouponGroup(t *testing.T) {
       // Mock DB
       // Mock Data
       // Call Method
       // Assert Result
   }
   ```

**验收标准**:
- [ ] 测试文件已创建
- [ ] 核心方法已测试
- [ ] 测试通过

**预估工时**: 2h

---

### Task 6.2: API 测试

**目标**: 使用 Postman 或 curl 测试 API 接口

**步骤**:

1. 准备测试数据：在总部创建测试数据（优惠券、满额减等）
2. 测试 `GetHeadquartersDataList` API
   ```bash
   curl -X POST http://localhost:8080/api/v1/shop/sync/headquarters_data_list \
   -H "Authorization: Bearer {token}" \
   -H "Content-Type: application/json" \
   -d '{"data_types": ["coupon", "product_label"]}'
   ```

3. 测试 `GranularSync` API
   ```bash
   curl -X POST http://localhost:8080/api/v1/shop/sync/granular_sync \
   -H "Authorization: Bearer {token}" \
   -H "Content-Type: application/json" \
   -d '{
     "sync_data": {
       "coupon": [123456],
       "product_label": [789012]
     }
   }'
   ```

4. 验证响应数据格式和内容

**验收标准**:
- [ ] 两个API都能正常调用
- [ ] 响应格式正确
- [ ] 数据内容正确

**预估工时**: 1h

---

### Task 6.3: 集成测试（端到端）

**目标**: 测试完整的同步流程

**步骤**:

1. 准备测试环境：
   - 1个总部账号
   - 1个分店账号
   - 总部创建测试数据（优惠券、满额减、菜品标签等）

2. 执行测试流程：
   - 分店查询总部数据列表
   - 验证数据列表正确（包含总部数据，标记已同步状态）
   - 分店执行颗粒化同步（勾选部分数据）
   - 验证同步任务创建
   - 等待同步完成
   - 验证分店数据库中的数据（已勾选的存在，未勾选的被删除）

3. 测试边界情况：
   - 非分店账号调用（应失败）
   - 同步任务已在运行（应失败）
   - 支付方式同名冲突（应跳过）

**验收标准**:
- [ ] 完整流程测试通过
- [ ] 边界情况处理正确
- [ ] 数据一致性验证通过

**预估工时**: 2h

---

### Task 6.4: 多语言数据同步测试

**目标**: 测试满额减、营销活动的多语言数据同步

**步骤**:

1. 在总部创建满额减活动（包含多语言名称）
2. 分店同步该活动
3. 验证分店数据库中：
   - 满额减活动记录正确
   - 多语言数据（`multi_language_name` 表）已同步

**验收标准**:
- [ ] 多语言数据同步正确
- [ ] 前端显示多语言名称正确

**预估工时**: 1h

---

### Task 6.5: 支付方式同步规则测试（⚠️ 重点）

**目标**: 测试支付方式的复杂同步规则

**步骤**:

1. **测试用例1：过滤特定code**
   - 总部创建支付方式（code=40, code=10）
   - 分店查询可同步列表
   - 验证：这两个支付方式不出现在列表中

2. **测试用例2：已同步状态判断（名称匹配）**
   - 总部DB：支付方式 "微信支付"（uuid=111111）、"支付宝"（uuid=222222）
   - 分店DB：支付方式 "微信支付"（uuid=999991, **headquarter_uuid=总部uuid**）
   - 分店查询可同步列表
   - 验证：
     - `synced_uuids = [111111]`（总部uuid，通过名称匹配）
     - items 包含111111和222222

3. **测试用例3：首次同步**
   - 总部创建支付方式 "支付宝"（code=90001）
   - 分店不存在 "支付宝"
   - 分店同步
   - 验证：创建新支付方式，code自动生成，logo_file_uuid=0

4. **测试用例4：同名普通code**
   - 总部创建支付方式 "微信支付"（code=90002）
   - 分店已有 "微信支付"（code=90002, headquarter_uuid=0）
   - 分店同步
   - 验证：跳过同步，分店支付方式不变

5. **测试用例5：同名特殊code**
   - 总部创建支付方式 "现金"（code=90111）
   - 分店已有 "现金"（code=90111, headquarter_uuid=0）
   - 分店同步
   - 验证：只更新 headquarter_uuid=总部uuid，其他字段不变

6. **测试用例6：不删除未勾选**
   - 分店已有总部同步的支付方式 "A"（headquarter_uuid=总部uuid）
   - 分店执行同步，但未勾选 "A"
   - 验证：支付方式 "A" 仍然存在（不被删除）

**验收标准**:
- [ ] 所有6个测试用例通过
- [ ] 名称匹配逻辑正确（用例2）
- [ ] code 生成正确（用例3）
- [ ] 特殊code规则正确（用例5）
- [ ] 不删除未勾选数据（用例6）

**预估工时**: 1.5h

---

### Task 6.6: 性能测试

**目标**: 测试大数据量下的性能

**步骤**:

1. 在总部创建大量测试数据（100条优惠券、100条菜品标签等）
2. 分店查询总部数据列表，记录响应时间
3. 分店执行颗粒化同步，记录同步时间
4. 验证性能指标：
   - 查询总部数据列表: < 1s
   - 颗粒化同步100条数据: < 30s

**验收标准**:
- [ ] 性能指标达标
- [ ] 无明显性能瓶颈

**预估工时**: 1h

---

## Phase 7: 部署和联调

### Task 7.1: 更新 API 文档

**目标**: 生成最新的 Swagger API 文档

**步骤**:

1. 运行 Swagger 生成命令
   ```bash
   cd main
   swag init -g main.go
   ```

2. 验证 API 文档正确显示
   - 访问 `http://localhost:8080/swagger/index.html`
   - 查看新增的2个API接口

**验收标准**:
- [ ] Swagger 文档已更新
- [ ] 新API接口正确显示
- [ ] 参数和响应模型正确

**预估工时**: 0.5h

---

### Task 7.2: 部署数据库迁移

**目标**: 在测试环境和生产环境执行数据库迁移

**步骤**:

1. 备份数据库
2. 在测试环境执行迁移
3. 验证迁移成功
4. 在生产环境执行迁移（在维护窗口）

**验收标准**:
- [ ] 测试环境迁移成功
- [ ] 生产环境迁移成功
- [ ] 所有表都有 `headquarter_uuid` 字段

**预估工时**: 1h

---

### Task 7.3: 联调前端

**目标**: 与前端联调，确保接口对接正确

**步骤**:

1. 前端集成2个API接口
2. 联调测试：
   - 前端调用 `GetHeadquartersDataList`，显示数据列表
   - 前端调用 `GranularSync`，执行同步
   - 验证同步结果

3. 修复联调中发现的问题

**验收标准**:
- [ ] 前端能正常调用API
- [ ] 数据显示正确
- [ ] 同步功能正常
- [ ] 无阻塞问题

**预估工时**: 2h

---

## 📊 任务统计

### 按优先级

- **P0 (高优先级)**: 25 个任务
  - Phase 1: 数据库和模型（⚠️ Task 1.2 支付方式规则确认是关键）
  - Phase 2: 常量定义
  - Phase 3: Repository 扩展
  - Phase 4: Service 层实现（⚠️ Task 4.12 支付方式同步最复杂）
  - Phase 5: API 层实现

- **P1 (中优先级)**: 6 个任务
  - Phase 6: 测试（⚠️ Task 6.5 支付方式6个测试用例必做）

- **P2 (低优先级)**: 3 个任务
  - Phase 7: 部署和联调

### 按模块

| 模块 | 任务数 | 预估工时 |
|------|--------|----------|
| 数据库 | 5 | 2.5h |
| 常量 | 1 | 0.5h |
| Repository | 6 | 2h |
| Service | 12 | 9.5h |
| API | 3 | 2h |
| 测试 | 6 | 5.5h |
| 部署 | 3 | 2h |

### 关键路径

```
Task 1.1 → Task 1.3 → Task 1.4 → Task 2.1 → Task 3.x → Task 4.1-4.12 → Task 5.1-5.3 → Task 6.3 → Task 7.3
```

---

## 💡 关键改进说明

### 改进1：关联数据类型明确化

**问题**：原设计中 `RelatedUuids` 只返回uuid列表，前端无法知道这些uuid是什么类型的数据。

**改进**：使用 `RelatedData` 结构，明确每个关联数据的类型。

### 改进2：已同步数据列表优化

**问题**：原设计中每个 `DataItem` 都有 `IsSynced` 字段，数据冗余，且需要统计 `TotalCount` 和 `SyncedCount`。

**改进**：在 `DataGroup` 层级返回 `SyncedUuids` 列表，前端根据此列表标记勾选状态，更清晰高效。

**关联数据示例**：
```json
{
  "uuid": 345678,
  "name": "招牌菜",
  "related_data": [
    {
      "type": "product",        // 明确是商品类型
      "uuids": [111111, 222222] // 关联的商品uuid列表
    }
  ]
}
```

**已同步数据示例**：
```json
{
  "type": "product_label",
  "type_name": "菜品标签",
  "synced_uuids": [345678, 456789], // 已同步的uuid列表
  "items": [
    {
      "uuid": 345678,
      "name": "招牌菜"
    },
    {
      "uuid": 456789,
      "name": "新品"
    },
    {
      "uuid": 567890,
      "name": "热卖"
    }
  ]
}
```
> 前端根据 `synced_uuids` 判断哪些item应该默认勾选

**支持的关联类型**（参考 product_package商品关联表.txt）：

| 数据类型 | 可能的关联类型 | 说明 |
|---------|---------------|------|
| product_label（菜品标签） | product | 标签关联的商品 |
| product（商品） | unit, category, **tax**, flavor, sauce, attribute, bom_card | 商品的基础配置 |
| bom_card（成本卡） | material, unit | 成本卡关联的物品和单位 |
| material（物品） | **unit**（→ product_unit）, material_category | 物品关联单位和分类 |
| payment_method（支付方式） | 无 | 支付方式无关联数据 |

**物品关联单位的获取方式**：
- 直接字段：`unit_uuid`（基准单位）、`purchase_unit_uuid`（采购单位）、`cost_unit_uuid`（成本单位）
- 通过关联表：`material_unit` 表中的 `unit_uuid`（非基准单位列表）
- ⚠️ **所有 unit_uuid 都指向 `product_unit` 表**（不是 material_unit 表）
- 需要 `Preload("NotBaseUnitList")` 加载非基准单位
- 去重后返回所有关联的单位uuid

**商品关联税类的获取方式**：
- 直接字段：`dine_tax_uuid`（堂食税）、`takeout_tax_uuid`（外卖税）
- 去重后返回（可能相同）

**注意**：支付方式的 `synced_uuids` 通过**名称匹配**获得，详见 `PAYMENT_METHOD_SYNC_RULES.md`

**实现要点**：
1. 查询数据时使用 `Preload` 加载关联数据
2. 提取关联数据的 uuid，并明确类型
3. 构建 `RelatedData` 数组返回给前端
4. 前端根据类型提示用户勾选依赖数据
5. **支付方式特殊**：通过 `payment_name` 匹配判断已同步状态（不是 `headquarter_uuid`）

---

## 📝 注意事项

### 开发顺序建议

1. **先完成 Phase 1-2**：数据库和常量（基础）
2. **再完成 Phase 3**：Repository 扩展（数据访问层）
3. **重点 Phase 4**：Service 层实现（核心业务逻辑）
4. **快速 Phase 5**：API 层实现（接口暴露）
5. **最后 Phase 6-7**：测试和部署

### 风险点

1. **支付方式规则复杂**：Task 1.2 和 Task 4.12 涉及复杂的业务规则
   - 需要确认 code 生成规则
   - 需要确认特殊 code（90111, 90222, 90333）的业务含义
   - 需要确认所有字段的默认值
   - **建议**：详细阅读 `PAYMENT_METHOD_SYNC_RULES.md`

2. **外键约束**：Task 4.10 删除数据时可能遇到外键约束，需要按顺序删除

3. **多语言同步**：Task 4.11-4.12 需要确保多语言数据一并同步

4. **性能问题**：Task 6.6 可能发现性能瓶颈，需要优化

### 依赖关系

- Task 1.4 依赖 Task 1.3（数据库迁移完成后才能更新Model）
- Task 4.x 依赖 Task 3.x（Repository 扩展完成后才能使用）
- Task 5.x 依赖 Task 4.x（Service 实现完成后才能调用）
- Task 6.x 依赖 Task 5.x（API 实现完成后才能测试）

---

---

## 📚 参考文档

- **支付方式同步特殊规则**：`PAYMENT_METHOD_SYNC_RULES.md`（⚠️ 必读）
- **关联数据获取指南**：`RELATED_DATA_GUIDE.md`（🔗 各数据类型关联关系）
- 设计文档：`design.md`
- 商品关联关系：`product_package商品关联表.txt`

---

**版本**: v1.1.0  
**创建日期**: 2025-12-05  
**更新日期**: 2025-12-05  
**作者**: 曾振华  
**审核者**: 待分配  
**关联任务**: DooTask #37462
