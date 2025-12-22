package model

// TakeoutOrderReceiver 外卖订单收货人信息表
// 存储外卖订单的收货人信息，包括姓名、电话、地址等
type TakeoutOrderReceiver struct {
	BaseModel
	TakeoutOrderUuid uint64 `gorm:"column:takeout_order_uuid;uniqueIndex:idx_takeout_order_uuid;not null;comment:外卖订单UUID" json:"takeout_order_uuid"`
	Platform         string `gorm:"column:platform;size:50;not null;comment:平台名称(grab/lineman/foodpanda等)" json:"platform"`

	// 收货人基本信息
	ReceiverName   string `gorm:"column:receiver_name;size:100;comment:收货人姓名" json:"receiver_name"`
	ReceiverPhones string `gorm:"column:receiver_phones;size:50;comment:收货人电话" json:"receiver_phones"`

	// 地址信息
	UnitNumber          string `gorm:"column:unit_number;size:50;comment:单元号/门牌号" json:"unit_number"`
	DeliveryInstruction string `gorm:"column:delivery_instruction;size:500;comment:配送说明" json:"delivery_instruction"`
	PoiSource           string `gorm:"column:poi_source;size:50;comment:POI来源(GRAB/GOOGLE/FACEBOOK等)" json:"poi_source"`
	PoiID               string `gorm:"column:poi_id;size:100;comment:POI ID" json:"poi_id"`
	Address             string `gorm:"column:address;size:500;comment:完整地址" json:"address"`
	Postcode            string `gorm:"column:postcode;size:20;comment:邮政编码" json:"postcode"`

	// 坐标信息
	Latitude  float64 `gorm:"column:latitude;type:decimal(10,7);comment:纬度" json:"latitude"`
	Longitude float64 `gorm:"column:longitude;type:decimal(10,7);comment:经度" json:"longitude"`
}

func (*TakeoutOrderReceiver) TableName() string {
	return "ttpos_takeout_order_receiver"
}
