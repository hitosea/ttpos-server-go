package grab_menu

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/test/gtest"
)

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

// setupTestConfig 设置测试配置
func setupTestConfig(content string) {
	adapter, _ := gcfg.NewAdapterContent(content)
	g.Cfg().SetAdapter(adapter)
}

// ============================================================================
// fetchMenuFromTTpos 单元测试 (Mock HTTP)
// ============================================================================
// 注意：GenerateTtposAuth 的单元测试已移至 utility/ttpos_auth_test.go

func TestFetchMenuFromTTpos_Success(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 构造 Mock 服务器返回的菜单数据
		merchantID := "M-12345"
		partnerMerchantID := "12345"
		menuData := map[string]interface{}{
			"merchantID":        merchantID,
			"partnerMerchantID": partnerMerchantID,
			"currency":          map[string]string{"code": "THB"},
		}

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 验证请求头
			t.Assert(r.Header.Get("X-TTPOS-SECRET") != "", true)
			t.Assert(r.Header.Get("Content-Type"), "application/json")

			// 验证请求路径
			t.Assert(r.URL.Path, "/api/v1/takeout/menu/export")

			// 返回成功响应
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			resp := map[string]interface{}{
				"code":    200,
				"message": "success",
				"data": map[string]interface{}{
					"platform": "grab",
					"menuData": menuData,
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer mockServer.Close()

		// 设置配置
		setupTestConfig(`
app:
  ttposEndpoint: "` + mockServer.URL + `"
  callbackSecret: "test-secret"
`)

		// 测试
		ctx := context.Background()
		svc := &sGrabMenu{}
		result, err := svc.fetchMenuFromTTpos(ctx, 12345)

		// 验证
		t.AssertNil(err)
		t.AssertNE(result, nil)
		// 验证返回的是 *grabfood.GetMenuNewResponse 类型
		t.Assert(*result.MerchantID, merchantID)
		t.Assert(*result.PartnerMerchantID, partnerMerchantID)
	})
}

func TestFetchMenuFromTTpos_Non200Status(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// Mock 服务器返回 500 错误
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal Server Error"))
		}))
		defer mockServer.Close()

		// 设置配置
		setupTestConfig(`
app:
  ttposEndpoint: "` + mockServer.URL + `"
  callbackSecret: "test-secret"
`)

		// 测试
		ctx := context.Background()
		svc := &sGrabMenu{}
		result, err := svc.fetchMenuFromTTpos(ctx, 12345)

		// 验证：应该返回错误
		t.AssertNE(err, nil)
		t.Assert(result, nil)
	})
}

func TestFetchMenuFromTTpos_MissingConfig(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 设置空配置（缺少 ttposEndpoint）
		setupTestConfig(`
app:
  callbackSecret: "test-secret"
`)

		// 测试
		ctx := context.Background()
		svc := &sGrabMenu{}
		result, err := svc.fetchMenuFromTTpos(ctx, 12345)

		// 验证：应该返回配置缺失错误
		t.AssertNE(err, nil)
		t.Assert(result, nil)
	})
}

func TestFetchMenuFromTTpos_APIBusinessError(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// Mock 服务器返回业务错误
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			resp := map[string]interface{}{
				"code":    400,
				"message": "menu not found",
				"data":    nil,
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer mockServer.Close()

		// 设置配置
		setupTestConfig(`
app:
  ttposEndpoint: "` + mockServer.URL + `"
  callbackSecret: "test-secret"
`)

		// 测试
		ctx := context.Background()
		svc := &sGrabMenu{}
		result, err := svc.fetchMenuFromTTpos(ctx, 12345)

		// 验证：应该返回业务错误
		t.AssertNE(err, nil)
		t.Assert(result, nil)
	})
}

// ============================================================================
// HandleGetMenu 集成测试说明
// ============================================================================

// TestHandleGetMenu_LocalSnapshotExists 和 TestHandleGetMenu_FallbackToTTpos
// 需要 Mock 数据库和 service.ChannelMenu() 服务
// 由于需要完整的 GoFrame 服务注册环境，建议在集成测试环境中运行
//
// 集成测试用例：
// 1. 本地快照存在 → 直接返回
// 2. 本地快照为空 → 回退调用 TTPOS 成功
// 3. 本地快照为空 → 回退调用 TTPOS 失败 → 返回 CodeNotFound
// 4. partnerMerchantID 无效 → 返回 CodeInvalidParameter
