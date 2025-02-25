package captcha

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang/freetype"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand"
	"strconv"
	"time"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/captcha/fonts"

	"github.com/golang/freetype/truetype"
)

// Captcha 结构体用于存储验证码相关的状态
type Captcha struct {
	font        *truetype.Font
	randGen     *rand.Rand
	cache       cache.Cache
	cachePrefix string
	maxAge      time.Duration
}

// New 创建一个新的验证码生成器
func New(cache cache.Cache, cachePrefix string, maxAge time.Duration) (*Captcha, error) {
	font, err := truetype.Parse(fonts.DefaultFontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	// 创建随机数生成器
	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)

	return &Captcha{
		font:        font,
		randGen:     randGen,
		cache:       cache,
		cachePrefix: cachePrefix,
		maxAge:      maxAge,
	}, nil
}

// Generate 生成新的验证码
func (cap *Captcha) Generate() (string, string, error) {
	// 生成随机ID
	id := cap.generateID()
	// 生成算术题
	question, answer := cap.generateMathQuestion()
	// 生成图片
	base64Img, err := cap.generateCaptchaBase64(question)
	if err != nil {
		return "", "", err
	}
	cacheKey := cap.cachePrefix + id
	cap.cache.Set(cacheKey, answer, cap.maxAge)
	return id, base64Img, nil
}

// Verify 验证用户输入
func (cap *Captcha) Verify(id, answer string) (bool, error) {
	cacheKey := cap.cachePrefix + id
	info, exists := cap.cache.Get(cacheKey)
	if !exists {
		return false, fmt.Errorf("captcha not found")
	}
	passAnswer, ok := info.(string)
	if !ok {
		return false, fmt.Errorf("captcha not found")
	}
	// 验证后删除
	defer cap.cache.Del(cacheKey)
	return answer == passAnswer, nil
}

// generateID 生成随机ID
func (cap *Captcha) generateID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[cap.randGen.Intn(len(charset))]
	}
	return string(b)
}

// generateMathQuestion 生成算术题
func (cap *Captcha) generateMathQuestion() (string, string) {
	var a, b int
	var operator string
	var result int

	if cap.randGen.Intn(2) == 1 {
		operator = "-"
		a = cap.randGen.Intn(10) + 1
		b = cap.randGen.Intn(a) + 1
		result = a - b
	} else {
		operator = "+"
		a = cap.randGen.Intn(10) + 1
		b = cap.randGen.Intn(10) + 1
		result = a + b
	}

	question := fmt.Sprintf("%d %s %d = ", a, operator, b)
	return question, strconv.Itoa(result)
}

// 生成验证码图片并返回 Base64 URL
func (cap *Captcha) generateCaptchaBase64(text string) (string, error) {
	width := 200 // 减小宽度
	height := 50 // 减小高度

	// 创建一个新的 RGBA 图片
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充白色背景（JPEG不支持透明）
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	// 创建背景字体上下文（较小字体）
	bgContext := freetype.NewContext()
	bgContext.SetDPI(72)
	bgContext.SetFont(cap.font)
	bgContext.SetFontSize(14) // 减小背景字体大小
	bgContext.SetClip(img.Bounds())
	bgContext.SetDst(img)

	// 添加随机背景数字（减少数量）
	for i := 0; i < 20; i++ {
		x := rand.Intn(width - 15)
		y := rand.Intn(height-8) + 15

		// 使用更深的背景色 (160-190)
		gray := uint8(rand.Intn(30) + 160)
		bgContext.SetSrc(image.NewUniform(color.RGBA{R: gray, G: gray, B: gray, A: 255}))

		num := strconv.Itoa(rand.Intn(10))
		pt := freetype.Pt(x, y)
		_, err := bgContext.DrawString(num, pt)
		if err != nil {
			return "", errors.New("failed to generate captcha")
		}
	}

	// 添加随机背景点和半点（减少数量）
	for i := 0; i < 300; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		// 保持背景点颜色较浅
		gray := uint8(rand.Intn(20) + 210)

		if rand.Intn(2) == 0 {
			img.Set(x, y, color.RGBA{R: gray, G: gray, B: gray, A: 255})
		} else {
			numPixels := rand.Intn(2) + 1 // 减少半点的像素数
			pixels := [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}}
			rand.Shuffle(len(pixels), func(i, j int) {
				pixels[i], pixels[j] = pixels[j], pixels[i]
			})
			for p := 0; p < numPixels; p++ {
				px, py := pixels[p][0], pixels[p][1]
				if x+px < width && y+py < height {
					img.Set(x+px, y+py, color.RGBA{R: gray, G: gray, B: gray, A: 255})
				}
			}
		}
	}

	// 创建主文字上下文
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(cap.font)
	c.SetFontSize(28)
	c.SetClip(img.Bounds())
	c.SetDst(img)

	// 分割文本
	chars := []rune(text)
	charWidth := width / (len(chars) + 1)

	// 为所有字符生成一个随机颜色
	var r, g, b uint8
	colorType := rand.Intn(3) // 随机选择一种颜色类型
	switch colorType {
	case 0: // 深蓝色系
		r = uint8(rand.Intn(20))
		g = uint8(rand.Intn(20))
		b = uint8(rand.Intn(100) + 156) // 156-255
	case 1: // 深红色系
		r = uint8(rand.Intn(100) + 156) // 156-255
		g = uint8(rand.Intn(20))
		b = uint8(rand.Intn(20))
	case 2: // 深绿色系
		r = uint8(rand.Intn(20))
		g = uint8(rand.Intn(100) + 156) // 156-255
		b = uint8(rand.Intn(20))
	}
	textColor := image.NewUniform(color.RGBA{R: r, G: g, B: b, A: 255})
	c.SetSrc(textColor)

	// 为每个字符生成随机位置并绘制
	for i, char := range chars {
		xPos := charWidth*(i+1) - charWidth/2 + rand.Intn(8) - 4
		yPos := height/2 + rand.Intn(8) - 4 + 15

		pt := freetype.Pt(xPos, yPos)
		_, err := c.DrawString(string(char), pt)
		if err != nil {
			return "", errors.New("failed to generate captcha")
		}
	}

	// 使用 JPEG 编码（使用较高质量以保持文字清晰度）
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{
		Quality: 60, // 60%的质量
	})
	if err != nil {
		return "", errors.New("failed to generate captcha")
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/jpeg;base64," + base64Str, nil
}
