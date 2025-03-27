package utils

import "encoding/json"

type Material struct {
	ID              int             `json:"id"`
	FeedID          int             `json:"feed_id"`
	ProductFeedID   int             `json:"product_feed_id"`
	MaterialID      int             `json:"material_id"`
	MaterialNum     int             `json:"material_num"`
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
	Price          int        `json:"price"`
	StockNum       int        `json:"stock_num"`
	DefaultSelect  int        `json:"default_select"`
	ShopSupplierID int        `json:"shop_supplier_id"`
	Sort           int        `json:"sort"`
	AppID          int        `json:"app_id"`
	Material       []Material `json:"material"`
	UUID           string     `json:"uuid"`
}

type MaterialProduct struct {
	ProductSales         int    `json:"product_sales"`
	ProductNameText      string `json:"product_name_text"`
	ProductUnitText      string `json:"product_unit_text"`
	ProductID            int    `json:"product_id"`
	ProductName          string `json:"product_name"`
	ProductUnit          string `json:"product_unit"`
	ProductMaterialStock int    `json:"product_material_stock"`
	Sku                  []struct {
		MaterialStock int `json:"material_stock"`
		ProductID     int `json:"product_id"`
	} `json:"sku"`
}

func ParseFeedJson(jsonString string) ([]ProductFeed, error) {
	var productFeeds []ProductFeed
	err := json.Unmarshal([]byte(jsonString), &productFeeds)
	if err != nil {
		return nil, err
	}
	return productFeeds, nil
}
