package service

import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/copier"
	"strings"
	"time"

	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/encrypt"
	"ttpos-server-go/pkg/utils"
)

type IEncryptSrv interface {
	GetServerPublicKey(clientId string, encryptType string) (*resp.ServerKeyResponse, error)
}

func NewEncryptSrv(cache cache.Cache) IEncryptSrv {
	return NewEncryptSrvImpl(cache)
}

type EncryptSrv struct {
	cache cache.Cache
}

func NewEncryptSrvImpl(cache cache.Cache) *EncryptSrv {
	return &EncryptSrv{
		cache: cache,
	}
}

// GetServerPublicKey 获取服务端公钥
func (s *EncryptSrv) GetServerPublicKey(clientId string, encryptType string) (*resp.ServerKeyResponse, error) {
	cacheKey := config.Encrypt.CachePrefix + clientId + "_" + encryptType
	if data, ok := s.cache.Get(cacheKey); ok {
		var keyPair encrypt.KeyPair
		err := json.Unmarshal([]byte(data.(string)), &keyPair)
		if err != nil {
			return nil, errors.New("获取服务端公钥失败")
		}
		return &resp.ServerKeyResponse{
			Type:      encryptType,
			ClientId:  clientId,
			ServerKey: keyPair.PublicKey,
		}, nil
	}

	var keyPair encrypt.KeyPair
	switch encryptType {
	case "pgp":
		name := strings.ToLower(utils.RandomString(8, utils.LowerLetters, utils.UpperLetters, utils.Numbers))
		passphrase := strings.ToLower(utils.RandomString(8, utils.LowerLetters, utils.UpperLetters, utils.Numbers))
		kp, err := encrypt.GeneratePgpKeyPair(name, "aa@bb.cc", passphrase)
		if err != nil {
			return nil, errors.New("获取服务端公钥失败")
		}
		_ = copier.Copy(&keyPair, *kp)
	case "jsencrypt":
		var err error
		if keyPair, err = encrypt.GenerateRSAKeyPairPEM(2048); err != nil {
			return nil, errors.New("获取服务端公钥失败")
		}
	default:
		return nil, errors.New("获取服务端公钥失败")
	}

	b, _ := json.Marshal(keyPair)
	s.cache.Set(cacheKey, string(b), 86400*90*time.Second)

	return &resp.ServerKeyResponse{
		Type:      encryptType,
		ClientId:  clientId,
		ServerKey: keyPair.PublicKey,
	}, nil
}
