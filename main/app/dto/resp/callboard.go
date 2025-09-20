package resp

// BindCodeResp 获取绑定码响应
type BindCodeResp struct {
	BindCode   string `json:"bind_code"`
	ExpireTime int64  `json:"expire_time"`
}

// BindInfoResp 获取绑定信息响应
type BindInfoResp struct {
	DeviceSecret string `json:"device_secret"`
	Lang1        string `json:"lang1"`
	Lang2        string `json:"lang2"`
}

// QueueDataResp 队列数据响应
type QueueDataResp struct {
	Lang1          string   `json:"lang1"`
	Lang2          string   `json:"lang2"`
	UpdateTime     int64    `json:"update_time"`
	PreparingQueue []string `json:"preparing_queue"`
	PreparedQueue  []string `json:"prepared_queue"`
}

// DeviceListResp 设备列表响应
type DeviceListResp struct {
	List []DeviceItem `json:"list"`
}

// DeviceItem 设备项
type DeviceItem struct {
	Uuid     uint64 `json:"uuid"`
	Lang1    string `json:"lang1"`
	Lang2    string `json:"lang2"`
	DeviceId string `json:"device_id"`
	BindTime int64  `json:"bind_time"`
}

type UpdateBindInfoResp struct {
}
