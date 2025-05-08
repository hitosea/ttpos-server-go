package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type Supplier struct {
	ShopSupplierID    uint    `gorm:"primaryKey;autoIncrement;comment:主键id"`
	ParentID          int     `gorm:"default:0;comment:上级商家id"`
	Name              string  `gorm:"type:varchar(150);default:'';comment:商家姓名"`
	RealName          string  `gorm:"type:varchar(50);default:'';comment:真实姓名"`
	LinkName          string  `gorm:"type:varchar(50);default:'';comment:联系人"`
	LinkPhone         string  `gorm:"type:varchar(25);default:'';comment:联系电话"`
	Logo              string  `gorm:"type:text;comment:logo"`
	Level             int     `gorm:"default:1;comment:商家等级: 1开始"`
	SaleStock         int     `gorm:"default:0;comment:进销存: 0不开启, 1开启"`
	Reserve           int     `gorm:"default:0;comment:预订: 0不开启, 1开启"`
	IsOpenMember      int     `gorm:"default:0;comment:是否开启会员: 0不开启, 1开启"`
	IsOpenTablet      int     `gorm:"default:0;comment:是否开启平板: 0不开启, 1开启"`
	IsOpenScan        int     `gorm:"default:0;comment:是否开启扫码H5: 0不开启, 1开启"`
	IsOpenAssistant   int     `gorm:"default:0;comment:是否开启点餐助手: 0不开启, 1开启"`
	IsOpenKitchenKds  int     `gorm:"default:0;comment:是否开启后厨KDS: 0不开启, 1开启"`
	IsOpenBuffet      int     `gorm:"default:0;comment:是否开启自助餐: 0不开启, 1开启"`
	IsAcceptScanOrder int     `gorm:"default:0;comment:是否开启扫码点餐接单 0不开启, 1开启"`
	IsOpenLocalPrint  int     `gorm:"default:1;comment:是否开启本地打印服务 0不开启, 1开启"`
	CashLimit         int     `gorm:"default:0;comment:收银机上限"`
	KitchenLimit      int     `gorm:"default:0;comment:厨显上限"`
	TabletLimit       int     `gorm:"default:0;comment:平板上限"`
	AssistantLimit    int     `gorm:"default:0;comment:点餐助手上限"`
	TableLimit        int     `gorm:"default:0;comment:桌台上限"`
	PrinterLimit      int     `gorm:"default:0;comment:打印机上限"`
	Timezone          string  `gorm:"type:varchar(50);default:'Asia/Shanghai';comment:时区"`
	Languages         string  `gorm:"type:longtext;default:'';comment:支持语言"`
	Address           string  `gorm:"type:text;comment:联系地址"`
	DeployMode        int     `gorm:"not null;default:0;comment:部署方式 0局域网部署, 1云部署"`
	MacAddr           string  `gorm:"type:varchar(100);comment:mac地址"`
	SerialNumber      string  `gorm:"type:varchar(100);comment:服务序列号"`
	ChainNumber       string  `gorm:"type:varchar(100);default:'';comment:连锁编号"`
	BusinessID        uint    `gorm:"not null;default:0;comment:营业执照"`
	Description       string  `gorm:"type:text;comment:商家介绍"`
	TotalMoney        float64 `gorm:"type:decimal(12,2);default:0.00;comment:总货款"`
	Money             float64 `gorm:"type:decimal(12,2);default:0.00;comment:当前可提现金额"`
	FreezeMoney       float64 `gorm:"type:decimal(12,2);default:0.00;comment:已冻结金额"`
	CashMoney         float64 `gorm:"type:decimal(12,2);default:0.00;comment:累积提现佣金"`
	DepositMoney      float64 `gorm:"type:decimal(12,2);default:0.00;comment:保证金"`
	UserID            int     `gorm:"not null;default:0;comment:会员id"`
	FavCount          int     `gorm:"default:0;comment:关注人数"`
	Status            int8    `gorm:"not null;default:0;comment:店铺状态0营业中1停止营业"`
	StoreType         int8    `gorm:"not null;default:10;comment:店铺类型10加盟20自营"`
	TotalGift         int     `gorm:"default:0;comment:收到的礼物币总数"`
	IsRecycle         int8    `gorm:"not null;default:1;comment:是否禁用0否1是"`
	IsMain            int8    `gorm:"default:0;comment:是否总店，0否1是"`
	ProvinceID        uint    `gorm:"not null;default:0;comment:所在省份id"`
	CityID            uint    `gorm:"not null;default:0;comment:所在城市id"`
	RegionID          uint    `gorm:"not null;default:0;comment:所在辖区id"`
	Longitude         string  `gorm:"type:varchar(50);not null;default:'';comment:门店坐标经度"`
	Latitude          string  `gorm:"type:varchar(50);not null;default:'';comment:门店坐标纬度"`
	ShippingFee       float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:配送费"`
	BagType           int8    `gorm:"not null;default:0;comment:包装费类型0按商品收费1按单收费"`
	BagPrice          float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:包装费"`
	StorebagType      int8    `gorm:"not null;default:0;comment:店内包装费类型0按商品收费1按单收费"`
	StorebagPrice     float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:店内包装费"`
	DeliveryTime      string  `gorm:"type:varchar(100);not null;default:'';comment:外卖营业时间"`
	PickTime          string  `gorm:"type:varchar(100);not null;default:'';comment:自提营业时间"`
	StoreTime         string  `gorm:"type:varchar(100);not null;default:'';comment:店内营业时间"`
	DeliveryDistance  float64 `gorm:"type:float(10,2);not null;default:0.00;comment:配送范围km"`
	DeliverySet       string  `gorm:"type:varchar(150);not null;default:'';comment:外卖配送方式"`
	StoreSet          string  `gorm:"type:varchar(150);not null;default:'';comment:店内用餐方式"`
	MinMoney          float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:最低消费"`
	SettleType        int8    `gorm:"not null;default:10;comment:计算模式10先结账后用餐20先用餐后结账"`
	ServiceType       int8    `gorm:"not null;default:0;comment:服务费类型0按就餐人数1按桌台收费"`
	ServiceMoney      float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:服务费"`
	AutoClose         int8    `gorm:"not null;default:1;comment:0定时清台1立即清台"`
	CloseTime         int     `gorm:"not null;default:0;comment:0分钟清台"`
	CategorySet       int8    `gorm:"not null;default:10;comment:商品分类设置10同步主店20分店创建"`
	IsDelete          int8    `gorm:"default:0;comment:是否删除0，否1是"`
	AppID             uint    `gorm:"not null;default:0;comment:程序id"`
	CreateTime        uint    `gorm:"not null;default:0;comment:创建时间"`
	UpdateTime        int     `gorm:"not null;comment:更新时间"`
}

