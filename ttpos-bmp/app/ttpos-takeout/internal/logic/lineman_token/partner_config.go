// Package lineman_token 提供 LINE MAN OAuth Token 生成与验证服务
//
// ⚠️ 临时方案: 后续将迁移到统一权限中心 SSO
package lineman_token

import (
	"context"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
)

// PartnerConfigLoader 负责加载并缓存 lineman.partner.* 配置
type PartnerConfigLoader struct {
	once        sync.Once
	partners    map[string]*conf.LinemanPartner // code -> partner
	clientIndex map[string]*conf.LinemanPartner // clientId -> partner
	loadErr     error
}

// load 加载配置并构建索引
func (l *PartnerConfigLoader) load(ctx context.Context) {
	l.once.Do(func() {
		ctx = gctx.WithCtx(ctx)

		var partnerMap conf.LinemanPartnerMap
		if err := g.Cfg().MustGet(ctx, "app.provider.lineman.partner").Scan(&partnerMap); err != nil {
			l.loadErr = gerror.Wrap(err, "加载 app.provider.lineman.partner 配置失败")
			return
		}

		l.partners = make(map[string]*conf.LinemanPartner)
		l.clientIndex = make(map[string]*conf.LinemanPartner)

		for code, p := range partnerMap {
			if p == nil {
				continue
			}
			normalizedCode := strings.TrimSpace(code)
			l.partners[normalizedCode] = p
			if p.ClientID != "" {
				l.clientIndex[p.ClientID] = p
			}
		}
	})
}

// GetByCode 通过 partner code 获取配置
func (l *PartnerConfigLoader) GetByCode(ctx context.Context, code string) (*conf.LinemanPartner, error) {
	l.load(ctx)
	if l.loadErr != nil {
		return nil, l.loadErr
	}
	partner, ok := l.partners[strings.TrimSpace(code)]
	if !ok {
		return nil, gerror.Newf("未找到 LINE MAN Partner 配置: %s", code)
	}
	return partner, nil
}

// GetByClientID 通过 client_id 获取配置
func (l *PartnerConfigLoader) GetByClientID(ctx context.Context, clientID string) (*conf.LinemanPartner, error) {
	l.load(ctx)
	if l.loadErr != nil {
		return nil, l.loadErr
	}
	partner, ok := l.clientIndex[strings.TrimSpace(clientID)]
	if !ok {
		return nil, gerror.Newf("未找到 LINE MAN Partner 配置，client_id=%s", clientID)
	}
	return partner, nil
}
