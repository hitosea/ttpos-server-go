package service

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/pgp"
	"ttpos-server-go/pkg/utils"
)

type PGPService struct {
	cache cache.Cache
}

func NewPGPService(cache cache.Cache) *PGPService {
	return &PGPService{
		cache: cache,
	}
}

// GetServerPublicKey 获取服务端公钥
func (s *PGPService) GetServerPublicKey(clientId string) (*resp.ServerPublicKeyResponse, error) {
	cacheKey := config.Pgp.CachePrefix + clientId
	if data, ok := s.cache.Get(cacheKey); ok {
		var pgpPair pgp.KeyPair
		err := json.Unmarshal([]byte(data.(string)), &pgpPair)
		if err != nil {
			return nil, errors.New("获取服务端公钥失败")
		}
		return &resp.ServerPublicKeyResponse{
			Type:            "pgp",
			ClientId:        clientId,
			ServerPublicKey: pgpPair.PublicKey,
		}, nil
	}

	name := strings.ToLower(utils.RandomString(8, utils.LowerLetters, utils.UpperLetters, utils.Numbers))
	passphrase := strings.ToLower(utils.RandomString(8, utils.LowerLetters, utils.UpperLetters, utils.Numbers))
	pgpPair, err := pgp.GenerateKeyPair(name, "aa@bb.cc", passphrase)
	if err != nil {
		return nil, errors.New("获取服务端公钥失败")
	}

	b, _ := json.Marshal(pgpPair)
	s.cache.Set(cacheKey, string(b), 86400*90*time.Second)

	return &resp.ServerPublicKeyResponse{
		Type:            "pgp",
		ClientId:        clientId,
		ServerPublicKey: pgpPair.PublicKey,
	}, nil
}
