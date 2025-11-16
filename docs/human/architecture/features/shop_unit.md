### 模块：ShopUnit（商品单位）

定位：对SHOP端提供标准化的HTTP API能力，负责商品单位展示、新增、编辑、删除、关联商品

业务背景

- 商家在商品管理中需要统一维护计量单位（份、杯、盒等），单位数据需支撑商品包、套餐、ERP 同步等场景
- 单位数据具备多语言能力，需确保必填语言完整以及与 ERPNext UOM 的映射
- 商家自建单位可编辑/删除，总部下发的单位只读，需在界面展示 `is_editable`

主要接口（main/api/v1/shop/shop_product.go）

- 获取单位列表
  - 方法：GET
  - 路径：/shop/product/unit/list
  - 入参：page_no、page_size（query）
  - 出参：product_resp.ProductUnitListResp
- 获取单位详情
  - 方法：GET
  - 路径：/shop/product/unit
  - 入参：uuid（query）
  - 出参：product_resp.ProductUnitDetail
- 新增单位
  - 方法：POST
  - 路径：/shop/product/unit/add
  - 入参：req.ProductUnitAddReq（body）
  - 出参：空对象
- 编辑单位
  - 方法：POST
  - 路径：/shop/product/unit/edit
  - 入参：req.ProductUnitEditReq（body）
  - 出参：空对象
- 删除单位
  - 方法：DELETE
  - 路径：/shop/product/unit
  - 入参：req.ProductUnitReq（body）
  - 出参：空对象
- 排序单位
  - 方法：POST
  - 路径：/shop/product/unit/sort
  - 入参：req.ProductUnitSortReq（body）
  - 出参：空对象

核心数据结构（main/dto）

- 请求（req/product.go）
  - ProductUnitListReq：分页查询
  - ProductUnitReq：{ uuid }
  - ProductUnitAddReq：{ locale_name, product_package_uuids[] }
  - ProductUnitEditReq：{ uuid, locale_name, product_package_uuids[] }
  - ProductUnitSortReq：{ list[{ uuid, sort }] }
- 响应（resp/product_resp/product.go）
  - ProductUnitListResp：{ list[ProductUnitItem], meta{ page_no,page_size,total } }
  - ProductUnitItem：{ uuid, sort, product_package_count, name, is_editable }
  - ProductUnitDetail：{ uuid, locale_name, product_packages[{ uuid,name }], is_editable }

数据库模型（main/app/model/product.go）

- 表名：ttpos_product_unit（ProductUnit）
- 关键字段：
  - uuid：主键
  - name：单位名称（JSON，多语言原始存储）
  - multi_language_name_uuid：多语言名称关联
  - sort：排序（越小越靠前）
  - erpnext_uom：ERPNext 对应单位
  - headquarter_uuid：总部下发单位的总部标识（0 表示门店自建）
- 关联：ProductPackage.unit_uuid，统计字段 `product_package_count` 用于列表展示关联数量

权限与鉴权

- 路由均挂载在 `/shop` 私有路由下，依赖 `middleware.Auth`，必须携带 JwtToken
- 删除、编辑前端需根据返回的 `is_editable` 控制按钮，后端再次校验，防止越权

请求/响应示例

```json
GET /shop/product/unit/list?page_no=1&page_size=20
Response {
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "uuid": 10001,
        "name": "份",
        "sort": 1,
        "product_package_count": 5,
        "is_editable": true
      }
    ],
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 1
    }
  }
}
```

```json
POST /shop/product/unit/add
Body {
  "locale_name": {
    "zh_cn": "份",
    "en_us": "portion"
  },
  "product_package_uuids": [20001, 20002]
}
```

关键流程

- 列表查询
  - 服务：productSrv.GetProductUnitList(ctx, req)
  - 逻辑：分页查询单位，按语言返回名称，返回分页信息与是否可编辑标记
- 详情查询
  - 服务：productSrv.GetProductUnit(ctx, req)
  - 逻辑：加载单位多语言与关联商品包，按语言组装名称列表与商品包名称；计算 is_editable
- 新增
  - 服务：productSrv.AddProductUnit(ctx, req)
  - 校验：必填语言是否齐全（基于门店语言设置）
  - 事务：创建单位（排序为当前最大排序+1）→ 可选更新商品包的单位UUID
  - ERP：如开启ERP，创建/同步 UOM，并回写 erpnext_uom；必要时触发商品同步
