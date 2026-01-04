package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/do"
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	setup2 "ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
	"ttpos-bmp/app/ttpos-erp/internal/service"
	"ttpos-bmp/utility"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

var (
	Setup = new(sSetup)
)

type sSetup struct{}

func init() {
	service.RegisterSetup(Setup)
}

// pruneNumericPrefix 删除目录名称中的数字前缀
// 例如: "00_custom_fields" -> "custom_fields"
// 参数：
//   - dirName: 原始目录名称
//
// 返回：
//   - string: 处理后的目录名称
func (s *sSetup) pruneNumericPrefix(dirName string) string {
	// 匹配以数字开头，后跟下划线的模式，如 "00_", "1_", "123_"
	re := regexp.MustCompile(`^\d+_`)
	return re.ReplaceAllString(dirName, "")
}

// CreateBranch 创建分店
// 参数：
//   - ctx: 上下文对象
//   - req: 初始化店铺请求参数
//
// 返回：
//   - branchName: 分店名称
//   - err: 错误信息
func (s *sSetup) CreateBranch(ctx context.Context, req *setup.InitShopReq) (branchName string, err error) {
	// 参数验证
	if err := s.validateCreateBranchReq(req); err != nil {
		return "", err
	}
	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: erp.DocTypeBranch,
	}, &erp.RequestParams{
		Filters: [][]string{{"branch", "=", req.ShopName}},
	})
	if err != nil {
		g.Log().Error(ctx, "查询分支失败", err)
		return "", gerror.Wrap(err, "查询分支失败")
	}

	if count == 0 {
		// 获取公司信息
		company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
		if err != nil {
			g.Log().Error(ctx, "获取公司信息失败", err)
			return "", gerror.Wrap(err, "获取公司信息失败")
		}

		// 创建分支
		branchPayload := g.Map{
			"branch":          req.ShopName,
			"custom_company":  company.CompanyName,
			"custom_shopuuid": req.ShopUuid,
		}

		if _, err := service.Document().Create(ctx, erp.DocTypeBranch, branchPayload); err != nil {
			g.Log().Error(ctx, "创建分支失败", err)
			return "", gerror.Wrapf(err, "创建分支失败")
		}
	}
	//默认忽略已存在，如果不行再改这里
	//else {
	//
	//	// 店铺已存在，根据ignoreExists参数判断是否忽略
	//	//if !ignoreExists {
	//	//	return "", gerror.New("店铺已初始化，不能重复初始化")
	//	//}
	//}

	return req.ShopName, nil
}

// validateCreateBranchReq 验证创建分支请求参数
func (s *sSetup) validateCreateBranchReq(req *setup.InitShopReq) error {
	if strings.TrimSpace(req.ShopUuid) == "" {
		return gerror.New("店铺UUID不能为空")
	}
	if strings.TrimSpace(req.ShopName) == "" {
		return gerror.New("店铺名称不能为空")
	}
	if strings.TrimSpace(req.CompanyAbbr) == "" {
		return gerror.New("公司缩写编码不能为空")
	}
	return nil
}

// CreateUser 创建网站用户，门店收银账户
// 参数：
//   - ctx: 上下文对象
//   - req: 创建用户请求参数
//
// 返回：
//   - err: 错误信息
func (s *sSetup) CreateUser(ctx context.Context, req *setup2.CreateUserInp) error {
	userPayload := g.Map{
		"email":              req.UserEmail,
		"first_name":         req.FirstName,
		"user_type":          "Website User",
		"enabled":            1,
		"send_welcome_email": false,
		"role_profile_name":  "ShopCashier",
		"time_zone":          "Asia/Bangkok",
	}

	if _, err := service.Document().Create(ctx, "User", userPayload); err != nil {
		g.Log().Error(ctx, "创建网站用户失败", err)
		return gerror.Wrapf(err, "创建网站用户失败")
	}

	return nil
}

