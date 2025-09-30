package erp

// Address 地址信息结构体
// 用于表示联系人或供应商的地址信息
type Address struct {
	Name                 string        `json:"name,omitempty"`                    // 地址名称
	Owner                string        `json:"owner,omitempty"`                   // 所有者
	Creation             string        `json:"creation,omitempty"`                // 创建时间
	Modified             string        `json:"modified,omitempty"`                // 修改时间
	ModifiedBy           string        `json:"modified_by,omitempty"`             // 修改者
	Docstatus            int           `json:"docstatus,omitempty"`               // 文档状态
	Idx                  int           `json:"idx,omitempty"`                     // 索引
	AddressTitle         string        `json:"address_title,omitempty"`           // 地址标题
	AddressType          string        `json:"address_type,omitempty"`            // 地址类型
	AddressLine1         string        `json:"address_line1,omitempty"`           // 地址行1
	AddressLine2         *string       `json:"address_line2,omitempty"`           // 地址行2
	City                 string        `json:"city,omitempty"`                    // 城市
	County               *string       `json:"county,omitempty"`                  // 县/区
	State                *string       `json:"state,omitempty"`                   // 州/省
	Country              string        `json:"country,omitempty"`                 // 国家
	Pincode              *string       `json:"pincode,omitempty"`                 // 邮政编码
	EmailId              *string       `json:"email_id,omitempty"`                // 邮箱ID
	Phone                *string       `json:"phone,omitempty"`                   // 电话
	Fax                  *string       `json:"fax,omitempty"`                     // 传真
	TaxCategory          *string       `json:"tax_category,omitempty"`            // 税务分类
	IsPrimaryAddress     int           `json:"is_primary_address,omitempty"`      // 是否主要地址
	IsShippingAddress    int           `json:"is_shipping_address,omitempty"`     // 是否发货地址
	Disabled             int           `json:"disabled,omitempty"`                // 是否禁用
	IsYourCompanyAddress int           `json:"is_your_company_address,omitempty"` // 是否公司地址
	Doctype              string        `json:"doctype,omitempty"`                 // 文档类型
	Links                []DynamicLink `json:"links,omitempty"`                   // 关联链接
	Localname            string        `json:"localname,omitempty"`               // 本地名称
}

// DynamicLink 动态链接结构体
// 用于表示地址与其他文档的关联关系
type DynamicLink struct {
	Name        string `json:"name,omitempty"`         // 链接名称
	Owner       string `json:"owner,omitempty"`        // 所有者
	Creation    string `json:"creation,omitempty"`     // 创建时间
	Modified    string `json:"modified,omitempty"`     // 修改时间
	ModifiedBy  string `json:"modified_by,omitempty"`  // 修改者
	Docstatus   int    `json:"docstatus,omitempty"`    // 文档状态
	Idx         int    `json:"idx,omitempty"`          // 索引
	LinkDoctype string `json:"link_doctype,omitempty"` // 链接文档类型
	LinkName    string `json:"link_name,omitempty"`    // 链接名称
	LinkTitle   string `json:"link_title,omitempty"`   // 链接标题
	Parent      string `json:"parent,omitempty"`       // 父级
	Parentfield string `json:"parentfield,omitempty"`  // 父级字段
	Parenttype  string `json:"parenttype,omitempty"`   // 父级类型
	Doctype     string `json:"doctype,omitempty"`      // 文档类型
}

// Contact 联系人信息结构体
// 用于表示联系人的详细信息
type Contact struct {
	Name                     string         `json:"name,omitempty"`                        // 联系人名称
	Owner                    string         `json:"owner,omitempty"`                       // 所有者
	Creation                 string         `json:"creation,omitempty"`                    // 创建时间
	Modified                 string         `json:"modified,omitempty"`                    // 修改时间
	ModifiedBy               string         `json:"modified_by,omitempty"`                 // 修改者
	Docstatus                int            `json:"docstatus,omitempty"`                   // 文档状态
	Idx                      int            `json:"idx,omitempty"`                         // 索引
	FirstName                string         `json:"first_name,omitempty"`                  // 名
	MiddleName               *string        `json:"middle_name,omitempty"`                 // 中间名
	LastName                 *string        `json:"last_name,omitempty"`                   // 姓
	FullName                 string         `json:"full_name,omitempty"`                   // 全名
	EmailId                  string         `json:"email_id,omitempty"`                    // 邮箱ID
	User                     *string        `json:"user,omitempty"`                        // 用户
	Address                  *string        `json:"address,omitempty"`                     // 地址
	SyncWithGoogleContacts   int            `json:"sync_with_google_contacts,omitempty"`   // 是否与Google联系人同步
	Status                   string         `json:"status,omitempty"`                      // 状态
	Salutation               *string        `json:"salutation,omitempty"`                  // 称谓
	Designation              *string        `json:"designation,omitempty"`                 // 职位
	Gender                   *string        `json:"gender,omitempty"`                      // 性别
	Phone                    string         `json:"phone,omitempty"`                       // 电话
	MobileNo                 string         `json:"mobile_no,omitempty"`                   // 手机号
	CompanyName              *string        `json:"company_name,omitempty"`                // 公司名称
	Image                    *string        `json:"image,omitempty"`                       // 头像
	GoogleContacts           *string        `json:"google_contacts,omitempty"`             // Google联系人
	GoogleContactsId         *string        `json:"google_contacts_id,omitempty"`          // Google联系人ID
	PulledFromGoogleContacts int            `json:"pulled_from_google_contacts,omitempty"` // 是否从Google联系人拉取
	IsPrimaryContact         int            `json:"is_primary_contact,omitempty"`          // 是否主要联系人
	IsBillingContact         int            `json:"is_billing_contact,omitempty"`          // 是否账单联系人
	Department               *string        `json:"department,omitempty"`                  // 部门
	Unsubscribed             int            `json:"unsubscribed,omitempty"`                // 是否取消订阅
	Doctype                  string         `json:"doctype,omitempty"`                     // 文档类型
	EmailIds                 []interface{}  `json:"email_ids,omitempty"`                   // 邮箱列表
	Links                    []DynamicLink  `json:"links,omitempty"`                       // 关联链接
	PhoneNos                 []ContactPhone `json:"phone_nos,omitempty"`                   // 电话号码列表
	Localname                string         `json:"localname,omitempty"`                   // 本地名称
}

// ContactPhone 联系人电话结构体
// 用于表示联系人的电话信息
type ContactPhone struct {
	Name              string `json:"name,omitempty"`                 // 电话记录名称
	Owner             string `json:"owner,omitempty"`                // 所有者
	Creation          string `json:"creation,omitempty"`             // 创建时间
	Modified          string `json:"modified,omitempty"`             // 修改时间
	ModifiedBy        string `json:"modified_by,omitempty"`          // 修改者
	Docstatus         int    `json:"docstatus,omitempty"`            // 文档状态
	Idx               int    `json:"idx,omitempty"`                  // 索引
	Phone             string `json:"phone,omitempty"`                // 电话号码
	IsPrimaryPhone    int    `json:"is_primary_phone,omitempty"`     // 是否主要电话
	IsPrimaryMobileNo int    `json:"is_primary_mobile_no,omitempty"` // 是否主要手机号
	Parent            string `json:"parent,omitempty"`               // 父级
	Parentfield       string `json:"parentfield,omitempty"`          // 父级字段
	Parenttype        string `json:"parenttype,omitempty"`           // 父级类型
	Doctype           string `json:"doctype,omitempty"`              // 文档类型
}
