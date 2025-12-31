package valueobject

// CarouselItem 轮播项值对象
type CarouselItem struct {
	FilePath string `json:"file_path"`
	RealName string `json:"real_name"`
	Sort     string `json:"sort"`
	Type     string `json:"type"`
}

// IsValid 验证轮播项是否有效
func (c CarouselItem) IsValid() bool {
	return c.FilePath != ""
}

// NewCarouselItem 创建新的轮播项
func NewCarouselItem(filePath, realName, sort, itemType string) CarouselItem {
	return CarouselItem{
		FilePath: filePath,
		RealName: realName,
		Sort:     sort,
		Type:     itemType,
	}
}
