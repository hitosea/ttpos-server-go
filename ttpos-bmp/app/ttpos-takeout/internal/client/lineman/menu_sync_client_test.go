// Package lineman Lineman API 客户端
package lineman

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gogf/gf/v2/test/gtest"

	dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

// TestMenuSyncClient_SyncMenu_Success 测试成功场景
// 注意：此测试需要完整的集成环境（配置、Redis、Token 服务等）
func TestMenuSyncClient_SyncMenu_Success(t *testing.T) {
	t.Skip("需要完整的集成测试环境（配置、Redis、Token 服务），在集成测试中运行")
	gtest.C(t, func(t *gtest.T) {
		// Mock Lineman API Server
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 验证请求方法
			t.Assert(r.Method, "PUT")
			// 验证请求头
			t.Assert(r.Header.Get("Authorization"), "Bearer test_token_123")
			t.Assert(r.Header.Get("Content-Type"), "application/json")
			// 验证 URL 路径（partnerId 现在从配置读取，可能不是 partner123）
			t.AssertIN("/stores/", r.URL.Path)

			// 返回成功响应
			resp := dto.MenuSyncResponse{
				BaseResponse: dto.BaseResponse{
					Status: "ok",
					Code:   "SUCCESS",
				},
				MenuSyncRequestId: "req-123456",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}))
		defer mockServer.Close()

		// 创建 Client
		client := &MenuSyncClient{
			endpoint: mockServer.URL,
			timeout:  5 * time.Second,
		}

		// 准备测试数据
		menuData := &dto.MenuSyncRequest{
			MenuGroups: []*dto.MenuGroup{
				{
					ID:             "TTPOS-CAT-1",
					Name:           dto.NameTrans{Thai: "เครื่องดื่ม", English: "Beverages"},
					UseSellingTime: false,
					MenuItems:      []*dto.MenuItem{},
				},
			},
		}

		// 调用同步方法（新签名：只需要 storeId 和 menuData）
		resp, err := client.SyncMenu(context.Background(), "store456", menuData)

		// 断言
		t.AssertNil(err)
		t.AssertNE(resp, nil)
		t.Assert(resp.Status, "ok")
		t.Assert(resp.Code, "SUCCESS")
		t.Assert(resp.MenuSyncRequestId, "req-123456")
	})
}

// TestMenuSyncClient_SyncMenu_ApiError 测试 API 错误场景
// 注意：此测试需要完整的集成环境（配置、Redis、Token 服务等）
func TestMenuSyncClient_SyncMenu_ApiError(t *testing.T) {
	t.Skip("需要完整的集成测试环境（配置、Redis、Token 服务），在集成测试中运行")
	gtest.C(t, func(t *gtest.T) {
		// Mock Server 返回错误
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := dto.MenuSyncResponse{
				BaseResponse: dto.BaseResponse{
					Status:  "error",
					Code:    "INVALID_PARTNER_ID",
					Message: "Invalid partner ID",
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}))
		defer mockServer.Close()

		client := &MenuSyncClient{
			endpoint: mockServer.URL,
			timeout:  5 * time.Second,
		}

		menuData := &dto.MenuSyncRequest{MenuGroups: []*dto.MenuGroup{}}

		// 调用同步方法（新签名）
		resp, err := client.SyncMenu(context.Background(), "store", menuData)

		// 断言
		t.AssertNE(err, nil)
		t.Assert(resp, nil)
		t.AssertIN("INVALID_PARTNER_ID", err.Error())
	})
}

// TestMenuSyncClient_SyncMenu_HttpError 测试 HTTP 错误场景
// 注意：此测试需要完整的集成环境（配置、Redis、Token 服务等）
func TestMenuSyncClient_SyncMenu_HttpError(t *testing.T) {
	t.Skip("需要完整的集成测试环境（配置、Redis、Token 服务），在集成测试中运行")
	gtest.C(t, func(t *gtest.T) {
		// Mock Server 返回 4xx 错误
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))
		}))
		defer mockServer.Close()

		client := &MenuSyncClient{
			endpoint: mockServer.URL,
			timeout:  5 * time.Second,
		}

		menuData := &dto.MenuSyncRequest{MenuGroups: []*dto.MenuGroup{}}

		// 调用同步方法（新签名）
		resp, err := client.SyncMenu(context.Background(), "store", menuData)

		// 断言
		t.AssertNE(err, nil)
		t.Assert(resp, nil)
		t.AssertIN("400", err.Error())
	})
}

// TestMenuSyncClient_SyncMenu_NetworkError 测试网络错误场景
// 注意：此测试需要完整的集成环境（配置、Redis、Token 服务等）
func TestMenuSyncClient_SyncMenu_NetworkError(t *testing.T) {
	t.Skip("需要完整的集成测试环境（配置、Redis、Token 服务），在集成测试中运行")
	gtest.C(t, func(t *gtest.T) {
		// 使用不存在的 endpoint
		client := &MenuSyncClient{
			endpoint: "http://invalid-host-that-does-not-exist-12345.com",
			timeout:  1 * time.Second,
		}

		menuData := &dto.MenuSyncRequest{MenuGroups: []*dto.MenuGroup{}}

		// 调用同步方法（新签名）
		resp, err := client.SyncMenu(context.Background(), "store", menuData)

		// 断言
		t.AssertNE(err, nil)
		t.Assert(resp, nil)
	})
}

// TestMenuSyncClient_SyncMenuWithRetry_Success 测试带重试的成功场景
// 注意：此测试需要完整的集成环境（配置、Redis、Token 服务等）
func TestMenuSyncClient_SyncMenuWithRetry_Success(t *testing.T) {
	t.Skip("需要完整的集成测试环境（配置、Redis、Token 服务），在集成测试中运行")
	gtest.C(t, func(t *gtest.T) {
		// Mock Server
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := dto.MenuSyncResponse{
				BaseResponse: dto.BaseResponse{
					Status: "ok",
					Code:   "SUCCESS",
				},
				MenuSyncRequestId: "req-retry-123",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}))
		defer mockServer.Close()

		client := &MenuSyncClient{
			endpoint: mockServer.URL,
			timeout:  5 * time.Second,
		}

		menuData := &dto.MenuSyncRequest{MenuGroups: []*dto.MenuGroup{}}

		// 调用带重试的同步方法（新签名）
		resp, err := client.SyncMenuWithRetry(context.Background(), "store", menuData)

		// 断言
		t.AssertNil(err)
		t.AssertNE(resp, nil)
		t.Assert(resp.MenuSyncRequestId, "req-retry-123")
	})
}
