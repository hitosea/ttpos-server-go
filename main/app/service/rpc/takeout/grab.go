package takeout

import (
	"context"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// TODO: 需要在 bmp 的 ttpos-takeout 服务中实现以下 RPC 接口：
// 1. GetGrabBindingLink - 获取 Grab 绑定链接
// 2. CheckGrabBindingStatus - 检查 Grab 绑定状态
// 3. GetGrabMenu - 获取 Grab 商品菜单
//
// 需要在 ttpos-bmp/app/ttpos-takeout/manifest/protobuf/takeout/takeout.proto 中添加：
//
// message GetGrabBindingLinkReq {
//   uint64 companyUuid = 1;
// }
//
// message GetGrabBindingLinkResp {
//   ResponseInfo responseInfo = 1;
//   string bindingLink = 2;  // 绑定链接 URL
//   int64 expiresAt = 3;      // 过期时间（Unix 时间戳）
// }
//
// message CheckGrabBindingStatusReq {
//   uint64 companyUuid = 1;
// }
//
// message CheckGrabBindingStatusResp {
//   ResponseInfo responseInfo = 1;
//   bool isBound = 2;         // 是否已绑定
//   int64 boundAt = 3;       // 绑定时间（Unix 时间戳）
//   string merchantId = 4;   // Grab 商户 ID
//   string merchantName = 5; // Grab 商户名称
// }
//
// message GetGrabMenuReq {
//   uint64 companyUuid = 1;
// }
//
// message GetGrabMenuResp {
//   ResponseInfo responseInfo = 1;
//   string menuData = 2;      // Grab 菜单 JSON 数据
// }
//
// 并在 TakeoutService 中添加：
// rpc GetGrabBindingLink (GetGrabBindingLinkReq) returns (GetGrabBindingLinkResp) {}
// rpc CheckGrabBindingStatus (CheckGrabBindingStatusReq) returns (CheckGrabBindingStatusResp) {}
// rpc GetGrabMenu (GetGrabMenuReq) returns (GetGrabMenuResp) {}

// GetGrabBindingLink 获取 Grab 绑定链接
// TODO: 等待 bmp 实现 RPC 接口后，调用 takeout.NewTakeoutClient() 获取客户端并调用 RPC
func GetGrabBindingLink(ctx context.Context, companyUuid uint64) (bindingLink string, expiresAt int64, err error) {
	// TODO: 实现 RPC 调用
	// client, conn, err := takeout.NewTakeoutClient()
	// if err != nil {
	// 	return "", 0, errors.WithMessage(err, "创建外送服务gRPC客户端失败")
	// }
	// defer conn.Close()
	//
	// req := &api.GetGrabBindingLinkReq{
	// 	CompanyUuid: companyUuid,
	// }
	// resp, err := client.GetGrabBindingLink(ctx, req)
	// if err != nil {
	// 	return "", 0, errors.WithMessage(err, "调用获取绑定链接接口失败")
	// }
	// if resp.ResponseInfo.Code != "0" {
	// 	return "", 0, errors.New(resp.ResponseInfo.Message)
	// }
	// return resp.BindingLink, resp.ExpiresAt, nil

	logger.Logger.Warn("GetGrabBindingLink 尚未实现，等待 bmp RPC 接口", zap.Uint64("companyUuid", companyUuid))
	return "", 0, errors.New("GetGrabBindingLink 接口尚未实现，等待 bmp 提供 RPC 接口")
}

// CheckGrabBindingStatus 检查 Grab 绑定状态
// TODO: 等待 bmp 实现 RPC 接口后，调用 takeout.NewTakeoutClient() 获取客户端并调用 RPC
func CheckGrabBindingStatus(ctx context.Context, companyUuid uint64) (isBound bool, boundAt int64, merchantID string, merchantName string, err error) {
	// TODO: 实现 RPC 调用
	// client, conn, err := takeout.NewTakeoutClient()
	// if err != nil {
	// 	return false, 0, "", "", errors.WithMessage(err, "创建外送服务gRPC客户端失败")
	// }
	// defer conn.Close()
	//
	// req := &api.CheckGrabBindingStatusReq{
	// 	CompanyUuid: companyUuid,
	// }
	// resp, err := client.CheckGrabBindingStatus(ctx, req)
	// if err != nil {
	// 	return false, 0, "", "", errors.WithMessage(err, "调用检查绑定状态接口失败")
	// }
	// if resp.ResponseInfo.Code != "0" {
	// 	return false, 0, "", "", errors.New(resp.ResponseInfo.Message)
	// }
	// return resp.IsBound, resp.BoundAt, resp.MerchantId, resp.MerchantName, nil

	logger.Logger.Warn("CheckGrabBindingStatus 尚未实现，等待 bmp RPC 接口", zap.Uint64("companyUuid", companyUuid))
	return false, 0, "", "", errors.New("CheckGrabBindingStatus 接口尚未实现，等待 bmp 提供 RPC 接口")
}

// GetGrabMenu 获取 Grab 商品菜单
// TODO: 等待 bmp 实现 RPC 接口后，调用 takeout.NewTakeoutClient() 获取客户端并调用 RPC
func GetGrabMenu(ctx context.Context, companyUuid uint64) (menuData interface{}, err error) {
	// TODO: 实现 RPC 调用
	// client, conn, err := takeout.NewTakeoutClient()
	// if err != nil {
	// 	return nil, errors.WithMessage(err, "创建外送服务gRPC客户端失败")
	// }
	// defer conn.Close()
	//
	// req := &api.GetGrabMenuReq{
	// 	CompanyUuid: companyUuid,
	// }
	// resp, err := client.GetGrabMenu(ctx, req)
	// if err != nil {
	// 	return nil, errors.WithMessage(err, "调用获取 Grab 菜单接口失败")
	// }
	// if resp.ResponseInfo.Code != "0" {
	// 	return nil, errors.New(resp.ResponseInfo.Message)
	// }
	//
	// // 解析 JSON 数据
	// var menu interface{}
	// if err := json.Unmarshal([]byte(resp.MenuData), &menu); err != nil {
	// 	return nil, errors.WithMessage(err, "解析 Grab 菜单数据失败")
	// }
	// return menu, nil

	logger.Logger.Warn("GetGrabMenu 尚未实现，等待 bmp RPC 接口", zap.Uint64("companyUuid", companyUuid))
	return nil, errors.New("GetGrabMenu 接口尚未实现，等待 bmp 提供 RPC 接口")
}