- 编辑
  - 服务：productSrv.EditProductUnit(ctx, req)
  - 校验：单位存在、名称存在、单位可编辑
  - 事务：更新多语言与单位名称
- 删除
  - 服务：productSrv.DeleteProductUnit(ctx, req)
  - 校验：单位存在、名称存在、单位可删除、无关联商品包
  - 事务：如开启ERP则删除ERP UOM（忽略 not found），软删单位与多语言名称→重排 sort
- 排序
  - 服务：productSrv.SortProductUnit(ctx, req)
  - 校验：所有传入UUID均存在
  - 逻辑：批量更新 sort

错误与边界

- 通用
  - 获取单位列表失败：服务返回“获取单位列表失败”
  - 获取单位详情失败：服务返回“获取单位详情失败”
- 新增
  - 名称不能为空（必填语言不全）
  - 保存名称多语言失败 / 保存单位失败 / 同步单位到erp失败 / 保存erp单位失败
- 编辑
  - 单位不存在 / 单位名称不存在 / 单位不可编辑
  - 保存名称多语言失败 / 保存单位失败
- 删除
  - 单位不存在 / 单位名称不存在 / 单位不可删除
  - 该单位下存在商品，不允许删除
  - 删除单位失败 / 删除名称多语言失败 / 重新排序单位失败
- 排序
  - 单位不存在（当传入的UUID集合与实际数量不一致）

依赖

- 服务层（main/app/service/product.go）
  - 数据库：pkg/database.DBManager
  - 设置服务：setting.ISrv（获取门店语言）
  - 翻译服务：translateSrv（管理多语言名称待翻译集合）
  - ERP服务：erp.IErpSrv（UOM创建/删除与商品同步）
  - Repository：repository.NewProductRepo / NewProductUnitRepo（数据访问）

开发注意事项

- 多语言校验：`dto.LocaleResponse` 的必填语言列表来自 `settingSrv.GetStoreLanguage`，新增/编辑必须覆盖 `is_required=true` 的语言
- ERP 同步：仅当 `ctx.GetCompany().IsOpenErp()` 且配置完整（站点代码/缩写/分店）才会调用 ERP 服务；失败需回滚事务
- 可编辑判断：`headquarter_uuid>0` 表示总部数据，`isEditable` 返回 false；前端需根据 `is_editable` 禁用操作
- 排序策略：新增默认放在末尾（现有最大 sort +1）；删除后会重新顺序排序为 1..n；排序接口要求 list 覆盖所有需要调整的 uuid
- 关联更新：新增/编辑时可传 `product_package_uuids` 更新商品包单位，需考虑后续商品同步、库存影响

测试覆盖建议

- 列表接口
  - 查询空数据、分页边界、排序字段正确性
  - 多语言切换验证名称是否随语言切换
- 详情接口
  - 验证 `product_packages` 返回的数量和名称
  - `is_editable` 正确显示（总部与门店数据各测一次）
- 新增接口
  - 缺失必填语言 → 返回“名称不能为空”
  - 附带商品包 → 验证商品包单位被更新
  - ERP 开启/关闭两种情况的事务回滚、成功路径
- 编辑接口
  - 编辑总部单位 → 报错“单位不可编辑”
  - 修改多语言后，多语言服务待翻译集合应清理（验证调用）
- 删除接口
  - 删除已关联商品包的单位 → 报错“该单位下存在商品，不允许删除”
  - ERP UOM 不存在时，返回错误消息是否忽略 `not found`
- 排序接口
  - 传入重复/缺失 uuid → 返回“单位不存在”
  - 排序后列表顺序与 sort 字段一致
- 并发/数据一致性
  - 并发新增与排序的事务冲突
  - 删除后再次抓取列表确认重排成功

监控与日志

- 所有接口统一通过 `helper.Success` / `helper.ErrorWithDetail` 输出日志，异常路径附带栈信息
- ERP 同步失败时日志包含 ERP 请求参数与响应，便于排查

开放问题

- 是否需要在删除单位时同步清理 ERP 商品引用（当前仅删除 UOM）
- 商品包关联单位大批量更新时是否需要异步处理以避免事务过大
