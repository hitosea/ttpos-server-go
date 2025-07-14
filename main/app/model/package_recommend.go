package model

import "ttpos-server-go/app/constant"

// PackageRecommend 商品包推荐表  `ttpos_package_recommend`
type PackageRecommend struct {
	BaseModel
	Status   uint   `gorm:"column:status;type:tinyint;not null;default:0;comment:'状态'"` // 0-关闭 1-开启
	Title    string `gorm:"column:title;type:varchar(30);not null;default:'';comment:'推荐标题'"`
	Packages string `gorm:"column:packages;type:text;not null;default:'';comment:'推荐商品，对象数组'"` // JSON 格式. 如 [{"uuid":"1578832498688000","name":"\u514d\u8d39\u5238","sort":"3"},{"uuid":"2572828057600000","name":"\u732a\u8089\u996d","sort":"5"},{"uuid":"6250771292160000","name":"\u9999\u82b9\u7f8a\u8089\u9505","sort":"6"}]
}

// 是否开启推荐
func (p *PackageRecommend) IsOpen() bool {
	return p.Status == constant.PackageRecommendStatusOpen
}
