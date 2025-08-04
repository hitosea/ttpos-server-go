package service

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IStaffSrv interface {
	// 获取管理员列表
	GetStaffs(ctx context.Context, pageReq dto.PageReq) (resp.StaffListPaginationResp, error)
	// 修改管理员
	UpdateStaff(ctx context.Context, updateReq req.UpdateStaffReq) error
	// 设置启用禁用员工
	UpdateStaffStatus(ctx context.Context, updateReq req.UpdateStaffStatusReq) error
	// 删除员工
	DeleteStaff(ctx context.Context, deleteReq req.DeleteStaffReq) error
}

type staffSrv struct {
	dbm *database.DBManager
}

func NewStaffSrvImpl(dbm *database.DBManager) IStaffSrv {
	return &staffSrv{
		dbm: dbm,
	}
}

func NewStaffSrv(dbm *database.DBManager) IStaffSrv {
	return NewStaffSrvImpl(dbm)
}

func (s *staffSrv) GetStaffs(ctx context.Context, pageReq dto.PageReq) (resp.StaffListPaginationResp, error) {
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(ctx.GetDbId()))

	staffs, total, err := staffRepo.PaginateGetStaffs(pageReq.PageNo, pageReq.PageSize, staffRepo.WithRoles())
	if err != nil {
		return resp.StaffListPaginationResp{}, err
	}

	staffList := make([]resp.Staff, 0, len(staffs))

	for _, staff := range staffs {

		logger.Logger.Info("staff", zap.Any("staff", staff))

		roles := make([]resp.StaffRole, 0, len(staff.Roles))
		for _, role := range staff.Roles {
			roles = append(roles, resp.StaffRole{
				Uuid: role.Uuid,
				Name: role.Name,
			})
		}
		staffList = append(staffList, resp.Staff{
			Uuid:       staff.Uuid,
			Username:   staff.Username,
			RealName:   staff.RealName,
			Roles:      roles,
			IsDisable:  staff.IsDisable,
			IsSuper:    staff.IsSuper,
			CreateTime: staff.CreateTime,
		})
	}
	return resp.StaffListPaginationResp{
		List: staffList,
		Meta: dto.PageResponse{
			PageNo:   pageReq.PageNo,
			PageSize: pageReq.PageSize,
			Total:    total,
		},
	}, nil
}

// 修改管理员
func (s *staffSrv) UpdateStaff(ctx context.Context, updateReq req.UpdateStaffReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取管理员
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(updateReq.Uuid))
	if err != nil {
		return err
	}
	if staff.IsSuper == 1 {
		return errors.New("超级管理员不能修改")
	}
	// 判断角色参数是否正确
	roleRepo := repository.NewRoleRepo(db)
	roles, err := roleRepo.GetRoleList([]repository.DBOption{roleRepo.WhereUuids(updateReq.Roles)}...)
	if err != nil {
		return err
	}
	if len(roles) != len(updateReq.Roles) {
		return errors.New("角色参数错误")
	}
	update := map[string]any{
		"username":  updateReq.Username,
		"real_name": updateReq.RealName,
		"phone":     updateReq.Phone,
	}
	if updateReq.Password != "" {
		update["password"] = utils.EncryptPassword(updateReq.Password)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		staffRepo := repository.NewStaffRepo(tx)
		err = staffRepo.Update(updateReq.Uuid, update)
		if err != nil {
			return err
		}
		// 更新管理员角色
		err = staffRepo.UpdateStaffRoles(updateReq.Uuid, updateReq.Roles)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "更新管理员失败")
	}
	return nil
}

func (s *staffSrv) UpdateStaffStatus(ctx context.Context, updateReq req.UpdateStaffStatusReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(updateReq.Uuid))
	if err != nil {
		return err
	}
	if staff.IsSuper == 1 {
		return errors.New("超级管理员不能修改")
	}

	statusMap := map[int]int{
		1: 0,
		0: 1,
	}
	err = staffRepo.Update(updateReq.Uuid, map[string]any{
		"is_disable": statusMap[*updateReq.Status],
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *staffSrv) DeleteStaff(ctx context.Context, deleteReq req.DeleteStaffReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(deleteReq.Uuid))
	if err != nil {
		return err
	}
	if staff.IsSuper == 1 {
		return errors.New("超级管理员不能删除")
	}

	err = db.Model(&model.Staff{}).Where("uuid = ?", deleteReq.Uuid).Delete(&model.Staff{}).Error
	if err != nil {
		return err
	}
	return nil
}
