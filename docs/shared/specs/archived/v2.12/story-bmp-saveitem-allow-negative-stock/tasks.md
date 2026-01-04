# saveitem-allow-negative-stock 任务分解
> 执行 SaveItem `allow_negative_stock` 开放与校验的落地任务。

## 📊 进度总览
**总任务数**: 8  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: Proto 与模型
- [ ] 1.1 更新 Proto
  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/item.proto`
  - Purpose: 在 `ItemInfo` 添加 `allow_negative_stock`（bool），用于 SaveItem/查询透传
  - Requirements: Req1, Req2, Req3
  - Leverage: 现有字段布局（库存相关段）

- [ ] 1.2 生成代码
  - Command: `cd ttpos-bmp/app/ttpos-erp && make dao`（或项目规定命令）
  - Purpose: 生成/刷新 Go 代码
  - Requirements: Req1, Req3

## Phase 2: 业务与校验
- [ ] 2.1 SaveItem 入参透传与默认值保持
  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/...`（SaveItem 逻辑所在）
  - Purpose: 保存时支持 bool 入参；未传时保持原值
  - Requirements: Req1

- [ ] 2.2 关闭前负库存校验
  - File: 同上逻辑层
  - Purpose: true→false 时查询 Stock Projected Qty（库存查询能力），如负库存则拒绝，提示固定文案
  - Requirements: Req2

- [ ] 2.3 读取接口透传
  - File: `GetItem` / `GetItemList` 返回路径
  - Purpose: 返回体带上 `allow_negative_stock`
  - Requirements: Req1, Req3

## Phase 3: 测试
- [ ] 3.1 单元/逻辑测试
  - File: 与 SaveItem 逻辑同目录的 `_test.go`
  - Purpose: 覆盖 set/keep/default、true→false 阻断、有/无负库存分支
  - Requirements: Req1, Req2

- [ ] 3.2 API/集成测试
  - File: API 集成测试位置
  - Purpose: SaveItem 设置/关闭、GetItem/GetItemList 返回字段、兼容未传字段
  - Requirements: Req1, Req2, Req3

## Phase 4: 文档与检查
- [ ] 4.1 更新设计/需求引用
  - File: `design.md`, 如需补充变更日志/AC 明确处
  - Purpose: 确认字段位置、提示文案一致
  - Requirements: 全部

- [ ] 4.2 质量检查
  - Command: `go fmt`, `go test ./...`
  - Purpose: 确保格式与测试通过
  - Requirements: 全部

---

## 提交前检查
- [ ] requirements.md / design.md / tasks.md 同步且勾选完成
- [ ] gofmt / go vet / go test 通过
- [ ] 新增/变更接口的提示文案与提案一致
- [ ] Proto 生成产物已更新并纳入变更集
