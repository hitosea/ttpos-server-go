package model

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/utils"
)

// Company 集团表 ttpos_company
type Company struct {
	BaseModel
	Name          string `gorm:"column:name;type:varchar(255);comment:集团名称;NOT NULL" json:"name"`
	Logo          string `gorm:"column:logo;type:varchar(255);comment:logo;NOT NULL" json:"logo"`
	ExpireTime    int64  `gorm:"column:expire_time;type:int(10);default:0;comment:过期时间;not null;NOT NULL" json:"expire_time"`
	AuthDay       int    `gorm:"column:auth_day;type:int(11);default:0;comment:授权时间(天) 0为永不过期;NOT NULL" json:"auth_day"`
	Status        int    `gorm:"column:status;type:tinyint(1);default:0;comment:状态 1-启用 0-禁用;not null;NOT NULL" json:"status"`
	AuthStartTime int64  `gorm:"column:auth_start_time;type:int(10);default:0;comment:授权开始时间（时间戳）;NOT NULL" json:"auth_start_time"`
	OldCompanyId  int    `gorm:"column:old_company_id;type:int(11);default:0;comment:原商家ID;NOT NULL" json:"old_company_id"`
	IsEnableErp   int    `gorm:"column:is_enable_erp;type:int(10);default:0;comment:是否启用ERP: 0不启用, 1启用;NOT NULL" json:"is_enable_erp"`

	CompanySetting *CompanySetting `gorm:"foreignKey:CompanyUuid;references:Uuid" json:"company_setting"`
}

func (company *Company) IsOpenErp() bool {
	return company.IsEnableErp == 1
}

func (company *Company) GetLogo(baseURL string) string {
	logoBase64, err := utils.AddImageDomainAndConvertToBase64(company.Logo, baseURL, true)
	if err != nil {
		return utils.AddImageDomain(company.Logo, baseURL, true)
	}
	return logoBase64
}

func (company *Company) SetNil() {
	company.CompanySetting = nil
}

func (company *Company) IsExpired() bool {
	return company.ExpireTime > 0 && company.ExpireTime < time.Now().Unix()
}

func (company *Company) IsException() bool {
	return company.Status == 0 || company.IsDelete()
}

