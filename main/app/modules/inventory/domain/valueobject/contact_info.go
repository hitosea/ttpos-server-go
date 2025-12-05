package valueobject

// ContactInfo 联系信息值对象
type ContactInfo struct {
	contact string // 联系人
	phone   string // 联系电话
	address string // 地址
}

// NewContactInfo 创建联系信息
func NewContactInfo(contact, phone, address string) ContactInfo {
	return ContactInfo{
		contact: contact,
		phone:   phone,
		address: address,
	}
}

// Contact 获取联系人
func (c ContactInfo) Contact() string {
	return c.contact
}

// Phone 获取联系电话
func (c ContactInfo) Phone() string {
	return c.phone
}

// Address 获取地址
func (c ContactInfo) Address() string {
	return c.address
}

// Equals 比较两个联系信息是否相等
func (c ContactInfo) Equals(other ContactInfo) bool {
	return c.contact == other.contact &&
		c.phone == other.phone &&
		c.address == other.address
}