type App struct {
	AppID         uint   `gorm:"primaryKey;autoIncrement;comment:小程序id"`
	AppName       string `gorm:"type:varchar(255);default:'';comment:应用名称"`
	Logo          int    `gorm:"default:0;comment:logo"`
	IsRecycle     int8   `gorm:"type:tinyint(3);unsigned;not null;default:0;comment:是否回收"`
	IsChain       int8   `gorm:"type:tinyint(3);default:1;comment:是否连锁0否1是"`
	ExpireTime    int    `gorm:"not null;default:0;comment:过期时间"`
	AuthDay       int    `gorm:"default:0;comment:授权时间(天) 0为永不过期"`
	AuthStartTime int    `gorm:"default:0;comment:授权开始时间"`
	Status        int8   `gorm:"type:tinyint(1);not null;default:1;comment:状态1=》启用0禁用"`
	IsDelete      int8   `gorm:"type:tinyint(3);unsigned;not null;default:0;comment:是否删除"`
	CreateTime    int    `gorm:"unsigned;not null;default:0;comment:创建时间"`
	UpdateTime    int    `gorm:"unsigned;not null;default:0;comment:更新时间"`

	Supplier *Supplier `gorm:"foreignKey:AppID;references:AppID"`
}

type AppRepository interface {
	GetAppList() ([]*App, error)
	ConvertApp() error
}

