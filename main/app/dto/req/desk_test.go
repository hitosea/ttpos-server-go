package req

import (
	"testing"
)

func TestCreateMemberDineInOrderReq_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateMemberDineInOrderReq
		wantErr bool
		errMsg  string
	}{
		{
			name:    "空商品列表_应报错",
			req:     CreateMemberDineInOrderReq{Products: nil},
			wantErr: true,
			errMsg:  "未选购商品",
		},
		{
			name: "空切片商品列表_应报错",
			req:  CreateMemberDineInOrderReq{Products: []OrderProductAddReq{}},
			wantErr: true,
			errMsg:  "未选购商品",
		},
		{
			name: "商品数量为0_应报错",
			req: CreateMemberDineInOrderReq{
				Products: []OrderProductAddReq{
					{FlavorUuid: 1, Num: 0, ProductType: 0},
				},
			},
			wantErr: true,
			errMsg:  "商品数量不能为0",
		},
		{
			name: "商品数量为负数_应报错",
			req: CreateMemberDineInOrderReq{
				Products: []OrderProductAddReq{
					{FlavorUuid: 1, Num: -1, ProductType: 0},
				},
			},
			wantErr: true,
			errMsg:  "商品数量不能为0",
		},
		{
			name: "商品规格ID为0_应报错",
			req: CreateMemberDineInOrderReq{
				Products: []OrderProductAddReq{
					{FlavorUuid: 0, Num: 1, ProductType: 0},
				},
			},
			wantErr: true,
			errMsg:  "商品规格ID不能为0",
		},
		{
			name: "普通商品_参数正确_应通过",
			req: CreateMemberDineInOrderReq{
				Products: []OrderProductAddReq{
					{FlavorUuid: 100, Num: 2, ProductType: 0},
				},
			},
			wantErr: false,
		},
		{
			name: "多个普通商品_参数正确_应通过",
			req: CreateMemberDineInOrderReq{
				Products: []OrderProductAddReq{
					{FlavorUuid: 100, Num: 1, ProductType: 0},
					{FlavorUuid: 200, Num: 3, ProductType: 0},
				},
			},
			wantErr: false,
		},
		{
			name: "套餐_子商品为空_应报错",
			req: CreateMemberDineInOrderReq{
				Products: []OrderProductAddReq{
					{FlavorUuid: 100, Num: 1, ProductType: 1, Products: nil},
				},
			},
			wantErr: true,
			errMsg:  "套餐子商品不能为空",
		},
		{
			name: "套餐_子商品规格ID为0_应报错",
			req: CreateMemberDineInOrderReq{
				Products: []OrderProductAddReq{
					{FlavorUuid: 100, Num: 1, ProductType: 1, Products: []ProductRequest{
						{EditProductReq: EditProductReq{FlavorUuid: 0}, Num: 1, ProductPackageGroupUuid: 10},
					}},
				},
			},
			wantErr: true,
			errMsg:  "套餐子商品规格ID不能为0",
		},
		{
			name: "套餐_子商品数量为0_应报错",
			req: CreateMemberDineInOrderReq{
				Products: []OrderProductAddReq{
					{FlavorUuid: 100, Num: 1, ProductType: 1, Products: []ProductRequest{
						{EditProductReq: EditProductReq{FlavorUuid: 50}, Num: 0, ProductPackageGroupUuid: 10},
					}},
				},
			},
			wantErr: true,
			errMsg:  "套餐子商品数量不能为0",
		},
		{
			name: "套餐_子商品分组ID为0_应报错",
			req: CreateMemberDineInOrderReq{
				Products: []OrderProductAddReq{
					{FlavorUuid: 100, Num: 1, ProductType: 1, Products: []ProductRequest{
						{EditProductReq: EditProductReq{FlavorUuid: 50}, Num: 1, ProductPackageGroupUuid: 0},
					}},
				},
			},
			wantErr: true,
			errMsg:  "套餐子商品分组ID不能为0",
		},
		{
			name: "套餐_参数正确_应通过",
			req: CreateMemberDineInOrderReq{
				Products: []OrderProductAddReq{
					{FlavorUuid: 100, Num: 1, ProductType: 1, Products: []ProductRequest{
						{EditProductReq: EditProductReq{FlavorUuid: 50}, Num: 1, ProductPackageGroupUuid: 10},
						{EditProductReq: EditProductReq{FlavorUuid: 60}, Num: 2, ProductPackageGroupUuid: 20},
					}},
				},
			},
			wantErr: false,
		},
		{
			name: "混合_普通商品和套餐_参数正确_应通过",
			req: CreateMemberDineInOrderReq{
				SaleBillUuid: 0,
				Products: []OrderProductAddReq{
					{FlavorUuid: 100, Num: 1, ProductType: 0},
					{FlavorUuid: 200, Num: 1, ProductType: 1, Products: []ProductRequest{
						{EditProductReq: EditProductReq{FlavorUuid: 50}, Num: 1, ProductPackageGroupUuid: 10},
					}},
				},
			},
			wantErr: false,
		},
		{
			name: "第二个商品无效_应报错",
			req: CreateMemberDineInOrderReq{
				Products: []OrderProductAddReq{
					{FlavorUuid: 100, Num: 1, ProductType: 0},
					{FlavorUuid: 0, Num: 1, ProductType: 0},
				},
			},
			wantErr: true,
			errMsg:  "商品规格ID不能为0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Validate() = nil, want error")
					return
				}
				if tt.errMsg != "" && !containsStr(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %q, want containing %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
			}
		})
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
