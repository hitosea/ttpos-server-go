package constant

// 满减类型
const (
	FullReductionTypeStep  = 0 // 阶梯满减
	FullReductionTypeCycle = 1 // 循环满减
)

// 活动状态
const (
	ActivityStatusNotStarted = "not_start"   // 未开始
	ActivityStatusInProgress = "in_progress" // 进行中
	ActivityStatusEnded      = "end"         // 已结束
)