// CompanySetting 公司设置表 ttpos_company_setting
type CompanySetting struct {
	BaseModel
	CompanyUuid           uint64 `gorm:"column:company_uuid;type:bigint(20);default:0;comment:集团ID;NOT NULL" json:"company_uuid"`
	RealName              string `gorm:"column:real_name;type:varchar(50);comment:真实姓名;NOT NULL" json:"real_name"`
	LinkName              string `gorm:"column:link_name;type:varchar(50);comment:联系人;NOT NULL" json:"link_name"`
	LinkPhone             string `gorm:"column:link_phone;type:varchar(25);comment:联系电话;NOT NULL" json:"link_phone"`
	SaleStock             int    `gorm:"column:sale_stock;type:int(11);default:0;comment:进销存: 0不开启, 1开启;NOT NULL" json:"sale_stock"`
	IsOpenCoupon          int    `gorm:"column:is_open_coupon;type:int(11);default:0;comment:是否开启优惠券: 0不开启, 1开启;NOT NULL" json:"is_open_coupon"`
	IsOpenMarketing       int    `gorm:"column:is_open_marketing;type:int(11);default:0;comment:是否开启营销: 0不开启, 1开启;NOT NULL" json:"is_open_marketing"`
	IsOpenMember          int    `gorm:"column:is_open_member;type:int(11);default:0;comment:是否开启会员: 0不开启, 1开启;NOT NULL" json:"is_open_member"`
	IsOpenTablet          int    `gorm:"column:is_open_tablet;type:int(11);default:0;comment:是否开启平板: 0不开启, 1开启;NOT NULL" json:"is_open_tablet"`
	IsOpenH5              int    `gorm:"column:is_open_h5;type:int(11);default:0;comment:是否开启扫码H5: 0不开启, 1开启;NOT NULL" json:"is_open_scan"`
	IsOpenAssistant       int    `gorm:"column:is_open_assistant;type:int(11);default:0;comment:是否开启点餐助手: 0不开启, 1开启;NOT NULL" json:"is_open_assistant"`
	IsOpenKitchenKds      int    `gorm:"column:is_open_kitchen_kds;type:int(11);default:0;comment:是否开启后厨KDS: 0不开启, 1开启;NOT NULL" json:"is_open_kitchen_kds"`
	IsOpenBuffet          int    `gorm:"column:is_open_buffet;type:int(11);default:0;comment:是否开启自助餐: 0不开启, 1开启;NOT NULL" json:"is_open_buffet"`
	EnableSms             int    `gorm:"column:enable_sms;type:int(11);default:0;comment:是否启用短信功能：0-否；1-是;NOT NULL" json:"enable_sms"`
	SmsQuota              int    `gorm:"column:sms_quota;type:int(11);default:0;comment:短信配额;NOT NULL" json:"sms_quota"`
	IsOpenH5Order         int    `gorm:"column:is_open_h5_order;type:int(11);default:0;comment:是否开启扫码点餐接单 0不开启, 1开启;NOT NULL" json:"is_open_h5_order"`
	IsOpenLocalPrint      int    `gorm:"column:is_open_local_print;type:int(11);default:0;comment:是否开启本地打印服务 0不开启, 1开启;NOT NULL" json:"is_open_local_print"`
	CashLimit             int    `gorm:"column:cash_limit;type:int(11);default:0;comment:收银机上限;NOT NULL" json:"cash_limit"`
	KitchenLimit          int    `gorm:"column:kitchen_limit;type:int(11);default:0;comment:厨显上限;NOT NULL" json:"kitchen_limit"`
	TabletLimit           int    `gorm:"column:tablet_limit;type:int(11);default:0;comment:平板上限;NOT NULL" json:"tablet_limit"`
	AssistantLimit        int    `gorm:"column:assistant_limit;type:int(11);default:0;comment:点餐助手上限;NOT NULL" json:"assistant_limit"`
	TableLimit            int    `gorm:"column:table_limit;type:int(11);default:0;comment:桌台上限;NOT NULL" json:"table_limit"`
	PrinterLimit          int    `gorm:"column:printer_limit;type:int(11);default:0;comment:打印机上限;NOT NULL" json:"printer_limit"`
	Timezone              string `gorm:"column:timezone;type:varchar(50);default:Asia/Shanghai;comment:时区;NOT NULL" json:"timezone"`
	Languages             string `gorm:"column:languages;type:varchar(255);comment:支持语言;NOT NULL" json:"languages"`
	Address               string `gorm:"column:address;type:varchar(255);comment:联系地址;NOT NULL" json:"address"`
	Coordinates           string `gorm:"column:coordinates;type:varchar(255);comment:经纬度，如：13.721899,100.52900;NOT NULL" json:"coordinates"`
	DeliveryConfig        string `gorm:"column:delivery_config;type:text;comment:外送配置;NOT NULL" json:"delivery_config"`
	DeliveryStatus        int    `gorm:"column:delivery_status;type:int(11);default:0;comment:外送配置状态：0-关,1-开;NOT NULL" json:"delivery_status"`
	ErpnextSiteCode       string `gorm:"column:erpnext_site_code;type:varchar(255);default:'';comment:ERPNext站点编码;NOT NULL" json:"erpnext_site_code"`
	ErpnextCompanyAbbr    string `gorm:"column:erpnext_company_abbr;type:varchar(255);default:'';comment:ERPNext公司缩写;NOT NULL" json:"erpnext_company_abbr"`
	ErpnextBranchName     string `gorm:"column:erpnext_branch_name;type:varchar(255);default:'';comment:ERPNext分支名称;NOT NULL" json:"erpnext_branch_name"`
	ErpnextPosProfileName string `gorm:"column:erpnext_pos_profile_name;type:varchar(255);default:'';comment:ERPNext Pos Profile名称;NOT NULL" json:"erpnext_pos_profile_name"`
}

func (model *CompanySetting) GetTimezone() string {
	if model.Timezone == "" {
		return string(utils.ZH_TIMEZONE)
	}
	return model.Timezone
}

func (model *CompanySetting) GetCoordinates() (latitude, longitude string) {
	if model.Coordinates == "" {
		return
	}
	// 分隔后去掉前后空格,转成float64保留6位小数
	latLng := strings.Split(strings.TrimSpace(model.Coordinates), ",")
	if len(latLng) == 2 {
		latFloat, _ := strconv.ParseFloat(strings.TrimSpace(latLng[0]), 64)
		lngFloat, _ := strconv.ParseFloat(strings.TrimSpace(latLng[1]), 64)
		latitude = fmt.Sprintf("%.6f", latFloat)
		longitude = fmt.Sprintf("%.6f", lngFloat)
	}
	return
}

// GetDeliveryConfig 获取外送配置
func (model *CompanySetting) GetDeliveryConfig(channel string, distance float64) (*DeliveryConfigResponse, error) {
	// 如果配置为空，返回空配置
	if model.DeliveryConfig == "" || model.DeliveryConfig == "[]" || len(model.DeliveryConfig) < 10 {
		return nil, errors.WithMessage(errors.ErrInternal, "delivery config is empty")
	}
	var deliveryConfig DeliveryConfig
	if err := json.Unmarshal([]byte(model.DeliveryConfig), &deliveryConfig); err != nil {
		return nil, errors.WithMessage(err, "unmarshal delivery config failed")
	}

	for index := range deliveryConfig {
		item := &deliveryConfig[index]
		item.Channel = strings.ToLower(item.Channel) // 外送渠道转换为小写
		// 按照距离排序,从小到大
		sort.Slice(item.DistanceRange, func(i, j int) bool {
			return item.DistanceRange[i].End < item.DistanceRange[j].End
		})
	}

	config, err := deliveryConfig.GetConfigByChannel(channel, distance)
	if err != nil {
		return nil, errors.WithMessage(err, "get delivery config failed")
	}
	return config, nil
}

