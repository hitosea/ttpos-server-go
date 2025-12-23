package req

import (
	"regexp"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/context"
)

// DefaultPermissionPassword 默认权限密码（用于低于 2.10.0 版本的兼容性处理）
const DefaultPermissionPassword = "666888"

// CompanyRoleItem 门店角色配置项
type CompanyRoleItem struct {
	CompanyUuid uint64   `json:"company_uuid" binding:"required"` // 门店UUID
	RoleUuids   []uint64 `json:"role_uuids"`                      // 在该门店的角色UUID列表
}

type UpdateStaffReq struct {
	Uuid               uint64   `json:"uuid" binding:"required"`                               // 员工ID
	RealName           string   `json:"real_name" binding:"required,max=100"`                  // 姓名，限制100个字符
	Username           string   `json:"username" binding:"required,max=64,email"`              // 邮箱，限制64个字符
	Phone              string   `json:"phone" binding:"required,max=20"`                       // 手机号，限制20个字符
	Roles              []uint64 `json:"roles"`                                                 // 角色ID列表，非必填，超管没有角色
	Password           string   `json:"password" binding:"omitempty,strong_password"`          // 密码，如果不为空，则不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种
	ConfirmPassword    string   `json:"confirm_password" binding:"omitempty,eqfield=Password"` // 确认密码
	PermissionPassword string   `json:"permission_password"`                                   // 权限密码，4-8位数字，非必填

	// 统一账号
	IsDisable         *int               `json:"is_disable" binding:"omitempty,oneof=1 0"` // 状态 1:禁用 0:启用
	CompanyRoleList   []*CompanyRoleItem `json:"company_role_list,omitempty"`              // 多门店角色配置（上级门店使用，可选）
	RemoveCompanyList []uint64           `json:"remove_company_list,omitempty"`            // 移除多门店角色配置（上级门店使用，可选）
}

// Validate 验证编辑员工请求参数
// 版本兼容性处理：低于 2.10.0 版本，直接设置默认值 666888
func (r *UpdateStaffReq) Validate(ctx context.Context) error {
	// 版本兼容性处理：低于 2.10.0 版本，直接设置默认值
	if ctx.Version(context.LT, constant.ClientVersionV2100) {
		r.PermissionPassword = DefaultPermissionPassword
		return nil
	}
	// >= 2.10.0 版本，如果权限密码不为空，验证格式（4-8位数字）
	if r.PermissionPassword != "" {
		if !isValidPermissionPassword(r.PermissionPassword) {
			return errors.New("密码必须为 4 - 8 位数字")
		}
	}

	// 如果大于2.11.0版本，则IsDisable必填
	if ctx.Version(context.GTE, constant.ClientVersionV2110) && r.IsDisable == nil {
		return errors.New("状态不能为空")
	}

	if len(r.Roles) == 0 && len(r.CompanyRoleList) == 0 {
		return errors.New("角色不能为空")
	}
	return nil
}

var UpdateStaffRequestMessage = map[string]string{
	"uuid.required":            "员工ID不能为空",
	"real_name.required":       "姓名不能为空",
	"real_name.max":            "姓名不能超过100个字",
	"username.required":        "邮箱不能为空",
	"username.max":             "邮箱不能超过64个字",
	"username.email":           "邮箱必须是有效的邮箱格式",
	"phone.required":           "手机号不能为空",
	"phone.max":                "手机号不能超过20个字",
	"password.strong_password": "密码不符合要求：不能包含空格，长度为8-16个字符，必须包含字母、数字、符号中至少2种",
	"confirm_password.eqfield": "两次密码不一致",
}

type UpdateStaffStatusReq struct {
	Uuid   uint64 `json:"uuid" binding:"required"`             // 员工ID
	Status *int   `json:"status" binding:"required,oneof=1 0"` // 状态 1:启用 0:禁用
}

var UpdateStaffStatusRequestMessage = map[string]string{
	"uuid.required":   "员工ID不能为空",
	"status.required": "状态不能为空",
}

type DeleteStaffReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 员工ID
}

var DeleteStaffRequestMessage = map[string]string{
	"uuid.required": "员工ID不能为空",
}

type AddStaffReq struct {
	RealName           string   `json:"real_name" binding:"required,max=100"`                 // 姓名，限制100个字符
	Username           string   `json:"username" binding:"required,max=64,email"`             // 邮箱，限制64个字符
	Phone              string   `json:"phone" binding:"required,max=20"`                      // 手机号，限制20个字符
	Roles              []uint64 `json:"roles" binding:"required"`                             // 角色ID列表
	Password           string   `json:"password" binding:"required,strong_password"`          // 密码，如果不为空，则不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种
	ConfirmPassword    string   `json:"confirm_password" binding:"required,eqfield=Password"` // 确认密码
	PermissionPassword string   `json:"permission_password"`                                  // 权限密码，4-8位数字，>= 2.10.0 版本必填

	// 统一账号
	IsDisable       *int               `json:"is_disable" binding:"omitempty,oneof=1 0"` // 状态 1:禁用 0:启用
	CompanyRoleList []*CompanyRoleItem `json:"company_role_list,omitempty"`              // 多门店角色配置（上级门店使用，可选）
}

