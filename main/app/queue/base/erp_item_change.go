package queue

import (
	"context"
	"encoding/json"
	"ttpos-server-go/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

// erpItemChangeHandler 处理商品变更
func ErpItemChangeHandler(ctx context.Context, msg *primitive.MessageExt) error {
	//TODO 处理商品变更
	logger.Logger.Info("处理商品变更", zap.String("msg_id", msg.MsgId))
	// 报文样例
	//{"event":"item-change","data":{"item_code":"TTT"},"docType":"","siteCode":"1"}
	// 解析报文
	var msgBody struct {
		Event string `json:"event"`
		Data  struct {
			ItemCode string `json:"item_code"`
		} `json:"data"`
		DocType  string `json:"docType"`
		SiteCode string `json:"siteCode"`
	}
	err := json.Unmarshal(msg.Body, &msgBody)
	if err != nil {
		logger.Logger.Error("解析商品变更报文失败", zap.Error(err))
		return err
	}
	// 处理商品变更
	if msgBody.Event == "item-change" {
		// 商品变更
		logger.Logger.Info("商品变更", zap.String("item_code", msgBody.Data.ItemCode))
	}
	return nil
}
