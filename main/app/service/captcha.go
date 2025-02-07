package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image/color"
	"time"

	"github.com/mojocn/base64Captcha"

	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
)

type ICaptchaSrv interface {
	Generate() (*resp.CaptchaResponse, error)
	Verify(sign string, answer string) bool
}

func NewCaptchaSrv(cache cache.Cache) ICaptchaSrv {
	return NewCaptchaSrvImpl(cache)
}

// CustomMathDriver 自定义数学验证码驱动
type CustomMathDriver struct {
	base64Captcha.DriverMath
}

// GenerateIdQuestionAnswer 重写生成问题和答案的方法
func (d *CustomMathDriver) GenerateIdQuestionAnswer() (id, question, answer string) {
	id = base64Captcha.RandomId()
	// 生成两个10以内的随机数
	b := make([]byte, 4)
	rand.Read(b)
	// 使用多个字节生成随机数，增加随机性
	num1 := (int(b[0]) + int(b[1])) % 10
	num2 := (int(b[2]) + int(b[3])) % 10
	if num1 == 0 {
		num1 = 1
	} // 避免出现0
	if num2 == 0 {
		num2 = 1
	} // 避免出现0
	isAdd := b[0]%2 == 0 // 随机决定是加法还是减法

	if isAdd {
		// 随机决定数字的位置
		if b[1]%2 == 0 {
			question = fmt.Sprintf("%d + %d =", num1, num2) // 添加空格
		} else {
			question = fmt.Sprintf("%d+%d=", num1, num2) // 不添加空格
		}
		answer = fmt.Sprintf("%d", num1+num2)
	} else {
		// 确保结果为正数
		if num1 < num2 {
			num1, num2 = num2, num1
		}
		// 随机决定数字的位置
		if b[1]%2 == 0 {
			question = fmt.Sprintf("%d - %d =", num1, num2) // 添加空格
		} else {
			question = fmt.Sprintf("%d-%d=", num1, num2) // 不添加空格
		}
		answer = fmt.Sprintf("%d", num1-num2)
	}
	return
}

type CaptchaSrv struct {
	cache  cache.Cache
	driver *CustomMathDriver
	store  base64Captcha.Store
}

func NewCaptchaSrvImpl(cache cache.Cache) *CaptchaSrv {
	// 配置验证码选项
	driver := &CustomMathDriver{
		DriverMath: base64Captcha.DriverMath{
			Height:          60,
			Width:           150,
			NoiseCount:      30,
			ShowLineOptions: 0,
			BgColor: &color.RGBA{
				R: 250,
				G: 251,
				B: 252,
				A: 255,
			},
			Fonts: []string{"wqy-microhei.ttc", "RitaSmith.ttf"},
		},
	}

	return &CaptchaSrv{
		cache:  cache,
		driver: driver,
		store:  base64Captcha.DefaultMemStore,
	}
}

// 生成随机标识
func generateSign() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *CaptchaSrv) Generate() (*resp.CaptchaResponse, error) {
	// 生成随机标识
	sign, err := generateSign()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sign: %v", err)
	}

	// 生成验证码
	c := base64Captcha.NewCaptcha(s.driver, s.store)
	id, b64s, err := c.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate captcha: %v", err)
	}

	// 从验证码图片中获取答案
	answer := s.store.Get(id, false)
	if answer == "" {
		return nil, fmt.Errorf("failed to get captcha answer")
	}

	// 将答案存储到缓存，设置5分钟过期
	err = s.cache.Set(config.Captcha.CachePrefix+sign, answer, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to store captcha: %v", err)
	}

	return &resp.CaptchaResponse{
		Sign:   sign,
		Base64: b64s,
	}, nil
}

func (s *CaptchaSrv) Verify(sign, answer string) bool {
	// 开发调试使用
	if config.Server.Mode == "debug" && answer == "123456" {
		return true
	}
	// 从缓存中获取正确答案
	correctAnswer, exists := s.cache.Get(config.Captcha.CachePrefix + sign)
	if !exists {
		return false
	}

	// 验证后立即删除答案和ID
	s.cache.Del(config.Captcha.CachePrefix + sign)
	return correctAnswer.(string) == answer
}