// 外送配置
type DeliveryConfig []DeliveryConfigItem

type DeliveryConfigResponse struct {
	Channel                string  `json:"channel"`                  // 外送渠道
	BasicFee               float64 `json:"basic_fee"`                // 基础服务费
	BaseDeliveryFee        float64 `json:"base_delivery_fee"`        // 起步配送费
	RiderAcceptanceTimeout int     `json:"rider_acceptance_timeout"` // 骑手接单超时时间,单位分钟
	PricePerKm             float64 `json:"price_per_km"`             // 每公里价格
	IsInDeliveryRange      bool    `json:"is_in_delivery_range"`     // 是否在配送范围内。如果不在配送范围内，则置灰提交订单按钮
}

// 根据渠道和距离获取配置
// channel: 外送渠道. skootar、grab
// distance: 距离
func (model *DeliveryConfig) GetConfigByChannel(channel string, distance float64) (*DeliveryConfigResponse, error) {
	var config *DeliveryConfigResponse
	for _, item := range *model {
		if item.Channel == channel {
			for index, distanceRange := range item.DistanceRange {
				if float64(distanceRange.End) >= distance {
					config = &DeliveryConfigResponse{
						Channel:                item.Channel,
						BasicFee:               item.BasicFee,
						BaseDeliveryFee:        item.BaseDeliveryFee,
						RiderAcceptanceTimeout: item.RiderAcceptanceTimeout,
						PricePerKm:             distanceRange.PricePerKm,
						IsInDeliveryRange:      true,
					}
					break // 找到第一个符合条件的配置就退出
				}
				if index == len(item.DistanceRange)-1 {
					config = &DeliveryConfigResponse{
						Channel:                item.Channel,
						BasicFee:               item.BasicFee,
						BaseDeliveryFee:        item.BaseDeliveryFee,
						RiderAcceptanceTimeout: item.RiderAcceptanceTimeout,
						PricePerKm:             distanceRange.PricePerKm,
						IsInDeliveryRange:      false,
					}
				}
			}
		}
	}
	if config == nil {
		return nil, errors.WithMessage(errors.NewWithCode(constant.CodeOrderAddressNotInDeliveryRange, "delivery config not found"), "delivery config not found")
	}
	return config, nil
}

type DeliveryConfigItem struct {
	Channel                string          `json:"channel"`                  // 外送渠道
	ConfigType             string          `json:"config_type"`              // 配置类型 auto_sync: 自动同步，manual: 手动配置
	BasicFee               float64         `json:"basic_fee"`                // 基础服务费
	BaseDeliveryFee        float64         `json:"base_delivery_fee"`        // 起步配送费
	RiderAcceptanceTimeout int             `json:"rider_acceptance_timeout"` // 骑手接单超时时间
	DistanceRange          []DistanceRange `json:"distance_range"`           // 距离范围
}
type DistanceRange struct {
	End         float64 `json:"end"`          // 距离
	PricePerKm  float64 `json:"price_per_km"` // 每公里价格
	IsUnlimited bool    `json:"is_unlimited"` // 是否无限
}

// GetIsOpenCoupon 是否开启优惠券
func (model *CompanySetting) GetIsOpenCoupon() bool {
	return model.IsOpenCoupon == 1
}

// GetIsOpenH5Order 是否开启扫码点餐接单
func (model *CompanySetting) GetIsOpenH5Order() bool {
	return model.IsOpenH5Order == 1
}

// SmsEnabled 短信功能是否开启
func (model *CompanySetting) SmsEnabled() bool {
	return model.EnableSms == 1 && model.SmsQuota > 0
}

// GetDefaultLanguage 获取默认语言。 注意：这是admin端的显示的第一个语言
func (model *CompanySetting) GetDefaultLanguage() string {
	languages := model.GetLanguages()
	if len(languages) > 0 {
		return languages[0]
	}
	return "en"
}

func (model *CompanySetting) GetLanguages() []string {
	languages := make([]string, 0)
	if err := json.Unmarshal([]byte(model.Languages), &languages); err != nil {
		tmp := ""
		err = json.Unmarshal([]byte(model.Languages), &tmp)
		if err != nil {
			return nil
		}
		err = json.Unmarshal([]byte(tmp), &languages)
		if err != nil {
			return nil
		}
	}
	return languages
}

// CompanyStaff saas库保存的集团员工关联表 ttpos_company_staff
type CompanyStaff struct {
	BaseModel
	CompanyUuid uint64 `gorm:"column:company_uuid;type:bigint(20) unsigned;default:0;comment:集团ID;NOT NULL" json:"company_uuid"`
	Username    string `gorm:"column:username;type:varchar(255);comment:员工账号;NOT NULL" json:"username"`
	Phone       string `gorm:"column:phone;type:varchar(255);comment:员工手机号;NOT NULL" json:"phone"`
	IsSuper     int    `gorm:"column:is_super;type:int(11);default:0;comment:是否超级管理员" json:"is_super"`

	Company *Company `gorm:"foreignKey:CompanyUuid;references:Uuid"`
}
