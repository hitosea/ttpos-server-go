package utils

import (
	"encoding/json"
	"fmt"
	"ttpos-server-go/app/errors"
)

type Material struct {
	ID              int             `json:"id"`
	FeedID          int             `json:"feed_id"`
	ProductFeedID   int             `json:"product_feed_id"`
	MaterialID      int             `json:"material_id"`
	MaterialNum     float64         `json:"material_num"`
	ShopSupplierID  int             `json:"shop_supplier_id"`
	AppID           int             `json:"app_id"`
	CreateTime      string          `json:"create_time"`
	MaterialProduct MaterialProduct `json:"materialProduct"`
	ProductID       int             `json:"product_id"`
}

type ProductFeed struct {
	FeedNameText   string     `json:"feed_name_text"`
	ProductFeedID  int        `json:"product_feed_id"`
	ProductID      int        `json:"product_id"`
	FeedID         int        `json:"feed_id"`
	FeedName       string     `json:"feed_name"`
	Price          any        `json:"price"`
	StockNum       int        `json:"stock_num"`
	DefaultSelect  any        `json:"default_select"`
	ShopSupplierID int        `json:"shop_supplier_id"`
	Sort           int        `json:"sort"`
	AppID          int        `json:"app_id"`
	Material       []Material `json:"material"`
	UUID           string     `json:"uuid"`
}

// GetPrice 获取商品规格价格.
// 注意：Price 类型为 any，需要先转换为字符串，再转换为 float64
// price的原始数据有可能是 float64 类型，也有可能是 string 类型
func (p *ProductFeed) GetPrice() (float64, error) {
	price, err := ConvertToFloat64(p.Price)
	if err != nil {
		return 0, errors.WithMessage(err, fmt.Sprintf("ConvertToFloat64 failed, price: %v", p.Price))
	}
	return price, nil
}

func (p *ProductFeed) GetDefaultSelect() (int, error) {
	defaultSelect, err := ConvertToFloat64(p.DefaultSelect)
	if err != nil {
		return 0, errors.WithMessage(err, fmt.Sprintf("ConvertToFloat64 failed, defaultSelect: %v", p.DefaultSelect))
	}
	return int(defaultSelect), nil
}

type MaterialProduct struct {
	ProductSales         int     `json:"product_sales"`
	ProductNameText      string  `json:"product_name_text"`
	ProductUnitText      string  `json:"product_unit_text"`
	ProductID            int     `json:"product_id"`
	ProductName          string  `json:"product_name"`
	ProductUnit          string  `json:"product_unit"`
	ProductMaterialStock float64 `json:"product_material_stock"`
	Sku                  []struct {
		MaterialStock float64 `json:"material_stock"`
		ProductID     int     `json:"product_id"`
	} `json:"sku"`
}

func ParseFeedJson(jsonString string) ([]ProductFeed, error) {
	var productFeeds []ProductFeed
	err := json.Unmarshal([]byte(jsonString), &productFeeds)
	if err != nil {
		return nil, errors.WithMessage(err, "解析商品规格失败")
	}
	return productFeeds, nil
}
