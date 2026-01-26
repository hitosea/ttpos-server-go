package delivery_note

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	dto "ttpos-bmp/app/ttpos-erp/api/delivery_note"
	pb "ttpos-bmp/app/ttpos-erp/api/delivery_note_pb"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

// Controller 送货单服务控制器
type Controller struct {
	pb.UnimplementedDeliveryNoteServiceServer
}

// Register 注册送货单服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	pb.RegisterDeliveryNoteServiceServer(s.Server, &Controller{})
}

// GetDeliveryNoteList 获取送货单列表
// 参数：ctx 上下文，req 查询条件
// 返回：送货单列表和操作结果
func (*Controller) GetDeliveryNoteList(ctx context.Context, req *pb.GetDeliveryNoteListReq) (*api.ResponseInfo, error) {
	// 参数验证
	if len(req.CompanyAbbr) == 0 {
		return rpc.ApiError("公司简称不能为空"), nil
	}

	// 转换请求参数
	logicReq := &dto.GetDeliveryNoteListReq{
		CompanyAbbr:  req.CompanyAbbr,
		Branch:       req.Branch,
		Customer:     req.Customer,
		Warehouse:    req.Warehouse,
		Status:       req.Status,
		FromDate:     req.FromDate,
		ToDate:       req.ToDate,
		Limit:        req.Limit,
		IncludeItems: req.IncludeItems,
		PoNo:         req.PoNo,
	}

	// 调用 Logic 层
	resp, err := service.DeliveryNote().GetDeliveryNoteList(ctx, logicReq)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换响应数据为 protobuf 格式
	pbResp := &pb.GetDeliveryNoteListResp{
		DeliveryNoteList: make([]*pb.DeliveryNote, 0, len(resp.DeliveryNoteList)),
	}

	for _, dn := range resp.DeliveryNoteList {
		pbDn := &pb.DeliveryNote{
			Name:             dn.Name,
			Company:          dn.Company,
			Customer:         dn.Customer,
			CustomerName:     dn.CustomerName,
			PostingDate:      dn.PostingDate,
			PostingTime:      dn.PostingTime,
			SetWarehouse:     dn.SetWarehouse,
			SellingPriceList: dn.SellingPriceList,
			TotalQty:         dn.TotalQty,
			GrandTotal:       dn.GrandTotal,
			Status:           dn.Status,
			Docstatus:        int32(dn.Docstatus),
		}

		// 转换明细项
		if len(dn.Items) > 0 {
			pbDn.Items = make([]*pb.DeliveryNoteItem, 0, len(dn.Items))
			for _, item := range dn.Items {
				pbItem := &pb.DeliveryNoteItem{
					ItemCode:          item.ItemCode,
					ItemName:          item.ItemName,
					Qty:               item.Qty,
					Uom:               item.Uom,
					Rate:              item.Rate,
					Amount:            item.Amount,
					Warehouse:         item.Warehouse,
					AgainstSalesOrder: item.AgainstSalesOrder,
				}
				pbDn.Items = append(pbDn.Items, pbItem)
			}
		}

		pbResp.DeliveryNoteList = append(pbResp.DeliveryNoteList, pbDn)
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取送货单列表成功", pbResp), nil
}
