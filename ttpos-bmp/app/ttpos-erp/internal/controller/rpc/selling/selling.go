package selling

import (
	"context"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Controller 销售服务控制器
type Controller struct {
	selling.UnsafeSellingServiceServer
}

// Register 注册销售服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	selling.RegisterSellingServiceServer(s.Server, &Controller{})
}

// GetPosProfileList 获取POS配置文件列表
// 参数：
//   - ctx: 上下文对象
//   - req: 获取POS配置文件列表请求参数
//
// 返回：
//   - *api.ResponseInfo: 响应信息
//   - error: 错误信息
func (c *Controller) GetPosProfileList(ctx context.Context, req *selling.PosProfileReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validatePosProfileReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层获取数据
	dataList, err := service.Selling().GetPosProfileList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取POS配置文件列表成功", dataList), nil
}

// validatePosProfileReq 验证获取POS配置文件列表请求参数
func (c *Controller) validatePosProfileReq(req *selling.PosProfileReq) error {
	if strings.TrimSpace(req.CompanyAbbr) == "" {
		return gerror.New("公司简称不能为空")
	}
	return nil
}

// CreatePaymentAccount 创建支付账户
// 参数：
//   - ctx: 上下文对象
//   - req: 创建支付账户请求参数
//
// 返回：
//   - *api.ResponseInfo: 响应信息
//   - error: 错误信息
func (c *Controller) CreatePaymentAccount(ctx context.Context, req *selling.CreatePaymentAccountReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateCreatePaymentAccountReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 处理连连支付前缀
	paymentType := req.PaymentType
	if req.PaymentSource == "2" {
		paymentType = consts.LianlianPayPrefix + paymentType
	}

	// 调用服务层创建支付账户
	err := service.Selling().CreateModePaymentAccount(ctx, &setup.CreateModePaymentAccountInp{
		CompanyAbbr: req.CompanyAbbr,
		PaymentType: paymentType,
	})
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccess("创建支付账户成功"), nil
}

// validateCreatePaymentAccountReq 验证创建支付账户请求参数
func (c *Controller) validateCreatePaymentAccountReq(req *selling.CreatePaymentAccountReq) error {
	if strings.TrimSpace(req.CompanyAbbr) == "" {
		return gerror.New("公司简称不能为空")
	}
	if strings.TrimSpace(req.PaymentType) == "" {
		return gerror.New("支付类型不能为空")
	}
	return nil
}