// Validate 验证添加员工请求参数
// 版本兼容性处理：低于 2.10.0 版本，直接设置默认值 666888
func (r *AddStaffReq) Validate(ctx context.Context) error {
	// 版本兼容性处理：低于 2.10.0 版本，直接设置默认值
	if ctx.Version(context.LT, constant.ClientVersionV2100) {
		r.PermissionPassword = DefaultPermissionPassword
		return nil
	}
	// >= 2.10.0 版本，权限密码必填
	if r.PermissionPassword == "" {
		return errors.New("权限密码不能为空")
	}
	// 验证权限密码格式（4-8位数字）
	if !isValidPermissionPassword(r.PermissionPassword) {
		return errors.New("密码必须为 4 - 8 位数字")
	}

	if len(r.Roles) == 0 && len(r.CompanyRoleList) == 0 {
		return errors.New("角色不能为空")
	}
	return nil
}

// isValidPermissionPassword 验证权限密码格式（4-8位数字）
func isValidPermissionPassword(password string) bool {
	matched, _ := regexp.MatchString(`^\d{4,8}$`, password)
	return matched
}

var AddStaffRequestMessage = map[string]string{
	"real_name.required":        "姓名不能为空",
	"real_name.max":             "姓名不能超过100个字",
	"username.required":         "邮箱不能为空",
	"username.max":              "邮箱不能超过64个字",
	"username.email":            "邮箱必须是有效的邮箱格式",
	"phone.required":            "手机号不能为空",
	"phone.max":                 "手机号不能超过20个字",
	"roles.required":            "角色不能为空",
	"password.required":         "密码不能为空",
	"password.strong_password":  "密码不符合要求：不能包含空格，长度为8-16个字符，必须包含字母、数字、符号中至少2种",
	"confirm_password.required": "确认密码不能为空",
	"confirm_password.eqfield":  "两次密码不一致",
}

type AddRoleReq struct {
	Name        string   `json:"name" binding:"required,min=1,max=50"`       // 角色名称，1-50字符
	AccessUuids []uint64 `json:"access_uuids" binding:"required,min=1,dive"` // 权限ID列表，至少选择一个
}

var AddRoleRequestMessage = map[string]string{
	"name.required":         "角色名称不能为空",
	"name.min":              "角色名称至少1个字符",
	"name.max":              "角色名称不能超过50个字符",
	"access_uuids.required": "权限列表不能为空",
	"access_uuids.min":      "至少选择一个权限",
}

type UpdateRoleReq struct {
	Uuid        uint64   `json:"uuid" binding:"required"`                    // 角色ID
	Name        string   `json:"name" binding:"required,min=1,max=50"`       // 角色名称，1-50字符
	AccessUuids []uint64 `json:"access_uuids" binding:"required,min=1,dive"` // 权限ID列表，至少选择一个
	StaffUuids  []uint64 `json:"staff_uuids"`                                // 员工ID列表（编辑时关联员工），可选
}

var UpdateRoleRequestMessage = map[string]string{
	"uuid.required":         "角色ID不能为空",
	"name.required":         "角色名称不能为空",
	"name.min":              "角色名称至少1个字符",
	"name.max":              "角色名称不能超过50个字符",
	"access_uuids.required": "权限列表不能为空",
	"access_uuids.min":      "至少选择一个权限",
}

type GetRoleDetailReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 角色UUID
}

var GetRoleDetailRequestMessage = map[string]string{
	"uuid.required": "角色UUID不能为空",
}

type DeleteRoleReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 角色ID
}

var DeleteRoleRequestMessage = map[string]string{
	"uuid.required": "角色ID不能为空",
}

type GetRoleReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 角色ID
}

var GetRoleRequestMessage = map[string]string{
	"uuid.required": "角色ID不能为空",
}

type GetStaffListReq struct {
	dto.PageReq
	IsFilterSuper int    `form:"is_filter_super"` // 是否过滤超级管理员
	Keyword       string `form:"keyword"`         // 关键词, 姓名、邮箱、手机号
	CompanyUuid   uint64 `form:"company_uuid"`    // 门店UUID
}

// PageReqMessage 分页请求参数错误信息
var GetStaffListReqReqMessage = map[string]string{
	"page_no.min":   "页码不能小于1",
	"page_size.min": "每页大小不能小于1",
	"page_size.max": "每页大小不能大于1100",
}

type SearchStaffByKeywordReq struct {
	Field   string `form:"field" binding:"required,oneof=email phone"` // 字段, 邮箱、手机号
	Keyword string `form:"keyword" binding:"required"`                 // 关键词, 邮箱、手机号
}

type GetStaffDetailReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 员工UUID
}

var GetStaffDetailRequestMessage = map[string]string{
	"uuid.required": "员工UUID不能为空",
}

// QueryStaffByContactReq 根据邮箱或手机号查询员工请求
type QueryStaffByContactReq struct {
	Keyword string `form:"keyword" binding:"omitempty"` // 搜索关键词（邮箱或手机号，支持模糊匹配）
}
