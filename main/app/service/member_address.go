package service

import (
	"time"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/dto/resp/member_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/validator"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// IMemberAddressSrv 定义会员地址服务接口
type IMemberAddressSrv interface {
	GetAddressList(ctx context.Context, req member_req.MemberAddressListReq) (*member_resp.MemberAddressListResp, error)       // 获取地址列表
	GetAddressDetail(ctx context.Context, req member_req.MemberAddressDetailReq) (*member_resp.MemberAddressDetailResp, error) // 获取地址详情
	AddAddress(ctx context.Context, req member_req.MemberAddressAddReq) error                                                  // 添加地址
	UpdateAddress(ctx context.Context, req member_req.MemberAddressUpdateReq) error                                            // 更新地址
	DeleteAddress(ctx context.Context, req member_req.MemberAddressDeleteReq) error                                            // 删除地址
	AuthAddress(ctx context.Context, req member_req.MemberAddressAuthReq) (member_resp.LoginResp, error)                       // 认证地址
}

// memberAddressSrv 会员地址服务结构体
type memberAddressSrv struct {
	dbm   *database.DBManager
	bus   *event.SystemEventBus
	cache cache.Cache
}

// NewMemberAddressSrv 创建新的会员地址服务
func NewMemberAddressSrv(dbm *database.DBManager, cache cache.Cache) IMemberAddressSrv {
	return NewMemberAddressSrvImpl(dbm, cache)
}

// NewMemberAddressSrvImpl 创建新的会员地址服务实现
func NewMemberAddressSrvImpl(dbm *database.DBManager, cache cache.Cache) IMemberAddressSrv {
	return &memberAddressSrv{
		dbm:   dbm,
		bus:   event.NewSystemBus(),
		cache: cache,
	}
}

// GetAddressList 获取会员地址列表
func (s *memberAddressSrv) GetAddressList(ctx context.Context, req member_req.MemberAddressListReq) (*member_resp.MemberAddressListResp, error) {
	memberAddressRepo := repository.NewMemberAddressRepo(ctx.GetDB())
	memberAddresses, total, err := memberAddressRepo.PaginateGet(
		req.PageNo, req.PageSize,
		repository.CommonRepo.WhereByMemberUuid(ctx.GetMemberUuid()),
	)
	if err != nil {
		return nil, err
	}
	respMemberAddresses := make([]member_resp.MemberAddressResp, 0)
	for _, memberAddress := range memberAddresses {
		var respMemberAddress member_resp.MemberAddressResp
		copier.Copy(&respMemberAddress, memberAddress)
		respMemberAddress.IsAuthPhone = memberAddress.IsAuthPhone()
		respMemberAddress.IsDefault = memberAddress.IsDefault == 1
		respMemberAddresses = append(respMemberAddresses, respMemberAddress)
	}
	return &member_resp.MemberAddressListResp{
		List: respMemberAddresses,
		Meta: dto.PageResponse{
			Total:    total,
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
		},
	}, nil
}

// GetAddressDetail 获取地址详情
func (s *memberAddressSrv) GetAddressDetail(ctx context.Context, req member_req.MemberAddressDetailReq) (*member_resp.MemberAddressDetailResp, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.New(err.Error())
	}

	memberAddressRepo := repository.NewMemberAddressRepo(ctx.GetDB())
	memberAddress, err := memberAddressRepo.GetMemberAddressByUuid(req.Uuid)
	if err != nil {
		return nil, errors.New("地址不存在")
	}

	// 验证地址归属权
	if memberAddress.MemberUuid != ctx.GetMemberUuid() {
		return nil, errors.New("地址不存在")
	}

	// 构建响应数据
	var respMemberAddress member_resp.MemberAddressDetailResp
	copier.Copy(&respMemberAddress, memberAddress)
	respMemberAddress.IsDefault = memberAddress.IsDefault == 1
	respMemberAddress.IsAuthPhone = memberAddress.IsAuthPhone()

	// 设置地址详情
	respMemberAddress.AddressDetail = memberAddress.GetAddressDetail()

	return &respMemberAddress, nil
}

