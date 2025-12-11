package jwt

const CompanyUuid = "company_uuid"                // 商家ID
const StaffUuid = "staff_uuid"                    // 员工ID
const MemberUuid = "member_uuid"                  // 会员ID
const Source = "source"                           // 来源: cashier-收银机; tablet-平板; kitchen-厨显; assistant-点餐助手; h5-扫码H5
const DeskUuid = "desk_uuid"                      // 桌子ID, 扫码点餐时token中携带
const DeviceId = "device_id"                      // 设备ID
const DeviceUuid = "device_uuid"                  // 设备ID
const CompanySetting = "company_setting"          // 商家设置
const Company = "company"                         // 商家信息
const Member = "member"                           // 会员信息
const Staff = "staff"                             // 员工信息
const AssistantStaffUuid = "assistant_staff_uuid" // 点餐助手员工ID
const AssistantDeviceId = "assistant_device_id"   // 点餐助手设备ID
const Brand = "brand"                             // 品牌名称

const RequestUuid = "request_uuid" // http请求的唯一uuid
const DB = "db"                    // 数据库连接

const (
	SourceCashier   = "cashier"
	SourceTablet    = "tablet"
	SourceKitchen   = "kitchen"
	SourceAssistant = "assistant"
	SourceH5        = "h5"
	SourceMember    = "member"
)
