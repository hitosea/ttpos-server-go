package erp

// User ERP用户信息结构体
type User struct {
	Name                              string     `json:"name,omitempty"`                                    // 用户名称
	Owner                             string     `json:"owner,omitempty"`                                   // 所有者
	Creation                          string     `json:"creation,omitempty"`                                // 创建时间
	Modified                          string     `json:"modified,omitempty"`                                // 修改时间
	ModifiedBy                        string     `json:"modified_by,omitempty"`                             // 修改者
	Docstatus                         int        `json:"docstatus,omitempty"`                               // 文档状态
	Idx                               int        `json:"idx,omitempty"`                                     // 索引
	Enabled                           int        `json:"enabled,omitempty"`                                 // 是否启用
	Email                             string     `json:"email,omitempty"`                                   // 邮箱地址
	FirstName                         string     `json:"first_name,omitempty"`                              // 名字
	FullName                          string     `json:"full_name,omitempty"`                               // 全名
	Username                          string     `json:"username,omitempty"`                                // 用户名
	Language                          string     `json:"language,omitempty"`                                // 语言
	TimeZone                          string     `json:"time_zone,omitempty"`                               // 时区
	SendWelcomeEmail                  int        `json:"send_welcome_email,omitempty"`                      // 发送欢迎邮件
	Unsubscribed                      int        `json:"unsubscribed,omitempty"`                            // 是否取消订阅
	RoleProfileName                   string     `json:"role_profile_name,omitempty"`                       // 角色配置名称
	MuteSounds                        int        `json:"mute_sounds,omitempty"`                             // 静音设置
	DeskTheme                         string     `json:"desk_theme,omitempty"`                              // 桌面主题
	SearchBar                         int        `json:"search_bar,omitempty"`                              // 搜索栏设置
	Notifications                     int        `json:"notifications,omitempty"`                           // 通知设置
	ListSidebar                       int        `json:"list_sidebar,omitempty"`                            // 列表侧边栏设置
	BulkActions                       int        `json:"bulk_actions,omitempty"`                            // 批量操作设置
	ViewSwitcher                      int        `json:"view_switcher,omitempty"`                           // 视图切换器设置
	FormSidebar                       int        `json:"form_sidebar,omitempty"`                            // 表单侧边栏设置
	Timeline                          int        `json:"timeline,omitempty"`                                // 时间线设置
	Dashboard                         int        `json:"dashboard,omitempty"`                               // 仪表板设置
	NewPassword                       string     `json:"new_password,omitempty"`                            // 新密码
	LogoutAllSessions                 int        `json:"logout_all_sessions,omitempty"`                     // 注销所有会话
	DocumentFollowNotify              int        `json:"document_follow_notify,omitempty"`                  // 文档关注通知
	DocumentFollowFrequency           string     `json:"document_follow_frequency,omitempty"`               // 文档关注频率
	FollowCreatedDocuments            int        `json:"follow_created_documents,omitempty"`                // 关注创建的文档
	FollowCommentedDocuments          int        `json:"follow_commented_documents,omitempty"`              // 关注评论的文档
	FollowLikedDocuments              int        `json:"follow_liked_documents,omitempty"`                  // 关注点赞的文档
	FollowAssignedDocuments           int        `json:"follow_assigned_documents,omitempty"`               // 关注分配的文档
	FollowSharedDocuments             int        `json:"follow_shared_documents,omitempty"`                 // 关注共享的文档
	ThreadNotify                      int        `json:"thread_notify,omitempty"`                           // 主题通知
	SendMeACopy                       int        `json:"send_me_a_copy,omitempty"`                          // 发送副本给我
	AllowedInMentions                 int        `json:"allowed_in_mentions,omitempty"`                     // 允许在提及中
	SimultaneousSessions              int        `json:"simultaneous_sessions,omitempty"`                   // 并发会话数
	LastIp                            string     `json:"last_ip,omitempty"`                                 // 最后登录IP
	LoginAfter                        int        `json:"login_after,omitempty"`                             // 登录后设置
	UserType                          string     `json:"user_type,omitempty"`                               // 用户类型
	LastActive                        string     `json:"last_active,omitempty"`                             // 最后活跃时间
	LoginBefore                       int        `json:"login_before,omitempty"`                            // 登录前设置
	BypassRestrictIpCheckIf2faEnabled int        `json:"bypass_restrict_ip_check_if_2fa_enabled,omitempty"` // 2FA启用时绕过IP检查
	LastLogin                         string     `json:"last_login,omitempty"`                              // 最后登录时间
	LastKnownVersions                 string     `json:"last_known_versions,omitempty"`                     // 最后已知版本
	ApiKey                            string     `json:"api_key,omitempty"`                                 // API密钥
	ApiSecret                         string     `json:"api_secret,omitempty"`                              // API密钥
	OnboardingStatus                  string     `json:"onboarding_status,omitempty"`                       // 入职状态
	Doctype                           string     `json:"doctype,omitempty"`                                 // 文档类型
	SocialLogins                      []string   `json:"social_logins,omitempty"`                           // 社交登录
	Defaults                          []string   `json:"defaults,omitempty"`                                // 默认设置
	Roles                             []UserRole `json:"roles,omitempty"`                                   // 用户角色列表
	UserEmails                        []string   `json:"user_emails,omitempty"`                             // 用户邮箱列表
	BlockModules                      []string   `json:"block_modules,omitempty"`                           // 阻止的模块
}

// UserRole 用户角色信息结构体
type UserRole struct {
	Name        string `json:"name,omitempty"`        // 角色ID
	Owner       string `json:"owner,omitempty"`       // 所有者
	Creation    string `json:"creation,omitempty"`    // 创建时间
	Modified    string `json:"modified,omitempty"`    // 修改时间
	ModifiedBy  string `json:"modified_by,omitempty"` // 修改者
	Docstatus   int    `json:"docstatus,omitempty"`   // 文档状态
	Idx         int    `json:"idx,omitempty"`         // 索引
	Role        string `json:"role,omitempty"`        // 角色名称
	Parent      string `json:"parent,omitempty"`      // 父级
	Parentfield string `json:"parentfield,omitempty"` // 父级字段
	Parenttype  string `json:"parenttype,omitempty"`  // 父级类型
	Doctype     string `json:"doctype,omitempty"`     // 文档类型
}