// CreateDefaultPosProfile 创建默认的POS配置文件
// 参数：
//   - ctx: 上下文对象
//   - req: 创建默认POS配置文件请求参数
//
// 返回：
//   - posProfileName: POS配置文件名称
//   - err: 错误信息
func (s *sSetup) CreateDefaultPosProfile(ctx context.Context, req *setup.CreateDefaultPosProfileReq) (posProfileName string, err error) {

	posProfileList, err := service.Selling().GetPosProfileList(ctx, &selling.PosProfileReq{
		Name: req.Name,
	})
	if err != nil {
		g.Log().Error(ctx, "获取POS配置文件列表失败", err)
		return "", gerror.Wrap(err, "获取POS配置文件列表失败")
	}
	if len(posProfileList.ProfileList) > 0 {
		g.Log().Info(ctx, "POS配置文件已存在，无需创建", req.Name)
		return req.Name, nil
	}
	// 获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		g.Log().Error(ctx, "获取公司信息失败", err)
		return "", gerror.Wrap(err, "获取公司信息失败")
	}

	// 获取默认仓库
	warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, companyName, req.Branch)
	if err != nil {
		return "", gerror.Wrapf(err, "获取默认仓库失败")
	}

	reqInfo := &setup2.CreatePosProfileInp{
		PosProfileName:     req.Name,
		CompanyAbbr:        req.CompanyAbbr,
		Warehouse:          warehouse.Name,
		Branch:             req.Branch,
		Payments:           g.ArrayStr{"Cash", "Balance", "Free Meal"},
		Currency:           "THB", // 泰铢
		WriteOffAccount:    "Sales - " + req.CompanyAbbr,
		WriteOffLimit:      100000.00, //销账约束
		WriteOffCostCenter: "Main - " + req.CompanyAbbr,
	}
	if len(req.Cashiers) > 0 {
		reqInfo.ApplicableForUsers = make([]setup2.Cashier, 0)
		for _, user := range req.Cashiers {
			reqInfo.ApplicableForUsers = append(reqInfo.ApplicableForUsers, setup2.Cashier{
				User: user.User,
			})
		}
	}
	// 创建POS配置文件
	posProfile, err := service.Selling().CreatePosProfile(ctx, reqInfo)
	if err != nil {
		return "", gerror.Wrapf(err, "创建POS配置文件失败")
	}

	return posProfile.Name, nil
}

