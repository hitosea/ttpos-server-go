package erpnext

import (
	"context"
	"fmt"
	"time"

	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/do"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"
	"ttpos-bmp/internal/pkg/cache"
	"ttpos-bmp/internal/pkg/crypto"

	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	TokenKeyPrefix = "erpnext_token:"
)

type SiteAuthorization struct {
	SiteUrl       string
	Authorization string
}

// GetAndProcessSiteApiSecret 获取并处理 api_secret
// 如果 api_secret 不是 base64 编码，则先使用 gaes.Encrypt 加密，再用 gbase64.Encode 编码后保存并返回
func (s *sRpc) GetAndProcessSiteAuthorization(ctx context.Context, siteCode string) (*SiteAuthorization, error) {
	cacheKey := fmt.Sprintf("%s%s", TokenKeyPrefix, siteCode)
	value, err := cache.Instance().GetOrSetFunc(ctx, cacheKey, func(ctx context.Context) (any, error) {
		var site *entity.Site
		err := dao.Site.Ctx(ctx).Where(dao.Site.Columns().SiteCode, siteCode).Limit(1).Scan(&site)
		if err != nil {
			return nil, gerror.Wrapf(err, "根据站点编码[%s]查询站点信息失败", siteCode)
		}
		if site == nil {
			return nil, gerror.Newf("站点编码[%s]不存在", siteCode)
		}

		// 如果 api_secret 为空，直接返回
		if site.ApiSecret == "" {
			return nil, gerror.New("api_secret 为空")
		}

		// 判断是否为 base64 编码
		if isBase64Encoded(site.ApiSecret) {
			apiSecret, err := crypto.Decrypt(ctx, site.ApiSecret)
			if err != nil {
				return nil, gerror.Wrapf(err, "解密 api_secret 失败: %s", site.SiteCode)
			}
			// 已经是 base64 编码，直接返回
			return &SiteAuthorization{
				SiteUrl:       site.SiteUrl,
				Authorization: fmt.Sprintf("token %s:%s", site.ApiKey, apiSecret),
			}, nil
		}

		// 不是 base64 编码，需要加密并编码
		// 使用加密包进行加密和编码
		encodedSecret, err := crypto.Encrypt(ctx, site.ApiSecret)
		if err != nil {
			return nil, gerror.Wrapf(err, "加密 api_secret 失败")
		}

		// 保存到数据库
		_, err = dao.Site.Ctx(ctx).Data(do.Site{ApiSecret: encodedSecret}).Where(dao.Site.Columns().SiteCode, siteCode).Update()
		if err != nil {
			return nil, gerror.Wrapf(err, "保存加密后的 api_secret 失败")
		}
		return &SiteAuthorization{
			SiteUrl:       site.SiteUrl,
			Authorization: fmt.Sprintf("token %s:%s", site.ApiKey, site.ApiSecret),
		}, nil
	}, 30*time.Second)

	if err != nil {
		return nil, gerror.Wrapf(err, "获取 api_secret 失败")
	}

	if value == nil || value.IsNil() {
		return nil, gerror.New("获取到的 api_secret 为空")
	}
	return value.Val().(*SiteAuthorization), nil
}

// GetAndProcessCashierApiSecret 获取并处理收银员的 api_secret
// 如果 api_secret 不是 base64 编码，则先使用 gaes.Encrypt 加密，再用 gbase64.Encode 编码后保存并返回
func (s *sRpc) GetAndProcessCashierAuthorization(ctx context.Context, cashierEmail string) (string, error) {
	cacheKey := fmt.Sprintf("%s%s", TokenKeyPrefix, cashierEmail)
	value, err := cache.Instance().GetOrSetFunc(ctx, cacheKey, func(ctx context.Context) (any, error) {
		var cashier *entity.ShopCashier
		err := dao.ShopCashier.Ctx(ctx).Where(dao.ShopCashier.Columns().CashierEmail, cashierEmail).Limit(1).Scan(&cashier)
		if err != nil {
			return nil, gerror.Wrapf(err, "根据收银员邮箱[%s]查询收银员信息失败", cashierEmail)
		}
		if cashier == nil {
			return nil, gerror.Newf("收银员邮箱[%s]对应的收银员不存在", cashierEmail)
		}

		// 如果 api_secret 为空，直接返回
		if cashier.ApiSecret == "" {
			return nil, gerror.New("api_secret 为空")
		}

		// 判断是否为 base64 编码
		if isBase64Encoded(cashier.ApiSecret) {
			apiSecret, err := crypto.Decrypt(ctx, cashier.ApiSecret)
			if err != nil {
				return nil, gerror.Wrapf(err, "解密收银员 api_secret 失败: %v", cashier)
			}
			// 已经是 base64 编码，直接返回
			return fmt.Sprintf("token %s:%s", cashier.ApiKey, apiSecret), nil
		}

		// 不是 base64 编码，需要加密并编码
		// 使用加密包进行加密和编码
		encodedSecret, err := crypto.Encrypt(ctx, cashier.ApiSecret)
		if err != nil {
			return nil, gerror.Wrapf(err, "加密 api_secret 失败")
		}

		// 保存到数据库
		_, err = dao.ShopCashier.Ctx(ctx).Data(do.ShopCashier{ApiSecret: encodedSecret}).Where(dao.ShopCashier.Columns().CashierEmail, cashierEmail).Update()
		if err != nil {
			return nil, gerror.Wrapf(err, "保存加密后的 api_secret 失败")
		}
		return fmt.Sprintf("token %s:%s", cashier.ApiKey, cashier.ApiSecret), nil

	}, 30*time.Second)

	if err != nil {
		return "", gerror.Wrapf(err, "获取 api_secret 失败")
	}

	if value == nil || value.IsNil() {
		return "", gerror.New("获取到的 api_secret 为空")
	}

	return value.String(), nil
}

// isBase64Encoded 判断字符串是否为 base64 编码
func isBase64Encoded(s string) bool {
	if s == "" {
		return false
	}
	// 尝试解码，如果成功则认为是 base64 编码
	_, err := gbase64.DecodeString(s)
	return err == nil
}
