package service

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/jinzhu/copier"

	"jjjshop-server-go/app/constant"
	"jjjshop-server-go/app/dto/resp"
	"jjjshop-server-go/app/model"
	"jjjshop-server-go/app/repository"
)

type RoleAccessService struct {
	shopUserRoleRepo *repository.UserRoleRepository
	shopAccessRepo   *repository.AccessRepository
	shopUserRepo     *repository.UserRepository
}

func NewRoleAccessService(userRoleRepo *repository.UserRoleRepository, shopAccessRepo *repository.AccessRepository, shopUserRepo *repository.UserRepository) *RoleAccessService {
	return &RoleAccessService{
		shopUserRoleRepo: userRoleRepo,
		shopAccessRepo:   shopAccessRepo,
		shopUserRepo:     shopUserRepo,
	}
}

// GetPermission 获取权限
func (s *RoleAccessService) GetPermission(isShow bool, routerName constant.RouteName, shopUserId uint) ([]*resp.Permission, error) {

	shopUser := s.shopUserRepo.GetById(shopUserId, s.shopUserRepo.WithSupplier())

	var permissions []resp.Permission
	var where []repository.Where

	if shopUser.IsSuper == 1 { // 超级管理员
		if shopUser.UserType == 1 {
			if shopUser.Supplier.CategorySet == 10 {
				where = append(where, s.shopAccessRepo.WherePath([]string{"/product/takeaway/category/index", "/product/store/category/index"})) // ToDo 修改为具体值
			}
			where = append(where, s.shopAccessRepo.WhereIsSupplier())
		}
	} else {
		roleIds, err := s.shopUserRoleRepo.GetRoleIds(shopUser.ShopUserId, shopUser.AppId)
		if err != nil {
			return nil, errors.New("获取用户角色失败")
		}
		accessIds, err := s.shopAccessRepo.GetAccessIds(roleIds, shopUser.AppId)
		if err != nil {
			return nil, errors.New("获取角色权限失败")
		}
		where = append(where, s.shopAccessRepo.WhereIds(accessIds))
	}

	dbPermissions, err := s.shopAccessRepo.GetPermissions(shopUser.AppId, where...)
	if err != nil {
		return nil, errors.New("获取权限失败")
	}

	for _, dbPermission := range dbPermissions {
		var permission resp.Permission
		copier.Copy(&permission, dbPermission)
		permission.CreateTime = time.Unix(int64(dbPermission.CreateTime), 0).Format("2006-01-02 15:04:05")
		permission.UpdateTime = time.Unix(int64(dbPermission.UpdateTime), 0).Format("2006-01-02 15:04:05")
		permission.Children = []*resp.Permission{}
		permissions = append(permissions, permission)
	}

	permissions = s.filterPermission(permissions, *shopUser.Supplier)

	return s.buildPermissionTree(permissions, routerName), nil
}

// filterPermission 筛选权限
func (s *RoleAccessService) filterPermission(permissions []resp.Permission, supplier model.Supplier) []resp.Permission {
	var filteredPermissions []resp.Permission
	for _, permission := range permissions {
		// 暂时去掉外卖管理
		if permission.AccessId == 1626688443 {
			continue
		}
		// 授权无进销存权限
		if supplier.SaleStock == 0 && slices.Contains([]int{1711006072, 1711009130}, permission.AccessId) {
			continue
		}
		// 授权无会员权限
		if supplier.IsOpenMember == 0 && slices.Contains([]int{1636183779, 1704881218}, permission.AccessId) {
			continue
		}
		// 授权无平板点餐权限
		if supplier.IsOpenTablet == 0 && permission.AccessId == 87 {
			continue
		}
		// 授权无H5点餐权限
		if supplier.IsOpenScan == 0 && permission.AccessId == 1724220505 {
			continue
		}
		// 授权无点餐助手权限
		if supplier.IsOpenAssistant == 0 && permission.AccessId == 1720753338 {
			continue
		}
		// 授权无后厨权限
		if supplier.IsOpenKitchenKds == 0 && permission.AccessId == 88 {
			continue
		}
		// 授权无自助餐权限
		if supplier.IsOpenBuffet == 0 && permission.AccessId == 1708671616 {
			continue
		}
		// 授权无扫码点餐接单权限
		if supplier.IsAcceptScanOrder == 0 && permission.AccessId == 1724320522 {
			continue
		}
		filteredPermissions = append(filteredPermissions, permission)
	}
	return filteredPermissions
}

// buildPermissionTree 构建权限树
func (s *RoleAccessService) buildPermissionTree(permissions []resp.Permission, routerName constant.RouteName) []*resp.Permission {
	permissionMap := make(map[int]*resp.Permission)
	var roots []*resp.Permission

	var accessIds []string

	// 第一步：建立ID到节点的映射
	for i := range permissions {
		permission := &permissions[i]
		permissionMap[permission.AccessId] = permission

		// fmt.Printf("%03d%010d\n", 99, 10)
		accessIds = append(accessIds, fmt.Sprintf("%03d%010d\n", permission.Sort, permission.AccessId))

	}

	sort.Strings(accessIds)
	// 第二步：构建树结构

	for _, accessId := range accessIds {
		for _, permission := range permissionMap {
			if fmt.Sprintf("%03d%010d\n", permission.Sort, permission.AccessId) != accessId {
				continue
			}
			if parent, exists := permissionMap[permission.ParentId]; exists {
				parent.Children = append(parent.Children, permission)
			} else {
				roots = append(roots, permission)
			}
		}
	}

	var filteredRoots []*resp.Permission
	for _, root := range roots {
		if root.Name == string(routerName) {
			filteredRoots = append(filteredRoots, root.Children...)
		}
	}

	return filteredRoots
}
