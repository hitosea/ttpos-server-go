package permission

import (
	"context"
	"strings"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var Permission = new(sPermission)

type sPermission struct{}

func init() {
	service.RegisterPermission(Permission)
}

// GetPosPermissionRuleList 获取POS权限规则列表
// 根据查询条件过滤并返回权限规则信息列表
func (s *sPermission) GetPosPermissionRuleList(ctx context.Context, req *erp.PosPermissionRule) (res []*erp.PosPermissionRule, err error) {
	// 构建查询过滤器
	filters := s.buildPermissionRuleListFilters(req)

	// 查询权限规则列表
	permissionRuleList, err := s.queryPermissionRuleList(ctx, filters)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询权限规则列表失败")
	}

	return permissionRuleList, nil
}

// buildPermissionRuleListFilters 构建权限规则列表查询过滤器
func (s *sPermission) buildPermissionRuleListFilters(req *erp.PosPermissionRule) [][]string {
	filters := make([][]string, 0, 4)

	// 按规则代码过滤
	if len(req.RuleCode) > 0 {
		if strings.Contains(req.RuleCode, "%") {
			filters = append(filters, g.ArrayStr{"rule_code", "like", req.RuleCode})
		} else {
			filters = append(filters, g.ArrayStr{"rule_code", "=", req.RuleCode})
		}
	}

	// 按规则名称过滤
	if len(req.RuleName) > 0 {
		if strings.Contains(req.RuleName, "%") {
			filters = append(filters, g.ArrayStr{"rule_name", "like", req.RuleName})
		} else {
			filters = append(filters, g.ArrayStr{"rule_name", "=", req.RuleName})
		}
	}

	// 按规则类型过滤
	if len(req.RuleType) > 0 {
		filters = append(filters, g.ArrayStr{"rule_type", "=", req.RuleType})
	}

	return filters
}

// queryPermissionRuleList 执行权限规则列表查询
func (s *sPermission) queryPermissionRuleList(ctx context.Context, filters [][]string) ([]*erp.PosPermissionRule, error) {
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosPermissionRule,
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "rule_code", "rule_name", "rule_type", "owner", "creation", "modified", "modified_by", "docstatus", "idx"},
		Filters: filters,
		Limit:   1000,
	})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析权限规则列表响应失败")
	}

	// 转换为权限规则信息列表
	dataArray := j.GetJsons("data")
	permissionRuleList := make([]*erp.PosPermissionRule, 0, len(dataArray))

	for _, data := range dataArray {
		rule := &erp.PosPermissionRule{
			Name:       data.Get("name").String(),
			RuleCode:   data.Get("rule_code").String(),
			RuleName:   data.Get("rule_name").String(),
			RuleType:   data.Get("rule_type").String(),
			Owner:      data.Get("owner").String(),
			Creation:   data.Get("creation").String(),
			Modified:   data.Get("modified").String(),
			ModifiedBy: data.Get("modified_by").String(),
			Docstatus:  data.Get("docstatus").Int(),
			Idx:        data.Get("idx").Int(),
			Doctype:    erp.DocTypePosPermissionRule,
		}
		permissionRuleList = append(permissionRuleList, rule)
	}

	return permissionRuleList, nil
}

// getPermissionCompanyList 获取权限规则的公司列表
func (s *sPermission) getPermissionCompanyList(ctx context.Context, parentName string) ([]erp.PermissionCompanyList, error) {
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: "Permission Company List",
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "company", "parent", "parentfield", "parenttype", "owner", "creation", "modified", "modified_by", "docstatus", "idx"},
		Filters: [][]string{{"parent", "=", parentName}},
		Limit:   100,
	})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析公司列表响应失败")
	}

	// 转换为公司列表
	dataArray := j.GetJsons("data")
	companyList := make([]erp.PermissionCompanyList, 0, len(dataArray))

	for _, data := range dataArray {
		company := erp.PermissionCompanyList{
			Name:        data.Get("name").String(),
			Company:     data.Get("company").String(),
			Parent:      data.Get("parent").String(),
			Parentfield: data.Get("parentfield").String(),
			Parenttype:  data.Get("parenttype").String(),
			Owner:       data.Get("owner").String(),
			Creation:    data.Get("creation").String(),
			Modified:    data.Get("modified").String(),
			ModifiedBy:  data.Get("modified_by").String(),
			Docstatus:   data.Get("docstatus").Int(),
			Idx:         data.Get("idx").Int(),
			Doctype:     "Permission Company List",
		}
		companyList = append(companyList, company)
	}

	return companyList, nil
}

// GetPosPermissionRule 获取POS权限规则详情
// 根据名称标识获取权限规则的完整信息
func (s *sPermission) GetPosPermissionRule(ctx context.Context, ruleCode string) (res *erp.PosPermissionRule, err error) {

	// 查询权限规则详情
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosPermissionRule,
		Name:    ruleCode,
	}, &erp.RequestParams{})

	if err != nil {
		return nil, gerror.Wrapf(err, "查询权限规则详情失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析权限规则详情响应失败")
	}

	// 转换为权限规则结构体
	var permissionRule = &erp.PosPermissionRule{}
	if err := gconv.Scan(j.Get("data"), &permissionRule); err != nil {
		return nil, gerror.Wrapf(err, "转换权限规则数据失败")
	}

	return permissionRule, nil
}