func NewAppService(db *gorm.DB, targetSaasDB *gorm.DB, targetDB *gorm.DB, sourceCompanyId int, targetCompanyUuid uint64) AppRepository {
	return &AppService{
		db:                db,
		targetDB:          targetDB,
		targetCompanyUuid: targetCompanyUuid,
		sourceCompanyId:   sourceCompanyId,
		targetSaasDB:      targetSaasDB,
	}
}

type AppService struct {
	db                *gorm.DB
	targetDB          *gorm.DB
	targetSaasDB      *gorm.DB
	targetCompanyUuid uint64
	sourceCompanyId   int
}

func (s *AppService) GetAppList() ([]*App, error) {
	var apps []*App
	if err := s.db.Preload("Supplier").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (s *AppService) ConvertApp() error {
	appList, err := s.GetAppList()
	if err != nil {
		return err
	}
	targetCompanyRepo := repository.NewCompanyRepo(s.targetSaasDB)
	saasCompany, err := targetCompanyRepo.GetCompanyInfoByUuid(s.targetCompanyUuid)
	if err != nil {
		return err
	}
	err = targetCompanyRepo.UpdateCompany(saasCompany.Uuid, map[string]any{
		"old_company_id": s.sourceCompanyId,
	})
	if err != nil {
		return err
	}
	for _, app := range appList {

		fmt.Println(fmt.Sprintf("app: %+v", app))

		company := model.Company{
			BaseModel: model.BaseModel{
				Uuid:       s.targetCompanyUuid,
				CreateTime: int64(app.CreateTime),
				UpdateTime: int64(app.UpdateTime),
			},
			Name:          app.Supplier.Name,
			Logo:          saasCompany.Logo,
			ExpireTime:    int64(app.ExpireTime),
			AuthDay:       app.AuthDay,
			Status:        int(app.Status),
			AuthStartTime: int64(app.AuthStartTime),
			OldCompanyId:  s.sourceCompanyId,
			CompanySetting: &model.CompanySetting{
				BaseModel: model.BaseModel{
					Uuid:       s.targetCompanyUuid,
					CreateTime: int64(app.CreateTime),
					UpdateTime: int64(app.UpdateTime),
				},
				CompanyUuid:      s.targetCompanyUuid,
				RealName:         app.Supplier.RealName,
				LinkName:         app.Supplier.LinkName,
				LinkPhone:        app.Supplier.LinkPhone,
				SaleStock:        app.Supplier.SaleStock,
				IsOpenMember:     app.Supplier.IsOpenMember,
				IsOpenTablet:     app.Supplier.IsOpenTablet,
				IsOpenH5:         app.Supplier.IsOpenScan,
				IsOpenAssistant:  app.Supplier.IsOpenAssistant,
				IsOpenKitchenKds: app.Supplier.IsOpenKitchenKds,
				IsOpenBuffet:     app.Supplier.IsOpenBuffet,
				IsOpenH5Order:    app.Supplier.IsAcceptScanOrder,
				IsOpenLocalPrint: app.Supplier.IsOpenLocalPrint,
				CashLimit:        app.Supplier.CashLimit,
				KitchenLimit:     app.Supplier.KitchenLimit,
				TabletLimit:      app.Supplier.TabletLimit,
				AssistantLimit:   app.Supplier.AssistantLimit,
				TableLimit:       app.Supplier.TableLimit,
				PrinterLimit:     app.Supplier.PrinterLimit,
				Timezone:         app.Supplier.Timezone,
				Languages:        app.Supplier.Languages,
				Address:          app.Supplier.Address,
			},
		}
		err = repository.NewCompanyRepo(s.targetDB).CreateCompany(company)
		if err != nil {
			return err
		}
	}
	return nil
}