// InitShop 初始化店铺
// 参数：
//   - ctx: 上下文对象
//   - req: 初始化店铺请求参数
//
// 返回：
//   - resp: 初始化店铺响应结果
//   - err: 错误信息
func (s *sSetup) InitShop(ctx context.Context, req *setup.InitShopReq) (resp *setup.InitShopResp, err error) {
	var (
		branchName  string
		adminEmail  = fmt.Sprintf("%s@ttpos-user.com", req.AdminUuid)
		companyName string
		siteCode    string
	)

	// 从 ctx 中获取站点编码
	siteCode = service.Rpc().GetSiteCode(ctx)
	g.Log().Info(ctx, "获取站点编码", g.Map{
		"site_code": siteCode,
	})

	// 根据公司缩写获取公司名称
	companyName, err = service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		g.Log().Error(ctx, "查询公司名称失败", g.Map{
			"company_abbr": req.CompanyAbbr,
			"error":        err,
		})
		return nil, gerror.Wrapf(err, "查询公司失败")
	}

	// 创建分支
	branchName, err = s.CreateBranch(ctx, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建店铺失败")
	}

	cashierCount, err := dao.ShopCashier.Ctx(ctx).Count(dao.ShopCashier.Columns().ShopUuid, req.ShopUuid)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询门店收银账户失败")
	}
	if cashierCount > 0 {
		g.Log().Infof(ctx, "收银员已初始化,%s", req.ShopUuid)
		//return nil, gerror.New("收银员已初始化，不能重复初始化")
	}

	if cashierCount == 0 {
		// 创建默认用户
		err = s.CreateUser(ctx, &setup2.CreateUserInp{
			UserEmail: adminEmail,
			FirstName: req.ShopName,
			AdminUuid: req.AdminUuid,
			ShopUuid:  req.ShopUuid,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "创建用户失败")
		}

		// 获取用户的API_KEY/API_SECRET
		apiKey, apiSecret, err := s.GetUserApiKeySecret(ctx, adminEmail)
		if err != nil {
			return nil, gerror.Wrapf(err, "获取用户API_KEY/API_SECRET失败")
		}

		// 记录门店管理员关联关系，记录api_key/api_secret
		_, err = dao.ShopCashier.Ctx(ctx).Insert(&do.ShopCashier{
			ShopUuid:     req.ShopUuid,
			AdminUuid:    req.AdminUuid,
			CashierEmail: adminEmail,
			ApiKey:       apiKey,
			ApiSecret:    apiSecret,
			CompanyAbbr:  req.CompanyAbbr,
			Branch:       branchName,
			SiteCode:     siteCode,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "创建门店管理员关联关系失败")
		}
	}

	// 创建默认仓库
	warehouseList, err := service.Warehouse().GetWarehouseList(ctx, &warehouse.GetWarehouseListReq{
		Branch:      branchName,
		CompanyAbbr: req.CompanyAbbr,
		AliasName:   "Default",
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取仓库列表失败")
	}
	if len(warehouseList.WarehouseList) == 0 {
		_, err = service.Warehouse().CreateWarehouse(ctx, &setup2.CreateWarehouseInp{
			Branch:      branchName,
			WhType:      "Normal",
			AliasName:   "Default",
			CompanyAbbr: req.CompanyAbbr,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "创建默认仓库失败")
		}
	}
	// 创建在途仓
	warehouseList, err = service.Warehouse().GetWarehouseList(ctx, &warehouse.GetWarehouseListReq{
		Branch:      branchName,
		CompanyAbbr: req.CompanyAbbr,
		AliasName:   "Transit",
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取仓库列表失败")
	}
	if len(warehouseList.WarehouseList) == 0 {
		_, err = service.Warehouse().CreateWarehouse(ctx, &setup2.CreateWarehouseInp{
			Branch:      branchName,
			WhType:      "Transit",
			AliasName:   "Transit",
			CompanyAbbr: req.CompanyAbbr,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "创建在途仓失败")
		}
	}
	// 创建默认的Cash，Balance，Free Meal账号关联
	if err = s.createPaymentAccounts(ctx, req.CompanyAbbr); err != nil {
		return nil, gerror.Wrapf(err, "创建支付方式账号关联失败")
	}

	//创建默认pos profile
	posProfileName, err := s.CreateDefaultPosProfile(ctx, &setup.CreateDefaultPosProfileReq{
		Name:        fmt.Sprintf("%s - POS", branchName),
		CompanyAbbr: req.CompanyAbbr,
		Branch:      branchName,
		Cashiers: []*setup.PosProfileCashier{
			{
				User: adminEmail,
			},
		},
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建默认pos profile失败")
	}

	//FIXME 目前除了 ttpos site 其他都是连锁
	ctxMap := grpcx.Ctx.IncomingMap(ctx)
	if ctxMap.Contains(consts.ContextSiteCode) && ctxMap.Get(consts.ContextSiteCode) != consts.SiteCodeTtpos {
		//连锁店模式下，将本公司增加至总店供应商内部交易对象
		err = service.Supplier().AddSupplerTransactCompany(ctx, &dto.AddSupplerTransactCompanyReq{
			Supplier:        erp.HeadquartersSupplier,
			WithCompanyAbbr: req.CompanyAbbr,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "添加供应商内部交易对象失败")
		}

		//为实现调拨功能，每个新增的门店都属于内部供应商
		supplierCount, err := service.Supplier().CountSupplier(ctx, &erp.Supplier{
			IsInternalSupplier: true,
			RepresentsCompany:  companyName,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "查询供应商数量失败")
		}
		if supplierCount > 0 {
			g.Log().Infof(ctx, "供应商已初始化,%s", branchName)
			//return nil, gerror.New("供应商已初始化，不能重复初始化")
		} else {
			_, err = service.Supplier().CreateSupplier(ctx, &buying.CreateSupplierReq{
				Supplier: &buying.SupplierData{
					AliasName:          branchName,
					Branch:             branchName,
					CompanyAbbr:        req.CompanyAbbr,
					IsInternalSupplier: true,
				},
			})
			if err != nil {
				return nil, gerror.Wrapf(err, "创建供应商失败")
			}
		}

	}

	return &setup.InitShopResp{
		BranchName: branchName,
		AdminEmail: adminEmail,
		PosProfile: posProfileName,
	}, nil
}

// createPaymentAccounts 创建支付方式账号关联
func (s *sSetup) createPaymentAccounts(ctx context.Context, companyAbbr string) error {
	// 创建Cash支付方式账号关联
	if err := service.Selling().CreateModePaymentAccount(ctx, &setup2.CreateModePaymentAccountInp{
		CompanyAbbr: companyAbbr,
		PaymentType: string(consts.ModeOfPaymentCash),
	}); err != nil {
		return gerror.Wrapf(err, "创建Cash支付方式账号关联失败")
	}

	// 创建Balance支付方式账号关联
	if err := service.Selling().CreateModePaymentAccount(ctx, &setup2.CreateModePaymentAccountInp{
		CompanyAbbr: companyAbbr,
		PaymentType: string(consts.ModeOfPaymentBalance),
	}); err != nil {
		return gerror.Wrapf(err, "创建Balance支付方式账号关联失败")
	}

	// 创建Free Meal支付方式账号关联
	if err := service.Selling().CreateModePaymentAccount(ctx, &setup2.CreateModePaymentAccountInp{
		CompanyAbbr: companyAbbr,
		PaymentType: string(consts.ModeOfPaymentFreeMeal),
	}); err != nil {
		return gerror.Wrapf(err, "创建Free Meal支付方式账号关联失败")
	}

	return nil
}

// GetUserApiKeySecret 获取用户的API密钥和密钥
// 参数：
//   - ctx: 上下文对象
//   - userEmail: 用户邮箱
//
// 返回：
//   - apiKey: API密钥
//   - apiSecret: API密钥
//   - err: 错误信息
func (s *sSetup) GetUserApiKeySecret(ctx context.Context, userEmail string) (apiKey, apiSecret string, err error) {
	// 生成API密钥
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: "frappe.core.doctype.user.user.generate_keys",
	}, g.Map{
		"user": userEmail,
	})
	if err != nil {
		return "", "", gerror.Wrapf(err, "获取用户API密钥失败")
	}

	// 解析响应数据获取API密钥
	j := resp

	apiSecret = j.Get("data.api_secret").String()

	// 获取用户信息
	resp, err = service.Document().Get(ctx, &erp.ErpReq{
		DocType: "User",
		Name:    userEmail,
	}, nil)
	if err != nil {
		return "", "", gerror.Wrapf(err, "获取用户信息失败")
	}

	// 解析响应数据获取API密钥
	j = resp
	apiKey = j.Get("data.api_key").String()

	return apiKey, apiSecret, nil
}

