package lineman

import (
	"context"
	"fmt"
	"testing"

	lineman_dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

// TestMapStatusToLineman 测试状态映射函数
func TestMapStatusToLineman(t *testing.T) {
	tests := []struct {
		name        string
		ttposStatus string
		wantStatus  string
		wantErr     bool
	}{
		{
			name:        "AVAILABLE 映射为 AVAILABLE",
			ttposStatus: "AVAILABLE",
			wantStatus:  "AVAILABLE",
			wantErr:     false,
		},
		{
			name:        "UNAVAILABLE 映射为 SUSPENDED",
			ttposStatus: "UNAVAILABLE",
			wantStatus:  "SUSPENDED",
			wantErr:     false,
		},
		{
			name:        "SOLD_OUT_TODAY 映射为 SOLD_OUT_TODAY",
			ttposStatus: "SOLD_OUT_TODAY",
			wantStatus:  "SOLD_OUT_TODAY",
			wantErr:     false,
		},
		{
			name:        "UNAVAILABLEHIDE 不支持，返回错误",
			ttposStatus: "UNAVAILABLEHIDE",
			wantStatus:  "",
			wantErr:     true,
		},
		{
			name:        "空状态返回错误",
			ttposStatus: "",
			wantStatus:  "",
			wantErr:     true,
		},
		{
			name:        "未知状态返回错误",
			ttposStatus: "UNKNOWN_STATUS",
			wantStatus:  "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, err := MapStatusToLineman(tt.ttposStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapStatusToLineman() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("MapStatusToLineman() gotStatus = %v, want %v", gotStatus, tt.wantStatus)
			}
		})
	}
}

// MockMenuStatusClient Mock 客户端（用于测试）
type MockMenuStatusClient struct {
	UpdateFunc func(ctx context.Context, storeId string, req *lineman_dto.MenuStatusUpdateReq) (*lineman_dto.MenuStatusUpdateResp, error)
}

func (m *MockMenuStatusClient) UpdateMenuStatusWithRetry(ctx context.Context, storeId string, req *lineman_dto.MenuStatusUpdateReq) (*lineman_dto.MenuStatusUpdateResp, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, storeId, req)
	}
	return &lineman_dto.MenuStatusUpdateResp{
		Status:  "ok",
		Code:    "SUCCESS",
		Message: "Success",
	}, nil
}

// TestUpdateMenuStatus_ParamValidation 测试参数校验
func TestUpdateMenuStatus_ParamValidation(t *testing.T) {
	mockClient := &MockMenuStatusClient{}
	logic := NewMenuStatusLogic(mockClient)

	tests := []struct {
		name    string
		storeId string
		req     *lineman_dto.MenuStatusUpdateReq
		wantErr bool
	}{
		{
			name:    "storeId 为空，返回错误",
			storeId: "",
			req: &lineman_dto.MenuStatusUpdateReq{
				MenuItems: []lineman_dto.MenuItemStatus{
					{ID: "item-1", MenuStatus: "AVAILABLE"},
				},
			},
			wantErr: true,
		},
		{
			name:    "menuItems 为空，返回错误",
			storeId: "store-123",
			req: &lineman_dto.MenuStatusUpdateReq{
				MenuItems: []lineman_dto.MenuItemStatus{},
			},
			wantErr: true,
		},
		{
			name:    "menuItems 超过 100 个，返回错误",
			storeId: "store-123",
			req: func() *lineman_dto.MenuStatusUpdateReq {
				items := make([]lineman_dto.MenuItemStatus, 101)
				for i := 0; i < 101; i++ {
					items[i] = lineman_dto.MenuItemStatus{
						ID:         fmt.Sprintf("item-%d", i),
						MenuStatus: "AVAILABLE",
					}
				}
				return &lineman_dto.MenuStatusUpdateReq{MenuItems: items}
			}(),
			wantErr: true,
		},
		{
			name:    "menuItems 数量正常，不返回错误",
			storeId: "store-123",
			req: &lineman_dto.MenuStatusUpdateReq{
				MenuItems: []lineman_dto.MenuItemStatus{
					{
						ID:         "item-1",
						MenuStatus: "AVAILABLE",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logic.UpdateMenuStatus(nil, tt.storeId, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateMenuStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
