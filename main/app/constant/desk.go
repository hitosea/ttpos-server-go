package constant

// 桌台状态
const (
	DeskStatusAvailable = 0 //空闲
	DeskStatusNotBuffet = 1 //非自助餐
	DeskStatusBuffet    = 2 //自助餐
	DeskStatusWait      = 3 //待清台
	DeskStatusLock      = 4 //锁单
)

const (
	DeskStatusClose = 0 //未开台
	DeskStatusOpen  = 1 //开台

	// 桌台禁用
	DeskDisable = 1 //禁用
	DeskEnable  = 0 //启用
)