// DocumentInitConfig 文档初始化配置
type DocumentInitConfig struct {
	DirBase    string // 迁移目录基础路径
	DirName    string // 目录名称，如 "custom_fields"
	DocType    string // 文档类型，如 "Custom Field"
	ItemName   string // 项目名称，用于日志，如 "自定义字段"
	ItemNameEn string // 项目英文名称，用于错误信息，如 "custom field"
}

// initDocumentsFromDir 通用的文档初始化方法
// 遍历指定目录下的所有JSON文件，解析并创建或更新对应类型的文档
// 如果JSON数据中包含非空的name字段，则更新已有文档；否则创建新文档
// 参数：
//   - ctx: 上下文对象
//   - config: 初始化配置
//
// 返回：
//   - err: 错误信息
func (s *sSetup) initDocumentsFromDir(ctx context.Context, config DocumentInitConfig) error {
	// 构建目录路径
	dirPath := fmt.Sprintf("%s/%s", config.DirBase, config.DirName)

	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return gerror.Newf("%s目录不存在: %s", config.ItemName, dirPath)
	}

	g.Log().Infof(ctx, "开始初始化%s，目录: %s", config.ItemName, dirPath)

	// 遍历目录下的所有文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理JSON文件
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			g.Log().Infof(ctx, "处理%s文件: %s", config.ItemName, path)

			// 读取JSON文件内容
			data, err := os.ReadFile(path)
			if err != nil {
				g.Log().Error(ctx, fmt.Sprintf("读取%s文件失败", config.ItemName), err, g.Map{"file": path})
				return gerror.Wrapf(err, "读取文件失败: %s", path)
			}

			// 解析JSON数据
			var docData map[string]interface{}
			if err := json.Unmarshal(data, &docData); err != nil {
				g.Log().Error(ctx, fmt.Sprintf("解析%sJSON失败", config.ItemName), err, g.Map{"file": path})
				return gerror.Wrapf(err, "解析JSON失败: %s", path)
			}

			// 检查 docData 中的 name 字段，智能判断是创建还是更新
			docName, hasName := docData["name"].(string)

			if hasName && docName != "" {
				// name 不为空，调用 Update 方法更新文档
				if _, err := service.Document().Update(ctx, &erp.ErpReq{
					DocType: config.DocType,
					Name:    docName,
				}, docData); err != nil {
					g.Log().Error(ctx, fmt.Sprintf("更新%s失败", config.ItemName), err, g.Map{"file": path, "data": docData})
				} else {
					g.Log().Infof(ctx, "%s更新成功: %s", config.ItemName, path)
				}
			} else {
				// name 为空，调用 Create 方法创建文档
				if _, err := service.Document().Create(ctx, config.DocType, docData); err != nil {
					g.Log().Error(ctx, fmt.Sprintf("创建%s失败", config.ItemName), err, g.Map{"file": path, "data": docData})
				} else {
					g.Log().Infof(ctx, "%s创建成功: %s", config.ItemName, path)
				}
			}
		}

		return nil
	})

	if err != nil {
		return gerror.Wrapf(err, "初始化%s失败", config.ItemName)
	}

	g.Log().Infof(ctx, "所有%s初始化完成", config.ItemName)
	return nil
}

