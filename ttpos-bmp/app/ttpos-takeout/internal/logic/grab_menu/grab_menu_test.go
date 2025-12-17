package grab_menu

// 注意：由于 GoFrame 的服务注册机制，logic 层的测试需要完整的配置环境。
// DTO 层的单元测试在 model/dto/grab/menu_update_test.go 中，可以独立运行。
//
// 如需测试 logic 层，请确保：
// 1. 配置文件存在于 manifest/config/ 目录
// 2. 数据库和 Redis 连接正常
// 3. RocketMQ 队列配置正确
//
// 集成测试示例：
//
// func TestUpdateMenuItem_Integration(t *testing.T) {
//     // 需要完整的环境配置
//     svc := New()
//     req := &grabDto.UpdateMenuItemReq{
//         MerchantID: "M-12345",
//         ItemID:     "ITEM-001",
//         Price:      ptrInt64(1000),
//     }
//     result, err := svc.UpdateMenuItem(context.Background(), req)
//     // 验证结果
// }
