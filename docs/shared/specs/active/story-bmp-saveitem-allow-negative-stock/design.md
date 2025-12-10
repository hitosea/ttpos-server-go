# saveitem-allow-negative-stock 设计文档

> 技术设计：在 ttpos-bmp `SaveItem` 流程开放并持久化 `allow_negative_stock`，保证兼容性，并在存在负库存时禁止关闭。

## 📋 概述

- 在 `item.proto` 的 `ItemInfo` 增加 `allow_negative_stock`（bool），`SaveItem` 请求/响应透传，`GetItem`/`GetItemList` 等读取时返回。
- 服务端保存时，若入参未提供该字段，沿用原有值；如从 true→false，先通过库存查询（含 Stock Projected Qty）判断是否有负库存，若有则拒绝并提示：“物料已产生负库存，请修正库存后再进行操作。”
- 不改动存量行为与默认值，保证旧调用方不传该字段时无回归。

## 🎯 规范对齐

- **Go BMP 规范**: 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`，不修改 `dao/entity/do`；使用 GoFrame 2.x。
- **Proto 规范**: 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`，字段命名 snake_case，补充注释。
- **API 设计**: 响应 `data` 为对象，错误提示使用统一 message。
- **安全**: 参数校验、权限照旧；不引入 panic。

## 🔄 代码复用与集成点

- 复用现有 `ItemService.SaveItem` 流程及 ERP 响应包装（`erp.ResponseInfo`）。
- 库存校验复用现有库存查询能力（Stock Projected Qty），如已有服务/DAO，可直接调用；否则复用 `GetItemStock` 同步逻辑。
- 读取场景复用当前 ItemInfo/ItemList 序列化链路，新增字段随结构透传。

## 🏗️ 架构与流程

1. **Proto 层**：`ItemInfo` 增加 `bool allow_negative_stock = N;`（位置靠近库存相关字段）。服务端生成代码。
2. **Controller/Logic**：`SaveItem` 写入时：
   - 若请求未带字段：保持原值（从存量记录读取）。
   - 若从 true→false：调用库存查询获取当前 projected/actual；若存在负库存则返回错误并提示固定文案。
   - 允许 true→true / false→true 直接保存。
3. **读取接口**：`GetItem`/`GetItemList` 等返回体带上 `allow_negative_stock`。
4. **存储/透传**：沿用现有存储/ERP 映射，新增字段一并读写。

## 🔌 API/Proto 设计

- **Proto**：`ItemInfo` 增补 `allow_negative_stock`；若有嵌套 DTO/存储模型，保持同名布尔字段。
- **SaveItem**：请求/响应包含该字段；错误提示固定：“物料已产生负库存，请修正库存后再进行操作。”
- **GetItem / GetItemList**：返回 `allow_negative_stock`。

## 🧩 组件设计

- **库存校验**：封装函数 `checkCanDisableNegativeStock(itemCode, branch, companyAbbr)`，内部查询 Stock Projected Qty；若 `<0` 则返回业务错误。
- **默认值策略**：请求未带字段时不覆盖；若存量无字段时采用当前线上默认（保持兼容）。
- **日志与监控**：在拒绝关闭时记录提示与 item_code/branch。

## 🚨 错误处理

- 关闭校验失败：返回业务错误，message 使用固定文案；code 复用现有业务错误码（与 SaveItem 其他校验一致）。
- 其他异常：保持现有错误包装与日志。

## 🧪 测试策略

- 单元/逻辑测试：保存时字段透传；未传字段保持原值；true→false 且存在负库存时报错；无负库存时成功关闭。
- API/集成测试：SaveItem 设置/关闭、GetItem/GetItemList 返回字段；兼容旧请求不带字段的路径。
- 如有缓存/投递，验证字段同步。

## 📚 实现清单

- 详见 `tasks.md`。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-10  
**作者**: rikugun  
**审核者**: 待定  