// OpenPosEntry 创建开帐
// 参数：
//   - ctx: 上下文对象
//   - req: 开帐请求参数
//
// 返回：
//   - *api.ResponseInfo: 响应信息
//   - error: 错误信息
func (c *Controller) OpenPosEntry(ctx context.Context, req *selling.OpenPosEntryReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateOpenPosEntryReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层创建开帐
	resp, err := service.Selling().OpenPosEntry(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	return rpc.ApiSuccessWithData("创建开帐成功", resp), nil
}

// validateOpenPosEntryReq 验证开帐请求参数
func (c *Controller) validateOpenPosEntryReq(req *selling.OpenPosEntryReq) error {
	if strings.TrimSpace(req.PosProfileName) == "" {
		return gerror.New("POS Profile名称不能为空")
	}
	if strings.TrimSpace(req.CashierEmail) == "" {
		return gerror.New("收银员邮箱不能为空")
	}
	if strings.TrimSpace(req.CompanyAbbr) == "" {
		return gerror.New("公司缩写不能为空")
	}
	if req.PeriodStartDate <= 0 {
		return gerror.New("周期开始时间不能为空")
	}
	if len(req.OpenPosEntryDetail) == 0 {
		return gerror.New("开帐详情不能为空")
	}

	// 验证开帐详情
	for i, detail := range req.OpenPosEntryDetail {
		if strings.TrimSpace(detail.ModeOfPayment) == "" {
			return gerror.Newf("第%d项支付方式不能为空", i+1)
		}
		if detail.OpeningAmount < 0 {
			return gerror.Newf("第%d项开帐金额不能为负数", i+1)
		}
	}

	return nil
}

// ClosePosEntry 创建关帐
// 参数：
//   - ctx: 上下文对象
//   - req: 关帐请求参数
//
// 返回：
//   - *api.ResponseInfo: 响应信息
//   - error: 错误信息
func (c *Controller) ClosePosEntry(ctx context.Context, req *selling.ClosePosEntryReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateClosePosEntryReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层创建关帐
	resp, err := service.Selling().ClosePosEntry(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	return rpc.ApiSuccessWithData("创建关帐成功", resp), nil
}

// validateClosePosEntryReq 验证关帐请求参数
func (c *Controller) validateClosePosEntryReq(req *selling.ClosePosEntryReq) error {
	if strings.TrimSpace(req.PosOpenEntryName) == "" {
		return gerror.New("开帐名称不能为空")
	}
	if req.PeriodEndDate <= 0 {
		return gerror.New("结账时间不能为空")
	}
	if len(req.ClosePosEntryDetail) == 0 {
		return gerror.New("关帐详情不能为空")
	}

	// 验证关帐详情
	for i, detail := range req.ClosePosEntryDetail {
		if strings.TrimSpace(detail.ModeOfPayment) == "" {
			return gerror.Newf("第%d项支付方式不能为空", i+1)
		}
		if detail.OpeningAmount < 0 {
			return gerror.Newf("第%d项开帐金额不能为负数", i+1)
		}
		if detail.ClosingAmount < 0 {
			return gerror.Newf("第%d项关帐金额不能为负数", i+1)
		}
	}

	return nil
}

// SavePosInvoice 保存POS发票
// 参数：
//   - ctx: 上下文对象
//   - req: 保存发票请求参数
//
// 返回：
//   - *api.ResponseInfo: 响应信息
//   - error: 错误信息
func (c *Controller) SavePosInvoice(ctx context.Context, req *selling.SavePosInvoiceReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateSavePosInvoiceReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层保存发票
	resp, err := service.Selling().SavePosInvoice(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	return rpc.ApiSuccessWithData("保存发票成功", resp), nil
}

// validateSavePosInvoiceReq 验证保存发票请求参数
func (c *Controller) validateSavePosInvoiceReq(req *selling.SavePosInvoiceReq) error {
	if strings.TrimSpace(req.OrderNo) == "" {
		return gerror.New("订单号不能为空")
	}
	if strings.TrimSpace(req.OpenPosEntryName) == "" {
		return gerror.New("POS开帐名称不能为空")
	}
	if strings.TrimSpace(req.CompanyAbbr) == "" {
		return gerror.New("公司缩写不能为空")
	}
	if req.PostingDatetime <= 0 {
		return gerror.New("过账日期时间不能为空")
	}
	if req.UpdateStock < 0 || req.UpdateStock > 1 {
		return gerror.New("是否更新库存必须为0或1")
	}
	if strings.TrimSpace(req.Currency) == "" {
		return gerror.New("货币不能为空")
	}
	if strings.TrimSpace(req.PriceListCurrency) == "" {
		return gerror.New("价格表货币不能为空")
	}
	if len(req.Items) == 0 {
		return gerror.New("商品项目列表不能为空")
	}
	if len(req.MaterialItems) == 0 {
		return gerror.New("原材料项目列表不能为空")
	}
	if len(req.Taxes) == 0 {
		return gerror.New("税费列表不能为空")
	}
	if len(req.Payments) == 0 {
		return gerror.New("付款列表不能为空")
	}

	// 验证商品项目
	for i, item := range req.Items {
		if strings.TrimSpace(item.ItemCode) == "" {
			return gerror.Newf("第%d项商品编码不能为空", i+1)
		}
		if item.Qty <= 0 {
			return gerror.Newf("第%d项商品数量必须大于0", i+1)
		}
		if item.Rate < 0 {
			return gerror.Newf("第%d项商品单价不能为负数", i+1)
		}
		if item.Amount < 0 {
			return gerror.Newf("第%d项商品金额不能为负数", i+1)
		}
	}

	// 验证税费
	for i, tax := range req.Taxes {
		if tax.Rate < 0 {
			return gerror.Newf("第%d项税率不能为负数", i+1)
		}
		if tax.TaxAmount < 0 {
			return gerror.Newf("第%d项税费金额不能为负数", i+1)
		}
		if strings.TrimSpace(tax.Description) == "" {
			return gerror.Newf("第%d项税费描述不能为空", i+1)
		}
	}

	// 验证付款
	for i, payment := range req.Payments {
		if strings.TrimSpace(payment.ModeOfPayment) == "" {
			return gerror.Newf("第%d项支付方式不能为空", i+1)
		}
		if payment.Amount <= 0 {
			return gerror.Newf("第%d项付款金额必须大于0", i+1)
		}
	}

	return nil
}