// AddAddress 添加地址
func (s *memberAddressSrv) AddAddress(ctx context.Context, req member_req.MemberAddressAddReq) error {
	if err := req.Validate(); err != nil {
		return errors.New(err.Error())
	}
	// 复制请求参数到会员地址
	var memberAddress model.MemberAddress
	copier.Copy(&memberAddress, req)
	memberAddress.MemberUuid = ctx.GetMemberUuid()

	// 如果国家代码为空，则根据手机号判断
	if memberAddress.Country == "" {
		if len(memberAddress.Phone) == 11 {
			memberAddress.Country = "+86"
		} else {
			memberAddress.Country = "+66"
		}
	}

	// 如果设置为默认地址，则需要将该会员的其他地址设置为非默认
	if req.IsDefault {
		// 开启事务
		err := ctx.GetDB().Transaction(func(tx *gorm.DB) error {
			// 先将该会员的所有地址设置为非默认
			if err := repository.NewMemberAddressRepo(tx).UpdateIsDefault(ctx.GetMemberUuid()); err != nil {
				return err
			}

			// 创建新地址
			_, err := repository.NewMemberAddressRepo(tx).Create(memberAddress)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	}

	// 如果不是设置为默认地址，则直接创建
	_, err := repository.NewMemberAddressRepo(ctx.GetDB()).Create(memberAddress)
	if err != nil {
		return err
	}

	return nil
}

// UpdateAddress 更新地址
func (s *memberAddressSrv) UpdateAddress(ctx context.Context, req member_req.MemberAddressUpdateReq) error {
	if err := req.Validate(); err != nil {
		return errors.New(err.Error())
	}

	memberAddressRepo := repository.NewMemberAddressRepo(ctx.GetDB())
	memberAddress, err := memberAddressRepo.GetMemberAddressByUuid(req.Uuid)
	if err != nil {
		return errors.New("地址不存在")
	}
	if memberAddress.MemberUuid != ctx.GetMemberUuid() {
		return errors.New("地址不存在")
	}

	// 复制请求参数到会员地址
	copier.Copy(&memberAddress, req)
	memberAddress.MemberUuid = ctx.GetMemberUuid()

	// 如果国家代码为空，则根据手机号判断
	if memberAddress.Country == "" {
		if len(memberAddress.Phone) == 11 {
			memberAddress.Country = "+86"
		} else {
			memberAddress.Country = "+66"
		}
	}

	// 如果设置为默认地址，则将该会员的其他地址设置为非默认
	if req.IsDefault {
		// 开启事务
		err = ctx.GetDB().Transaction(func(tx *gorm.DB) error {
			// 先将该会员的所有地址设置为非默认
			if err := repository.NewMemberAddressRepo(tx).UpdateIsDefault(ctx.GetMemberUuid()); err != nil {
				return err
			}

			// 更新当前地址
			_, err := repository.NewMemberAddressRepo(tx).Update(*memberAddress)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	}

	// 如果不是设置为默认地址，则直接更新
	_, err = memberAddressRepo.Update(*memberAddress)
	if err != nil {
		return err
	}
	return nil
}

// DeleteAddress 删除地址
func (s *memberAddressSrv) DeleteAddress(ctx context.Context, req member_req.MemberAddressDeleteReq) error {
	memberAddressRepo := repository.NewMemberAddressRepo(ctx.GetDB())

	// 获取要删除的地址信息
	memberAddress, err := memberAddressRepo.GetMemberAddressByUuid(req.Uuid)
	if err != nil {
		return errors.New("地址不存在")
	}

	// 验证该地址是否属于当前会员
	if memberAddress.MemberUuid != ctx.GetMemberUuid() {
		return errors.New("地址不存在")
	}

	// 判断是否为默认地址
	isDefault := memberAddress.IsDefault

	// 开启事务
	err = ctx.GetDB().Transaction(func(tx *gorm.DB) error {
		// 先删除当前地址
		if err := repository.NewMemberAddressRepo(tx).Delete(req.Uuid); err != nil {
			return err
		}

		// 如果删除的是默认地址，则将该会员的第一个地址设为默认地址
		if isDefault == 1 {
			// 获取该会员的所有地址
			addresses, err := repository.NewMemberAddressRepo(tx).GetMemberAddressByMemberUuid(ctx.GetMemberUuid())
			if err != nil {
				return err
			}

			// 如果还有地址，则将第一个地址设为默认地址
			if len(addresses) > 0 {
				addresses[0].IsDefault = 1
				_, err = repository.NewMemberAddressRepo(tx).Update(*addresses[0])
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// AuthAddress 认证地址
func (s *memberAddressSrv) AuthAddress(ctx context.Context, req member_req.MemberAddressAuthReq) (member_resp.LoginResp, error) {
	if err := req.Validate(); err != nil {
		return member_resp.LoginResp{}, errors.New(err.Error())
	}

	db := ctx.GetDB()
	companyUuid := ctx.GetCompanyUuid()

	// 获取地址信息
	memberAddressRepo := repository.NewMemberAddressRepo(db)
	memberAddress, err := memberAddressRepo.GetMemberAddressByUuid(req.Uuid)
	if err != nil {
		return member_resp.LoginResp{}, errors.New("地址不存在")
	}

	// 如果需要注册，则注册会员
	var tokens member_resp.LoginResp
	if req.IsRegister {
		// 如果当前不是游客，则不能注册
		if member := ctx.GetMember(); !member.IsVisitor {
			return member_resp.LoginResp{}, errors.New("当前不是游客，不能注册")
		}
		// 注册会员
		tokens, err = NewMemberSrv(s.dbm, s.cache).Register(ctx, member_req.MemberRegisterReq{
			Phone:         memberAddress.Phone,
			Code:          req.Code,
			ReferrerPhone: req.ReferrerPhone,
		})
		if err != nil {
			return member_resp.LoginResp{}, err
		}
	} else {
		// 验证验证码
		if err := validator.VerifyCode(s.cache, companyUuid, memberAddress.Phone, req.Code); err != nil {
			return member_resp.LoginResp{}, err
		}
	}

	// 获取该会员的所有地址 - 找出未认证并与当前手机号相同的地址，并认证
	addresses, err := repository.NewMemberAddressRepo(db).GetMemberAddressByMemberUuid(ctx.GetMemberUuid())
	if err == nil && len(addresses) > 0 {
		for _, address := range addresses {
			if address.IsAuthPhone() {
				continue
			}
			if address.Phone == memberAddress.Phone {
				address.AuthPhone = address.Phone
				address.AuthTime = time.Now().Unix()
				_, err = repository.NewMemberAddressRepo(db).Update(*address)
				if err != nil {
					return member_resp.LoginResp{}, err
				}
			}
		}
	}

	return tokens, nil
}
