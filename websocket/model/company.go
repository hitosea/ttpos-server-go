package model

// Company 集团表
type Company struct {
	ID            uint   `gorm:"primaryKey;AUTO_INCREMENT;comment:集团id"`
	Uuid          uint64 `gorm:"default:0;comment:UUID"`
	Name          string `gorm:"default:'';comment:集团名称"`
	Logo          string `gorm:"default:'';comment:logo"`
	IsRecycle     uint   `gorm:"default:0;comment:是否回收"`
	IsChain       int    `gorm:"default:1;comment:是否连锁0否1是"`
	ExpireTime    int    `gorm:"default:0;comment:过期时间"`
	AuthDay       int    `gorm:"default:0;comment:授权时间(天) 0为永不过期"`
	AuthStartTime int    `gorm:"default:0;comment:授权开始时间"`
	Status        int    `gorm:"default:1;comment:状态1=》启用0禁用"`
	IsDelete      uint   `gorm:"default:0;comment:是否删除"`
	CreateTime    uint   `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime    uint   `gorm:"autoUpdateTime;comment:更新时间"`
	DeleteTime    uint   `gorm:"default:0;comment:删除时间"`
}
