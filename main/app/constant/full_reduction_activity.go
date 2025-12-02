package constant

// 满减类型
const (
	FullReductionTypeStep  = 0 // 阶梯满减
	FullReductionTypeCycle = 1 // 循环满减
)

// 活动状态
const (
	ActivityStatusAll        = "all"         // 全部
	ActivityStatusNotStarted = "not_start"   // 未开始
	ActivityStatusInProgress = "in_progress" // 进行中
	ActivityStatusEnded      = "end"         // 已结束
)

// 满减类型名称映射
var FullReductionTypeNames = map[int]string{
	FullReductionTypeStep:  "阶梯满减", // 阶梯满减
	FullReductionTypeCycle: "循环满减", // 循环满减
}
