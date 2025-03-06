package service

import (
	"encoding/json"
	"time"

	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/encrypt"
)

type IEncryptSrv interface {
	GetServerPublicKey(clientId string, encryptType string) (*resp.ServerKey, error)
}

func NewEncryptSrv(cache cache.Cache) IEncryptSrv {
	return NewEncryptSrvImpl(cache)
}

type encryptSrv struct {
	cache cache.Cache
}

func NewEncryptSrvImpl(cache cache.Cache) IEncryptSrv {
	return &encryptSrv{
		cache: cache,
	}
}

// GetServerPublicKey 获取服务端公钥
func (s *encryptSrv) GetServerPublicKey(clientId string, encryptType string) (*resp.ServerKey, error) {
	cacheKey := config.Encrypt.CachePrefix + clientId + "_" + encryptType
	if data, ok := s.cache.Get(cacheKey); ok {
		var keyPair encrypt.KeyPair
		err := json.Unmarshal([]byte(data.(string)), &keyPair)
		if err != nil {
			return nil, errors.WithMessage(err, "获取服务端公钥失败")
		}
		return &resp.ServerKey{
			Type:      encryptType,
			ClientId:  clientId,
			ServerKey: keyPair.PublicKey,
		}, nil
	}

	var keyPair encrypt.KeyPair
	switch encryptType {
	case "jsencrypt":
		var err error
		if keyPair, err = encrypt.GenerateRSAKeyPairPEM(2048); err != nil {
			return nil, errors.WithMessage(err, "获取服务端公钥失败")
		}
	default:
		return nil, errors.New("获取服务端公钥失败")
	}

	b, _ := json.Marshal(keyPair)
	s.cache.Set(cacheKey, string(b), 86400*90*time.Second)

	return &resp.ServerKey{
		Type:      encryptType,
		ClientId:  clientId,
		ServerKey: keyPair.PublicKey,
	}, nil
}
