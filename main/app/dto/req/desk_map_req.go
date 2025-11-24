package req

import "ttpos-server-go/app/errors"

// DeskMapLayoutDetailReq 获取区域布局详情请求
//
// 任务: story-admin-desktop-table-map Phase 2.3
// 需求: R1.1-R1.6
//
// @version v2.10.0
type DeskMapLayoutDetailReq struct {
	RegionUuid uint64 `form:"region_uuid" binding:"required"` // 区域UUID
}

// DeskMapSaveLayoutReq 保存区域布局请求
//
// 任务: story-admin-desktop-table-map Phase 2.3
// 需求: R1.1-R1.6
//
// @version v2.10.0
type DeskMapSaveLayoutReq struct {
	RegionUuid uint64                 `json:"region_uuid" binding:"required"` // 区域UUID
	Desks      []DeskMapLayoutDeskReq `json:"desks" binding:"required"`       // 桌台布局数据
}

// DeskMapLayoutDeskReq 单个桌台的布局信息
type DeskMapLayoutDeskReq struct {
	DeskUuid uint64  `json:"desk_uuid" binding:"required"` // 桌台UUID
	Shape    string  `json:"shape" binding:"required"`     // 形状: circle-圆形, rect-矩形
	RangeMin uint    `json:"range_min"`                    // 最少人数
	RangeMax uint    `json:"range_max"`                    // 最多人数
	X        float64 `json:"x"`                            // X坐标
	Y        float64 `json:"y"`                            // Y坐标
	Width    float64 `json:"width"`                        // 宽度
	Height   float64 `json:"height"`                       // 高度
	Rotation float64 `json:"rotation"`                     // 旋转角度
}

// Validate 验证布局数据
func (req *DeskMapSaveLayoutReq) Validate() error {
	if req.RegionUuid == 0 {
		return errors.New("区域UUID不能为空")
	}
	if len(req.Desks) == 0 {
		return errors.New("布局数据不能为空")
	}
	// 验证每个桌台数据
	for _, desk := range req.Desks {
		if desk.DeskUuid == 0 {
			return errors.New("桌台UUID不能为空")
		}
		if desk.Shape != "circle" && desk.Shape != "rect" {
			return errors.New("桌台形状必须是 circle 或 rect")
		}
	}
	return nil
}
