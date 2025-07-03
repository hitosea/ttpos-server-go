package model

type MemberAddress struct {
	BaseModel
	MemberUuid uint64 `gorm:"column:member_uuid;type:bigint(20);not null;comment:会员uuid" json:"member_uuid"`
	Name       string `gorm:"column:name;type:varchar(50);not null;comment:联系人" json:"name"`
	Phone      string `gorm:"column:phone;type:varchar(20);not null;comment:手机号" json:"phone"`
	Country    string `gorm:"column:country;type:varchar(10);not null;comment:国家代码" json:"country"`
	Province   string `gorm:"column:province;type:varchar(50);not null;comment:省份" json:"province"`
	City       string `gorm:"column:city;type:varchar(50);not null;comment:城市" json:"city"`
	Area       string `gorm:"column:area;type:varchar(50);not null;comment:区" json:"area"`
	Address    string `gorm:"column:address;type:varchar(255);not null;comment:详细地址" json:"address"`
	Street     string `gorm:"column:street;type:varchar(255);not null;comment:街道/门牌号" json:"street"`
	IsDefault  int    `gorm:"column:is_default;type:int(1);not null;comment:是否默认" json:"is_default"`
	Location   string `gorm:"column:location;type:varchar(100);not null;comment:位置坐标" json:"location"`
}
