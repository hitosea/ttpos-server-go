package erp

import (
	"context"
	"fmt"
	"testing"

	"ttpos-bmp/app/ttpos-erp/api/buying"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestGetPurchaseOrder_Real 真实连接 ERP 服务测试获取采购单
// 直接连接 gRPC 地址，绕过 Nacos 服务发现
func TestGetPurchaseOrder_Real(t *testing.T) {
	// 直接连接 ERP gRPC 服务（请根据实际地址修改）
	addr := "10.180.10.13:14022" // ttpos-erp gRPC 端口

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("连接 ERP 服务失败: %v", err)
	}
	defer conn.Close()

	client := buying.NewBuyingServiceClient(conn)

	// 设置 site code
	ctx := WithSiteCode(context.Background(), "2")

	// 调用获取采购单
	result, err := client.GetPurchaseOrder(ctx, &buying.GetPurchaseOrderReq{
		PurchaseOrderName: "PUR-ORD-2026-00017",
	})
	if err != nil {
		t.Fatalf("调用 GetPurchaseOrder 失败: %v", err)
	}

	fmt.Printf("响应 Code: %s, Message: %s\n", result.Code, result.Message)

	if result.Code != "0" {
		t.Fatalf("业务错误: %s", result.Message)
	}

	if result.Data != nil {
		var resp buying.GetPurchaseOrderResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			t.Fatalf("解析响应数据失败: %v", err)
		}

		if resp.PurchaseOrder != nil {
			po := resp.PurchaseOrder
			fmt.Printf("采购单: %s, 供应商: %s, 已收货: %.0f%%, 物品数: %d\n",
				po.PurchaseOrderName, po.SupplierName, po.PerReceived, len(po.Items))

			for i, item := range po.Items {
				fmt.Printf("  %d. %s (%s) x %.2f\n", i+1, item.ItemName, item.ItemCode, item.Qty)
			}
		}
	}
}

// TestBatchGetPurchaseOrderPerReceived_Real 真实连接 ERP 服务测试批量获取采购单收货进度
func TestBatchGetPurchaseOrderPerReceived_Real(t *testing.T) {
	addr := "10.180.10.13:14022"

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("连接 ERP 服务失败: %v", err)
	}
	defer conn.Close()

	client := buying.NewBuyingServiceClient(conn)
	ctx := WithSiteCode(context.Background(), "2")

	// 批量查询采购单收货进度
	result, err := client.GetPurchaseOrderList(ctx, &buying.GetPurchaseOrderListReq{
		Name: "PUR-ORD-2026-00017,PUR-ORD-2026-00016,PUR-ORD-2026-00015",
	})
	if err != nil {
		t.Fatalf("调用 GetPurchaseOrderList 失败: %v", err)
	}

	fmt.Printf("响应 Code: %s, Message: %s\n", result.Code, result.Message)

	if result.Code != "0" {
		t.Fatalf("业务错误: %s", result.Message)
	}

	if result.Data != nil {
		var resp buying.GetPurchaseOrderListResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			t.Fatalf("解析响应数据失败: %v", err)
		}

		fmt.Printf("查询到 %d 个采购单收货进度:\n", len(resp.PurchaseOrders))
		for _, item := range resp.PurchaseOrders {
			fmt.Printf("  采购单: %s, 已收货: %.2f%%\n", item.Name, item.PerReceived)
		}
	}
}
