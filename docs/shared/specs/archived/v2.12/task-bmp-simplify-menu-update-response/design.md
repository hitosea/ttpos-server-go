# 简化菜单更新方法返回值 设计文档

> 本文档定义简化 `UpdateMenuItem` 和 `UpdateMenuModifier` 返回值的技术设计和实现方案。

## 📋 概述

移除冗余的 `grabDto.UpdateMenuResult` 结构体，将菜单更新方法返回值简化为 Go 标准的 `error` 类型，符合 Go 语言惯例。

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- 禁止修改 dao/entity/do/ 目录
- 遵循 GoFrame 项目结构
- 错误处理使用 `gerror` 包

---

## 🔄 代码复用分析

### 需要修改的现有组件

- **Logic 层**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go` - 修改方法签名和实现
- **DTO 层**: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update.go` - 删除 `UpdateMenuResult` 结构体
- **Service 接口**: `ttpos-bmp/app/ttpos-takeout/internal/service/grab_menu.go` - 更新接口定义

---

## 🏗️ 架构设计

### 变更前后对比

**Before**:

```go
func (s *sGrabMenu) UpdateMenuItem(ctx context.Context, req *grabDto.UpdateMenuItemReq) (*grabDto.UpdateMenuResult, error)
func (s *sGrabMenu) UpdateMenuModifier(ctx context.Context, req *grabDto.UpdateMenuModifierReq) (*grabDto.UpdateMenuResult, error)
```

**After**:

```go
func (s *sGrabMenu) UpdateMenuItem(ctx context.Context, req *grabDto.UpdateMenuItemReq) error
func (s *sGrabMenu) UpdateMenuModifier(ctx context.Context, req *grabDto.UpdateMenuModifierReq) error
```

---

## 🧩 组件和接口

### Logic 层变更

#### UpdateMenuItem 简化后

```go
// UpdateMenuItem 更新单个菜单项 (商品)
func (s *sGrabMenu) UpdateMenuItem(ctx context.Context, req *grabDto.UpdateMenuItemReq) error {
    g.Log().Infof(ctx, "[Grab] UpdateMenuItem: merchantID=%s, itemID=%s", req.MerchantID, req.ItemID)

    // 1. 参数验证
    if err := g.Validator().Data(req).Run(ctx); err != nil {
        g.Log().Errorf(ctx, "[Grab] UpdateMenuItem validation failed: %v", err)
        return gerror.NewCode(gcode.CodeValidationFailed, err.Error())
    }

    // 2. 构建 SDK 请求
    updateItem := req.ToSDKUpdateMenuItem()
    updateReq := grabfood.UpdateMenuItemAsUpdateMenuRequest(updateItem)

    // 3. 调用 Grab API
    if err := service.Grab().UpdateMenuRecord(ctx, req.MerchantID, updateReq); err != nil {
        s.logMenuRecordUpdate(ctx, req.MerchantID, req.ItemID, grabDto.MenuItemUpdateFieldItem, false, err.Error())
        g.Log().Errorf(ctx, "[Grab] UpdateMenuItem API failed: merchantID=%s, itemID=%s, error=%v",
            req.MerchantID, req.ItemID, err)
        return gerror.Wrap(err, "调用 Grab UpdateMenuItem API 失败")
    }

    // 4. 记录成功日志
    s.logMenuRecordUpdate(ctx, req.MerchantID, req.ItemID, grabDto.MenuItemUpdateFieldItem, true, "")
    g.Log().Infof(ctx, "[Grab] UpdateMenuItem success: merchantID=%s, itemID=%s", req.MerchantID, req.ItemID)
    return nil
}
```

#### UpdateMenuModifier 简化后

```go
// UpdateMenuModifier 更新单个修饰符
func (s *sGrabMenu) UpdateMenuModifier(ctx context.Context, req *grabDto.UpdateMenuModifierReq) error {
    g.Log().Infof(ctx, "[Grab] UpdateMenuModifier: merchantID=%s, modifierID=%s, modifierName=%s",
        req.MerchantID, req.ModifierID, req.ModifierName)

    // 1. 参数验证
    if err := g.Validator().Data(req).Run(ctx); err != nil {
        g.Log().Errorf(ctx, "[Grab] UpdateMenuModifier validation failed: %v", err)
        return gerror.NewCode(gcode.CodeValidationFailed, err.Error())
    }

    // 2. 构建 SDK 请求
    updateModifier := req.ToSDKUpdateMenuModifier()
    updateReq := grabfood.UpdateMenuModifierAsUpdateMenuRequest(updateModifier)

    // 3. 调用 Grab API
    if err := service.Grab().UpdateMenuRecord(ctx, req.MerchantID, updateReq); err != nil {
        s.logMenuRecordUpdate(ctx, req.MerchantID, req.ModifierID, grabDto.MenuItemUpdateFieldModifier, false, err.Error())
        g.Log().Errorf(ctx, "[Grab] UpdateMenuModifier API failed: merchantID=%s, modifierID=%s, error=%v",
            req.MerchantID, req.ModifierID, err)
        return gerror.Wrap(err, "调用 Grab UpdateMenuModifier API 失败")
    }

    // 4. 记录成功日志
    s.logMenuRecordUpdate(ctx, req.MerchantID, req.ModifierID, grabDto.MenuItemUpdateFieldModifier, true, "")
    g.Log().Infof(ctx, "[Grab] UpdateMenuModifier success: merchantID=%s, modifierID=%s", req.MerchantID, req.ModifierID)
    return nil
}
```

### DTO 层变更

删除 `menu_update.go` 中的 `UpdateMenuResult` 结构体定义。

### Service 接口变更

更新 `grab_menu.go` 接口定义：

```go
type IGrabMenu interface {
    // ... 其他方法 ...
    UpdateMenuItem(ctx context.Context, req *grabDto.UpdateMenuItemReq) error
    UpdateMenuModifier(ctx context.Context, req *grabDto.UpdateMenuModifierReq) error
}
```

---

## 📚 实现清单

### Phase 1: 代码修改

- [ ] 删除 `UpdateMenuResult` 结构体
- [ ] 修改 `UpdateMenuItem` 方法
- [ ] 修改 `UpdateMenuModifier` 方法
- [ ] 更新 service interface

### Phase 2: 验证

- [ ] 编译通过
- [ ] 无 lint 错误

**详细任务**: 参见 `tasks.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18  
**作者**: AI Agent