func (s *sSetup) InitErpDocTypeWithDirname(ctx context.Context, dirBase string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dirBase); os.IsNotExist(err) {
		return gerror.Newf("%s目录不存在", dirBase)
	}
	g.Log().Infof(ctx, "开始初始化%s", dirBase)

	// 读取目录中的所有文件
	files, err := os.ReadDir(dirBase)
	if err != nil {
		return gerror.Wrapf(err, "读取目录失败")
	}

	// 过滤出目录并按名称排序
	var dirs []string
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}

	// 按照目录名称进行排序
	sort.Strings(dirs)

	// 按照排序后的顺序处理每个目录
	for _, dirName := range dirs {
		// 删除数字前缀（如 "00_"）
		prunedDirName := s.pruneNumericPrefix(dirName)
		g.Log().Infof(ctx, "处理目录: %s -> %s", dirName, prunedDirName)

		config := DocumentInitConfig{
			DirBase:  dirBase,
			DirName:  dirName, // 保持原始目录名称用于文件路径
			DocType:  utility.ConvertToTitleCase(prunedDirName),
			ItemName: utility.ConvertToTitleCase(prunedDirName),
		}
		g.Log().Infof(ctx, "开始初始化%s，目录: %s", config.ItemName, prunedDirName)

		// 处理单个目录，记录错误但不中断其他目录的处理
		if err := s.initDocumentsFromDir(ctx, config); err != nil {
			g.Log().Errorf(ctx, "初始化%s失败: %v", config.ItemName, err)
			//return gerror.Wrapf(err, "初始化%s失败", config.ItemName)
		}
	}

	g.Log().Infof(ctx, "所有目录初始化完成")
	return nil
}

// InitCustomFields 初始化自定义字段
// 遍历manifest/erp-migrate/v2.5/custom_fields目录下所有JSON文件，创建Custom Field文档
// 参数：
//   - ctx: 上下文对象
//
// 返回：
//   - err: 错误信息
func (s *sSetup) InitCustomFields(ctx context.Context, dirBase string) error {
	return s.initDocumentsFromDir(ctx, DocumentInitConfig{
		DirBase:    dirBase,
		DirName:    "custom_fields",
		DocType:    "Custom Field",
		ItemName:   "自定义字段",
		ItemNameEn: "custom field",
	})
}

// InitCustomers 初始化客户
// 遍历manifest/erp-migrate/v2.5/new_customer目录下所有JSON文件，创建Customer文档
// 参数：
//   - ctx: 上下文对象
//
// 返回：
//   - err: 错误信息
func (s *sSetup) InitCustomers(ctx context.Context, dirBase string) error {
	return s.initDocumentsFromDir(ctx, DocumentInitConfig{
		DirBase:    dirBase,
		DirName:    "customer",
		DocType:    "Customer",
		ItemName:   "客户",
		ItemNameEn: "customer",
	})
}

// InitModeOfPayment 初始化支付方式
// 遍历manifest/erp-migrate/v2.5/mode_of_payment目录下所有JSON文件，创建Mode Of Payment文档
// 参数：
//   - ctx: 上下文对象
//
// 返回：
//   - err: 错误信息
func (s *sSetup) InitModeOfPayment(ctx context.Context, dirBase string) error {
	return s.initDocumentsFromDir(ctx, DocumentInitConfig{
		DirBase:    dirBase,
		DirName:    "mode_of_payment",
		DocType:    "Mode Of Payment",
		ItemName:   "支付方式",
		ItemNameEn: "mode of payment",
	})
}

// GetDefaultTimeZone 获取默认时区
// 参数：
//   - ctx: 上下文对象
//
// 返回：
//   - string: 时区
func (s *sSetup) GetDefaultTimeZone(ctx context.Context) string {
	return "Asia/Shanghai"
}

// MustGetLocalDateTime 获取本地时间
// 参数：
//   - ctx: 上下文对象
//   - utcTime: UTC时间
//
// 返回：
//   - *gtime.Time: 本地时间
//
// 注意：
//   - 如果转换失败，会返回UTC时间
func (s *sSetup) MustGetLocalDateTime(ctx context.Context, utcTime *gtime.Time) *gtime.Time {
	localTime, err := utcTime.ToZone(s.GetDefaultTimeZone(ctx))
	if err != nil {
		g.Log().Error(ctx, "转换时区失败", err, g.Map{"utcTime": utcTime})
		return utcTime
	}
	return localTime
}