// CreatePosPermissionRule 创建POS权限规则
// 创建新的权限规则记录
func (s *sPermission) CreatePosPermissionRule(ctx context.Context, req *erp.PosPermissionRule) (res *erp.PosPermissionRule, err error) {
	// 参数验证
	if err := s.validateCreatePermissionRuleReq(req); err != nil {
		return nil, err
	}

	// 创建权限规则
	resp, err := service.Document().Create(ctx, erp.DocTypePosPermissionRule, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建权限规则失败")
	}

	// 解析创建结果
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析创建权限规则响应失败")
	}

	// 转换为权限规则结构体
	var permissionRule = &erp.PosPermissionRule{}
	if err := gconv.Scan(j.Get("data"), &permissionRule); err != nil {
		return nil, gerror.Wrapf(err, "转换创建的权限规则数据失败")
	}

	return permissionRule, nil
}

// validateCreatePermissionRuleReq 验证创建权限规则请求参数
func (s *sPermission) validateCreatePermissionRuleReq(req *erp.PosPermissionRule) error {
	if len(req.RuleCode) == 0 {
		return gerror.New("规则代码不能为空")
	}
	if len(req.RuleName) == 0 {
		return gerror.New("规则名称不能为空")
	}
	if len(req.RuleType) == 0 {
		return gerror.New("规则类型不能为空")
	}
	if req.RuleType != "White" && req.RuleType != "Black" {
		return gerror.New("规则类型必须为White或Black")
	}
	return nil
}

// UpdatePosPermissionRule 更新POS权限规则
// 更新现有的权限规则记录
func (s *sPermission) UpdatePosPermissionRule(ctx context.Context, req *erp.PosPermissionRule) (res *erp.PosPermissionRule, err error) {
	// 参数验证
	if len(req.Name) == 0 {
		return nil, gerror.New("权限规则名称不能为空")
	}

	// 更新权限规则
	resp, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosPermissionRule,
		Name:    req.Name,
	}, req)

	if err != nil {
		return nil, gerror.Wrapf(err, "更新权限规则失败")
	}

	// 解析更新结果
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析更新权限规则响应失败")
	}

	// 转换为权限规则结构体
	var permissionRule = &erp.PosPermissionRule{}
	if err := gconv.Scan(j.GetJson("data"), &permissionRule); err != nil {
		return nil, gerror.Wrapf(err, "转换更新的权限规则数据失败")
	}

	return permissionRule, nil
}

// DeletePosPermissionRule 删除POS权限规则
// 删除指定的权限规则记录
func (s *sPermission) DeletePosPermissionRule(ctx context.Context, ruleCode string) (err error) {
	// 参数验证
	if len(ruleCode) == 0 {
		return gerror.New("权限规则编码不能为空")
	}

	// 删除权限规则
	_, err = service.Document().Delete(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosPermissionRule,
		Name:    ruleCode,
	})

	if err != nil {
		return gerror.Wrapf(err, "删除权限规则失败")
	}

	return nil
}

// CheckPermission 检查权限
// 根据权限规则列表和公司名称检查是否有权限
// 参数：
//   - ctx: 上下文对象
//   - permissionList: 权限规则列表
//   - company: 公司名称
//
// 返回：
//   - hasPermission: 是否有权限
//   - err: 错误信息
func (s *sPermission) CheckPermission(ctx context.Context, permissionList []erp.PermissionRule, company string) (hasPermission bool, err error) {
	if len(permissionList) == 0 {
		// 如果没有权限规则，默认允许访问
		return true, nil
	}

	// 参数验证
	if len(company) == 0 {
		return false, gerror.New("公司名称不能为空")
	}

	permissionRuleList := make([]*erp.PosPermissionRule, 0, len(permissionList))
	for _, rule := range permissionList {
		permissionRule, err := service.Permission().GetPosPermissionRule(ctx, rule.PermissionRule)
		if err != nil {
			g.Log().Errorf(ctx, "获取权限规则失败,permissionList:%s,err:%v", permissionList, err)
			return false, err // 获取权限规则失败，跳过该物品
		}
		permissionRuleList = append(permissionRuleList, permissionRule)
	}

	// 标记是否存在白名单规则
	hasWhiteRule := false
	// 标记是否在白名单中
	inWhiteList := false

	// 遍历权限规则列表
	for _, rule := range permissionRuleList {

		// 检查黑名单规则
		if rule.RuleType == "Black" {
			// 检查公司是否在黑名单中
			if s.isCompanyInList(rule.CompanyList, company) {
				g.Log().Infof(ctx, "公司 %s 在黑名单规则 %s 中，拒绝访问", company, rule.RuleName)
				return false, nil
			}
		}

		// 检查白名单规则,白名单优先
		if rule.RuleType == "White" {
			hasWhiteRule = true
			// 检查公司是否在白名单中
			if s.isCompanyInList(rule.CompanyList, company) {
				inWhiteList = true
				g.Log().Infof(ctx, "公司 %s 在白名单规则 %s 中，允许访问", company, rule.RuleName)
			}
		}
	}

	// 如果存在白名单规则，则只有在白名单中的公司才能访问
	if hasWhiteRule {
		return inWhiteList, nil
	}

	// 如果没有白名单规则，且没有被黑名单拒绝，则允许访问
	return true, nil
}

// isCompanyInList 检查公司是否在公司列表中
// 参数：
//   - companyList: 公司列表
//   - company: 要检查的公司名称
//
// 返回：
//   - bool: 是否在列表中
func (s *sPermission) isCompanyInList(companyList []erp.PermissionCompanyList, company string) bool {
	for _, companyItem := range companyList {
		if companyItem.Company == company {
			return true
		}
	}
	return false
}
