// 字体生成图像
// @wfs: 2024/04/19
package pkg

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"ttpos-server-go/app/printer/pkg/fonts"
	"ttpos-server-go/app/printer/pkg/rabbit"

	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
)

// ImgFont 类用于生成图像并添加文本
type ImgFont struct {
	// 图像对象
	Image *image.RGBA
	// 方向
	Direction int
	// 图片宽度
	ImageWidth int
	// 图片高度
	ImageHeight int
	// 图片内边距
	ImagePadding int
	// 文本行高
	DefaultTextLineHeight int
	TextLineHeight        int
	// 文本间距
	TextSpacing float64
	// 文本总高度
	TextTotalHeight int
	// 文本最后一行已使用的宽度
	TextLastLineUsedWidth int
	// 文本对齐方向
	Alignment int
	// 字体大小
	FontSize int
	// 字体粗细
	FontWeight int
	// 缅甸语的特殊字体
	MySpecialFonts map[string]string
	// 字体缓存
	FontCache map[string]*truetype.Font
}

// 字体常量
const (
	// 横向排列
	DIRECTION_X = 0

	// 竖向排列
	DIRECTION_Y = 1

	// 文本对齐方向
	ALIGN_LEFT   = 1
	ALIGN_CENTER = 2
	ALIGN_RIGHT  = 3
)

// NewImgFont 创建一个新的ImgFont实例
func NewImgFont(imageWidth int, defaultTextLineHeight int, direction int) *ImgFont {
	img := &ImgFont{
		ImageWidth:            567, // 默认宽度
		ImageHeight:           0,
		ImagePadding:          24,
		DefaultTextLineHeight: 45,
		TextLineHeight:        45,
		TextSpacing:           1.0,
		TextTotalHeight:       0,
		TextLastLineUsedWidth: 0,
		Alignment:             ALIGN_LEFT,
		FontSize:              20,
		FontWeight:            1,
		Direction:             0,
		MySpecialFonts:        make(map[string]string),
		FontCache:             make(map[string]*truetype.Font),
	}

	// 设置自定义宽度
	if imageWidth > 0 {
		img.ImageWidth = imageWidth
	}

	// 设置自定义行高
	if defaultTextLineHeight > 0 {
		img.DefaultTextLineHeight = defaultTextLineHeight
		img.TextLineHeight = defaultTextLineHeight
	}

	// 设置方向
	if direction == DIRECTION_Y {
		img.Direction = 90
		img.ImageHeight = imageWidth
	}

	// 初始化缅甸语特殊字体映射
	// 注意：这里需要实现与Rabbit.uni2zg相同的功能
	// 由于Go缺少类似的库，可以用简单的映射代替或引入第三方库
	img.MySpecialFonts = map[string]string{
		"ကြို": rabbit.Uni2Zg("ၾကိဳ"),
		"က္တြ": rabbit.Uni2Zg("ကြၽ"),
		"ကြွ":  rabbit.Uni2Zg("ကြြ"),
		"ပြု":  rabbit.Uni2Zg("ျပဳ"),
		"ကြော": rabbit.Uni2Zg("ေၾကာ"),
		"မှု":  rabbit.Uni2Zg("မႈ"),
		"က္ခ":  rabbit.Uni2Zg("ကၡ"),
		"ဏ္ဍ​": rabbit.Uni2Zg("႑"),
		"ဒ္ဒ​": rabbit.Uni2Zg("ဒၵ"),
		"န္တ":  rabbit.Uni2Zg("နၱ"),
		"န္အ":  rabbit.Uni2Zg("နၲ"),
	}

	// 创建图像
	img.createImg()

	return img
}

// createImg 创建一个新的图像
func (i *ImgFont) createImg() {
	// 创建一个空白画布
	i.Image = image.NewRGBA(image.Rect(0, 0, i.ImageWidth, 20000))

	// 设置背景颜色为白色
	draw.Draw(i.Image, i.Image.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
}

// GetFontPath 根据字符获取对应的字体路径
func (i *ImgFont) GetFontPath(char string) string {
	// 泰语
	thaiPattern := regexp.MustCompile(`[\p{Thai}]`)
	if thaiPattern.MatchString(char) || strings.Contains(char, "฿") {
		return fonts.FontTH
	}

	// 韩语
	hangulPattern := regexp.MustCompile(`[\p{Hangul}]`)
	if hangulPattern.MatchString(char) {
		return fonts.FontKO
	}

	// 中文
	chiPattern := regexp.MustCompile(`[\x{4E00}-\x{9FFF}\x{FF01}\x{FF0C}\x{FF08}\x{FF09}\x{FF1A}\x{FF5E}\x{2014}]`)
	if chiPattern.MatchString(char) || char == "、" || char == "-" || char == "￥" || char == "；" || char == "？" || char == "＋" || char == "：" {
		return fonts.FontZH
	}

	// 日文
	japPattern := regexp.MustCompile(`[\p{Hiragana}\p{Katakana}\p{Han}★]+`)
	if japPattern.MatchString(char) {
		return fonts.FontJA
	}

	// 缅甸语
	myPattern := regexp.MustCompile(`[\x{1000}-\x{109F}\x{AA60}-\x{AA7F}\x{A9E0}-\x{A9FF}\x{AA20}-\x{AA3F}]`)
	if myPattern.MatchString(char) {
		// 检查是否为特殊字体
		for _, specialChar := range i.MySpecialFonts {
			if specialChar == char {
				return fonts.FontMY2
			}
		}
		return fonts.FontMY
	}

	// 土耳其语
	trPattern := regexp.MustCompile(`[\x{011E}-\x{0130}\x{0131}\x{015E}-\x{015F}\x{00C7}-\x{00E7}\x{011F}]`)
	if trPattern.MatchString(char) {
		return fonts.FontTR
	}

	// 默认英文
	return fonts.FontEN
}

/**
* 获取字体宽度
* @access private
* @param string $text
* @return string
 */
// GetFontWeight 获取字体宽度
func (i *ImgFont) GetFontWeight(fontSize int, char string) int {
	fontPath := i.GetFontPath(char)

	// 特殊字符'ြ'的处理
	if char == "ြ" {
		return 5
	}

	// 获取字体宽度
	// 使用getCharWidth计算字符宽度
	contextWidth := i.getCharWidth(fontPath, fontSize, char)

	// 缅甸语特殊处理
	if fontPath == fonts.FontMY || fontPath == fonts.FontMY2 {
		if contextWidth < 0 {
			contextWidth = 0
		}
	}

	return contextWidth
}

// getCharWidth 使用freetype获取字符宽度 - 模拟PHP的imagettfbbox的行为
func (i *ImgFont) getCharWidth(fontPath string, fontSize int, text string) int {
	// 如果文本为空，返回默认值
	if len(text) == 0 {
		return fontSize
	}

	// 尝试从字体缓存获取
	var f *truetype.Font
	var ok bool

	if f, ok = i.FontCache[fontPath]; !ok {
		// 加载字体文件
		var fontBytes []byte
		var err error

		// 使用嵌入字体模块
		fontBytes, err = fonts.GetFontData(fontPath)
		if err != nil {
			fmt.Println("从嵌入字体加载失败:", err)
			return fontSize
		}

		// 解析字体
		f, err = freetype.ParseFont(fontBytes)
		if err != nil {
			fmt.Println("解析字体失败:", err)
			return fontSize
		}

		// 添加到缓存
		i.FontCache[fontPath] = f
	}

	// 创建字体选项
	opt := truetype.Options{
		Size:    float64(fontSize),
		DPI:     100,
		Hinting: font.HintingNone, // 完全禁用hinting模式以避免栈溢出错误
	}

	// 获取字体表面
	face := truetype.NewFace(f, &opt)

	// 计算整个文本的宽度
	totalWidth := 0
	for _, r := range text {
		advance, ok := face.GlyphAdvance(r)
		if !ok {
			// 如果无法获取字形宽度，使用默认值
			totalWidth += fontSize
		} else {
			// fixed.Int26_6 转换为整数
			totalWidth += advance.Round()
		}
	}

	// 应用缩放系数
	scaleFactor := 0.99

	return int(float64(totalWidth) * scaleFactor)
}

// GetFontArrays 获取字体数组
func (i *ImgFont) GetFontArrays(texts string) []string {
	// 初始化结果数组
	segments := []string{}

	// 创建特殊字体的键作为正则表达式
	specialFontKeys := []string{}
	for key := range i.MySpecialFonts {
		specialFontKeys = append(specialFontKeys, regexp.QuoteMeta(key))
	}

	// 如果没有特殊字体，则直接将文本作为单个字符返回
	if len(specialFontKeys) == 0 {
		// 分割文本为单个字符
		for _, r := range texts {
			segments = append(segments, string(r))
		}
		return segments
	}

	// 按特殊字体的键分割文本
	pattern := "(" + strings.Join(specialFontKeys, "|") + ")"
	re := regexp.MustCompile(pattern)
	textParts := re.Split(texts, -1)

	// 查找所有匹配的分隔符
	allMatches := re.FindAllStringIndex(texts, -1)

	// 合并文本块和分隔符
	for idx, part := range textParts {
		// 处理非特殊字体部分
		if part != "" {
			// 我们将每个字符取出来单独处理
			chars := []string{}
			for _, r := range part {
				chars = append(chars, string(r))
			}

			// 调整缅甸语的书写顺序
			for j := 1; j < len(chars); j++ {
				if chars[j] == "ြ" {
					// 交换字符顺序
					temp := ""
					if j > 0 {
						temp = chars[j-1]
					}
					chars[j-1] = chars[j]
					chars[j] = temp
				}
			}

			// 调整缅甸语的书写顺序
			for j := 1; j < len(chars); j++ {
				if chars[j] == "ေ" { // ေ 字符
					// 处理如果前两个字符是 ြ
					if j >= 2 {
						temp := chars[j-2]
						if temp == "ြ" { // ြ 字符
							chars[j-2] = chars[j]
							chars[j] = temp
						}
					}

					// 交换前一个字符
					temp := chars[j-1]
					chars[j-1] = chars[j]
					chars[j] = temp
				}
			}

			// 将字符添加到结果中
			segments = append(segments, chars...)
		}

		// 如果还有分隔符需要处理
		if idx < len(allMatches) {
			// 获取分隔符
			delimiter := texts[allMatches[idx][0]:allMatches[idx][1]]

			// 如果是特殊字体的键，则使用其值
			if val, ok := i.MySpecialFonts[delimiter]; ok {
				segments = append(segments, val)
			} else {
				segments = append(segments, delimiter)
			}
		}
	}

	return segments
}

// AddText 添加文本并返回当前高度和宽度
func (i *ImgFont) AddText(text string, height float64, fixedWidth, deviationWidth float64) map[string]float64 {
	// 字体大小和粗细
	fontSize := i.FontSize
	fontWeight := i.FontWeight

	// 设置文本颜色为黑色
	black := color.RGBA{0, 0, 0, 255}

	// 分隔字体为数组
	segments := i.GetFontArrays(text)

	// 获取已使用高度
	useWidth := 0.0

	// 1.居右并固定宽度的计算 ，2 正常
	if fixedWidth > 0 && i.Alignment == AlignRight {
		useWidth = float64(i.ImageWidth) - fixedWidth - deviationWidth - float64(i.ImagePadding)
		canWidth := float64(i.ImageWidth) - useWidth - float64(i.ImagePadding)*3

		tmpw := 0.0
		tmpStr := ""

		for _, char := range segments {
			tmpStr += char
			tmpw = float64(i.GetFontWeight(fontSize, tmpStr)) * i.TextSpacing
			if tmpw > canWidth {
				break
			}
		}

		useWidth = float64(i.ImageWidth) - fixedWidth - deviationWidth - float64(i.ImagePadding) - (tmpw + float64(i.ImagePadding))
	} else {
		useWidth = float64(i.TextLastLineUsedWidth) + deviationWidth
	}

	// 回归宽度
	i.TextLastLineUsedWidth = 0

	// 偏移标记
	isDeviation := deviationWidth

	// 执行绘制
	for key, char := range segments {
		// 记录当前宽度
		oldUseWidth := useWidth

		// 处理换行
		if char == "\n" {
			useWidth = deviationWidth

			// 剧右并固定宽度的计算
			if fixedWidth > 0 && i.Alignment == AlignRight {
				useWidth = float64(i.ImageWidth) - fixedWidth - float64(i.ImagePadding)
			}

			// 增加行高
			height += float64(i.TextLineHeight)
			continue
		}

		// 获取当前字符的宽度
		charWidth := float64(i.GetFontWeight(fontSize, char)) * i.TextSpacing

		// 非居左对齐的处理
		if i.Alignment != AlignLeft {
			// 获取当前到最后的所有字符
			lastText := segments[key:]

			// 计算最后一行的宽度
			lastTextWidth := 0.0
			for _, c := range lastText {
				lastTextWidth += float64(i.GetFontWeight(fontSize, c))
			}
			lastTextWidth = lastTextWidth * i.TextSpacing

			// 文本居中处理
			if (useWidth == 0 || isDeviation != 0) && i.Alignment == AlignCenter {
				oldUseWidth = (float64(i.ImageWidth) + deviationWidth - lastTextWidth - float64(i.ImagePadding)*2) / 2
				useWidth = oldUseWidth
				isDeviation = 0
			}

			// 文本居右处理
			if i.Alignment == AlignRight {
				calculateWidth := float64(i.ImageWidth) - useWidth - lastTextWidth
				if calculateWidth > float64(i.ImagePadding)*2 {
					oldUseWidth = (float64(i.ImageWidth) - lastTextWidth - float64(i.ImagePadding)*2) - 2
				}
			}

			// 小于则从0开始
			if useWidth <= 0 {
				useWidth = 0
			}
			if oldUseWidth <= 0 {
				oldUseWidth = 0
			}
		}

		// 累加宽度
		useWidth += charWidth

		// 判断是否需要换行
		isLinebreak := false

		// 正常计算，超过图像宽度需要换行
		if useWidth >= (float64(i.ImageWidth) - float64(i.ImagePadding)*2) {
			isLinebreak = true
		}

		// 居左并固定宽度的计算
		if fixedWidth > 0 && i.Alignment == AlignLeft {
			if useWidth-deviationWidth >= fixedWidth+float64(i.ImagePadding)*2 {
				isLinebreak = true
			}
		}

		// 如果不需要换行，绘制字符
		if !isLinebreak {
			fontPath := i.GetFontPath(char)
			// 使用freetype绘制文本
			i.drawText(
				char,
				fontPath,
				fontSize,
				fontWeight,
				int(oldUseWidth+float64(i.ImagePadding)),
				int(height),
				black,
			)
		} else {
			// 如果需要换行，处理剩余文本
			newArray := segments[key:]

			// 引用特殊字体的反向映射
			mySpecialFontsReversed := make(map[string]string)
			for k, v := range i.MySpecialFonts {
				mySpecialFontsReversed[v] = k
			}

			// 替换特殊字符
			for idx, c := range newArray {
				if val, ok := mySpecialFontsReversed[c]; ok {
					newArray[idx] = val
				}
			}

			// 递归处理剩余文本
			return i.AddText(
				strings.Join(newArray, ""),
				height+float64(i.TextLineHeight),
				fixedWidth,
				deviationWidth,
			)
		}
	}

	// 返回当前高度和宽度
	return map[string]float64{
		"height": height,
		"width":  useWidth,
	}
}

// drawText 使用freetype绘制文本
func (i *ImgFont) drawText(text, fontPath string, fontSize, fontWeight, x, y int, textColor color.RGBA) {
	var f *truetype.Font
	var err error

	// 使用嵌入的字体文件
	fontBytes, err := fonts.GetFontData(fontPath)
	if err != nil {
		fmt.Println("获取嵌入字体文件失败:", err)
		return
	}

	// 解析字体
	f, err = freetype.ParseFont(fontBytes)
	if err != nil {
		fmt.Println("解析字体失败:", err)
		return
	}

	// 先绘制描边（黑色或深色）
	outlineColor := color.RGBA{0, 0, 0, 255} // 黑色描边

	// 创建上下文
	ctx := freetype.NewContext()
	ctx.SetDPI(100)
	ctx.SetFont(f)
	ctx.SetClip(i.Image.Bounds())
	ctx.SetDst(i.Image)
	ctx.SetFontSize(float64(fontSize))
	ctx.SetSrc(image.NewUniform(outlineColor))
	ctx.SetHinting(font.HintingNone)

	// 后绘制原始文本
	ctx.SetSrc(image.NewUniform(textColor))
	pt := freetype.Pt(x, y+fontSize)
	_, err = ctx.DrawString(text, pt)
	if err != nil {
		fmt.Println("绘制原始文本失败:", err)
	}

}

// AppendPartingline 添加文本行
func (i *ImgFont) AppendPartingline(text string, fixedWidth int, deviationWidth float64) *ImgFont {
	// 确定起始高度，如果总高度为0则使用行高
	initHeight := float64(i.TextLineHeight)
	if i.TextTotalHeight > 0 {
		initHeight = float64(i.TextTotalHeight)
	}

	// 添加文本
	result := i.AddText(text, initHeight, float64(fixedWidth), deviationWidth)

	// 更新总高度
	i.TextTotalHeight = int(result["height"])

	// 更新行使用宽度
	i.TextLastLineUsedWidth = int(result["width"]) + 1

	return i
}

// AppendTextOption 定义 AppendText 的可选参数函数类型
type AppendTextOption func(*appendTextOptions)

// appendTextOptions 包含 AppendText 的所有可选参数
type appendTextOptions struct {
	fixedWidth     int
	deviationWidth float64
}

// DefaultAppendTextOptions 返回默认的 AppendText 选项
func DefaultAppendTextOptions() appendTextOptions {
	return appendTextOptions{
		fixedWidth:     0,
		deviationWidth: 0,
	}
}

// WithFixedWidth 设置固定宽度选项
func WithFixedWidth(width int) AppendTextOption {
	return func(o *appendTextOptions) {
		o.fixedWidth = width
	}
}

// WithDeviationWidth 设置偏差宽度选项
func WithDeviationWidth(deviation float64) AppendTextOption {
	return func(o *appendTextOptions) {
		o.deviationWidth = deviation
	}
}

// AppendText 添加多行文本（使用可选参数）
func (i *ImgFont) AppendText(text string, opts ...AppendTextOption) *ImgFont {
	// 应用默认选项
	options := DefaultAppendTextOptions()

	// 应用自定义选项
	for _, opt := range opts {
		opt(&options)
	}

	// 按换行符拆分文本
	texts := strings.Split(text, "\n")

	// 逐行处理
	for key, value := range texts {
		if value != "" {
			// 确定起始高度
			initHeight := float64(20)
			if i.TextTotalHeight > 0 {
				initHeight = float64(i.TextTotalHeight)
			}

			// 添加文本
			result := i.AddText(value, initHeight, float64(options.fixedWidth), options.deviationWidth)

			// 更新总高度
			i.TextTotalHeight = int(result["height"])

			// 更新行使用宽度
			i.TextLastLineUsedWidth = int(result["width"]) + 1
		}

		// 如果不是最后一行，添加换行
		if key != len(texts)-1 {
			i.LineFeed(1)
		}
	}

	return i
}

// AppendImg 添加图片
func (i *ImgFont) AppendImg(imgPath string, size int, isRoundness bool, topHeight int) *ImgFont {
	// 读取图片文件
	imgData, err := ioutil.ReadFile(imgPath)
	if err != nil {
		return i
	}

	// 解码图片
	srcImg, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return i
	}

	// 获取原始图片尺寸
	width := srcImg.Bounds().Dx()
	height := srcImg.Bounds().Dy()

	// 调整图片大小
	if size > 0 {
		ratio := float64(width) / float64(height)
		width = size
		height = int(float64(size) / ratio)

		// 创建新的调整大小后的图片
		rectangle := image.Rect(0, 0, width, height)
		resizedImg := image.NewRGBA(rectangle)

		// 使用resize包来调整图片大小
		resized := resize.Resize(uint(width), uint(height), srcImg, resize.Lanczos3)

		// 转换为RGBA图像
		if rgba, ok := resized.(*image.RGBA); ok {
			resizedImg = rgba
		} else {
			// 手动复制像素
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					resizedImg.Set(x, y, resized.At(x, y))
				}
			}
		}

		srcImg = resizedImg
	}

	// 处理圆角
	if isRoundness {
		// 创建圆形模板
		mask := image.NewRGBA(image.Rect(0, 0, width, height))

		// 填充透明背景
		draw.Draw(mask, mask.Bounds(), image.Transparent, image.Point{}, draw.Src)

		// 使用简单方法实现圆形裁剪
		r := float64(width) / 2
		centerX := float64(width) / 2
		centerY := float64(height) / 2

		// 遍历每个像素，将圆形区域内的像素设置为不透明
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				// 计算点到圆心的距离
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				distance := math.Sqrt(dx*dx + dy*dy)

				// 如果点在圆内，设置为不透明
				if distance <= r {
					mask.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
				}
			}
		}

		// 将裁剪后的图像绘制到新图像上
		roundedImg := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.DrawMask(roundedImg, roundedImg.Bounds(), srcImg, image.Point{}, mask, image.Point{}, draw.Over)
		srcImg = roundedImg
	}

	// 计算图片位置 (水平居中)
	x := (i.ImageWidth - width) / 2

	// 将图片绘制到目标图像上
	draw.Draw(
		i.Image,
		image.Rect(x, i.TextTotalHeight+topHeight, x+width, i.TextTotalHeight+topHeight+height),
		srcImg,
		image.Point{0, 0},
		draw.Over,
	)

	// 更新文本总高度和最后一行已使用宽度
	i.TextTotalHeight += height + topHeight
	i.TextLastLineUsedWidth += width

	// 添加换行
	i.LineFeed(1)

	return i
}

// AppendQrcode 添加二维码
func (i *ImgFont) AppendQrcode(data string, size int, margin int, isBase64 bool) *ImgFont {
	var qrImg image.Image
	var err error

	// 处理二维码数据
	if isBase64 {
		// 如果是Base64编码的图片数据
		decoded, decodeErr := base64.StdEncoding.DecodeString(data)
		if decodeErr != nil {
			return i
		}

		// 解码图片
		qrImg, _, err = image.Decode(bytes.NewReader(decoded))
		if err != nil {
			return i
		}

		// 调整总高度
		i.TextTotalHeight = i.TextTotalHeight - (margin * 2)
	} else {
		// 由于没有qrcode库，我们使用更简单的方法
		// 在这里创建一个空白图片作为占位符
		qrImg = image.NewRGBA(image.Rect(0, 0, size, size))
		// 在实际项目中应添加go-qrcode依赖并取消注释下面的代码
		/*
			qr, err := qrcode.New(data, qrcode.Medium)
			if err != nil {
				return i
			}
			qr.DisableBorder = false
			qrImg = qr.Image(size)
		*/
	}

	// 获取二维码尺寸
	width := qrImg.Bounds().Dx()
	height := qrImg.Bounds().Dy()

	// 调整图片大小
	if size > 0 && (width != size || height != size) {
		// 使用resize包来调整图片大小
		resized := resize.Resize(uint(size), uint(height*size/width), qrImg, resize.Lanczos3)

		// 转换为image.Image接口
		qrImg = resized

		// 更新尺寸
		width = size
		height = height * size / width
	}

	// 计算二维码位置 (水平居中)
	x := (i.ImageWidth - width) / 2

	// 将二维码绘制到目标图像上
	draw.Draw(
		i.Image,
		image.Rect(x, i.TextTotalHeight, x+width, i.TextTotalHeight+height),
		qrImg,
		image.Point{0, 0},
		draw.Over,
	)

	// 更新文本总高度和最后一行已使用宽度
	i.TextTotalHeight += height
	i.TextLastLineUsedWidth += width

	// 添加换行
	i.LineFeed(1)

	return i
}

// AppendSplitlineOption 定义 AppendSplitline 的可选参数函数类型
type AppendSplitlineOption func(*appendSplitlineOptions)

// appendSplitlineOptions 包含 AppendSplitline 的所有可选参数
type appendSplitlineOptions struct {
	isLineFeed bool
	lineHeight int
	fontWeight int
}

// DefaultAppendSplitLineOptions 返回默认的 AppendSplitline 选项
func DefaultAppendSplitLineOptions() appendSplitlineOptions {
	return appendSplitlineOptions{
		isLineFeed: false,
		lineHeight: 0,
		fontWeight: 1,
	}
}

// WithLineFeed 设置是否换行选项
func WithLineFeed(isLineFeed bool) AppendSplitlineOption {
	return func(o *appendSplitlineOptions) {
		o.isLineFeed = isLineFeed
	}
}

// WithSplitLineHeight 设置分割线行高选项
func WithSplitLineHeight(height int) AppendSplitlineOption {
	return func(o *appendSplitlineOptions) {
		o.lineHeight = height
	}
}

// WithSplitLineFontWeight 设置分割线字体粗细选项
func WithSplitLineFontWeight(weight int) AppendSplitlineOption {
	return func(o *appendSplitlineOptions) {
		o.fontWeight = weight
	}
}

// AppendSplitLine 添加分割行（使用可选参数）
func (i *ImgFont) AppendSplitLine(opts ...AppendSplitlineOption) *ImgFont {
	// 应用默认选项
	options := DefaultAppendSplitLineOptions()

	// 应用自定义选项
	for _, opt := range opts {
		opt(&options)
	}

	// 保存原始字体粗细
	originalFontWeight := i.FontWeight

	// 设置分割线的字体粗细
	i.SetFontWeight(options.fontWeight)

	// 根据图像宽度添加分割线
	if i.ImageWidth == 567 || i.ImageWidth == 568 {
		i.AppendText("---------------------------------------------------------------")
	}

	// 如果需要换行
	if options.isLineFeed {
		i.LineFeed(1, options.lineHeight)
	}

	// 恢复原始字体粗细
	i.SetFontWeight(originalFontWeight)

	return i
}

// ColumnConfig 列配置结构体
type ColumnConfig struct {
	Text       string // 文本内容
	Width      int    // 列宽度
	Align      int    // 对齐方式
	FontWeight int    // 字体粗细
	FontSize   int    // 字体大小
	LineHeight int    // 行高
}

// SetupColumns 设置列（在Go中没有实际作用，保留为兼容性接口）
func (i *ImgFont) SetupColumns(columns ...ColumnConfig) *ImgFont {
	return i
}

// PrintInColumns 打印多列内容
func (i *ImgFont) PrintInColumns(columns ...ColumnConfig) *ImgFont {
	// 保存原始设置
	imageWidth := i.ImageWidth
	height := i.TextTotalHeight
	if height == 0 {
		height = i.TextLineHeight
	}

	// 计算所有列的总宽度
	allColumnWidth := 0
	for _, column := range columns {
		allColumnWidth += column.Width
	}

	// 偏移量和其他参数
	deviationWidth := 0.0
	oldFontWeight := i.FontWeight
	oldFontSize := i.FontSize
	results := make([]map[string]float64, 0, len(columns))
	lineHeight := 0

	// 处理每一列
	for key, column := range columns {
		text := column.Text
		columnWidth := column.Width

		// 如果列宽为0，计算剩余宽度
		if columnWidth == 0 {
			columnWidth = imageWidth - allColumnWidth
		}

		// 获取对齐方式和其他参数
		align := column.Align
		fontWeight := column.FontWeight
		fontSize := column.FontSize

		// 如果有指定行高
		if column.LineHeight > 0 {
			lineHeight = int(float64(column.LineHeight) * 1.3)
		}

		// 处理居中和居右对齐
		if align == ALIGN_CENTER || (align == ALIGN_RIGHT && key != len(columns)-1) {
			i.ImageWidth = int(deviationWidth) + columnWidth + i.ImagePadding
		}

		// 设置字体粗细和大小
		if fontWeight > 0 {
			i.SetFontWeight(fontWeight)
		}
		if fontSize > 0 {
			i.SetFontSize(fontSize)
		}

		// 设置对齐方式，默认居左
		if align > 0 {
			i.SetAlignment(align)
		} else {
			i.SetAlignment(AlignLeft)
		}

		// 添加文本
		result := i.AddText(text, float64(height), float64(columnWidth-i.ImagePadding*2), deviationWidth)
		results = append(results, result)

		// 更新偏移量并恢复原始设置
		deviationWidth += float64(columnWidth)
		i.ImageWidth = imageWidth

		// 恢复字体设置
		if fontWeight > 0 {
			i.SetFontWeight(oldFontWeight)
		}
		if fontSize > 0 {
			i.SetFontSize(oldFontSize)
		}
	}

	// 找出最大高度
	maxHeight := 0.0
	for _, result := range results {
		if result["height"] > maxHeight {
			maxHeight = result["height"]
		}
	}

	// 更新总高度
	i.TextTotalHeight = int(maxHeight)

	// 重置行宽度
	i.TextLastLineUsedWidth = 0

	// 恢复居左对齐并添加换行
	i.SetAlignment(AlignLeft)
	i.LineFeed(1, lineHeight)

	return i
}

// SetTextLineHeight 设置文本行高
func (i *ImgFont) SetTextLineHeight(textLineHeight int) *ImgFont {
	if textLineHeight > 0 {
		i.TextLineHeight = textLineHeight
	}
	return i
}

// RecoverDefaultTextLineHeight 恢复默认文本行高
func (i *ImgFont) RecoverDefaultTextLineHeight() *ImgFont {
	i.TextLineHeight = i.DefaultTextLineHeight
	return i
}

// SetAlignment 设置对齐方式
func (i *ImgFont) SetAlignment(alignment int) *ImgFont {
	i.Alignment = alignment
	return i
}

/**
 * 换行
 * @access public
 * @param int $num
 * @return bool
 */
// LineFeed 添加换行
func (i *ImgFont) LineFeed(num int, lineHeight ...int) *ImgFont {
	// 重置当前行已使用宽度
	i.TextLastLineUsedWidth = 0

	// 计算行高
	addHeight := i.TextLineHeight * num
	if len(lineHeight) > 0 && lineHeight[0] > 0 {
		addHeight = lineHeight[0]
	}

	// 增加总高度
	i.TextTotalHeight += addHeight

	return i
}

// SetFontSize 设置字体大小
func (i *ImgFont) SetFontSize(fontSize int) *ImgFont {
	if fontSize > 0 {
		i.FontSize = fontSize
	}
	return i
}

// SetFontWeight 设置字体粗细
func (i *ImgFont) SetFontWeight(fontWeight int) *ImgFont {
	if fontWeight >= 1 {
		i.FontWeight = fontWeight * 2
	}
	return i
}

// SetTextSpacing 设置文本间距
func (i *ImgFont) SetTextSpacing(spacing float64) *ImgFont {
	i.TextSpacing = spacing
	return i
}

// RestoreDefault 恢复默认值
func (i *ImgFont) RestoreDefault() *ImgFont {
	i.RecoverDefaultTextLineHeight()
	i.SetFontWeight(1)
	i.SetFontSize(20)
	i.SetTextSpacing(1)
	i.SetAlignment(AlignLeft)
	return i
}

// SetImagePadding 设置图像边距
func (i *ImgFont) SetImagePadding(padding int) *ImgFont {
	i.ImagePadding = padding
	return i
}

// Save 保存图像并返回打印数据
func (i *ImgFont) Save(imageSrc string, reminderSound bool, openMoneybox bool) string {
	var data []string
	maxHeight := 2200
	height := i.TextTotalHeight + i.TextLineHeight
	headHeight := int(i.TextLineHeight/2) - 10

	// 分割高度
	var heights []int
	for height > maxHeight {
		heights = append(heights, maxHeight)
		height -= maxHeight
	}
	heights = append(heights, height)

	// 处理每一个分割的区域
	for key, h := range heights {
		h = int(h)

		// 创建色裁剪后的图像
		extraWidth := 0
		if i.Direction != 0 {
			extraWidth = 180
		}

		croppedRect := image.Rect(0, 0, i.ImageWidth+extraWidth, h)
		croppedImage := image.NewRGBA(croppedRect)

		// 填充白色背景
		draw.Draw(croppedImage, croppedImage.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

		// 复制分割区域
		srcPt := image.Point{0, 0}
		if key == 0 {
			srcPt = image.Point{0, headHeight}
		} else {
			srcPt = image.Point{0, maxHeight*key + headHeight}
		}

		// 将原始图像的部分区域复制到裁剪图像
		r := image.Rect(0, 0, i.ImageWidth, h)
		draw.Draw(croppedImage, r, i.Image, srcPt, draw.Over)

		// 旋转图像
		rotatedImage := i.rotateImage(croppedImage, float64(-i.Direction))

		// 如果需要保存图像
		if imageSrc != "" {
			// 确保目录存在
			dir := filepath.Dir(imageSrc)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				os.MkdirAll(dir, 0777)
			}

			// 保存图像
			outFile, err := os.Create(imageSrc)
			if err == nil {
				png.Encode(outFile, rotatedImage)
				outFile.Close()
			}
		}

		// 生成打印数据
		printData := string([]byte{29, 118, 48, 0}) + i.GetBytesFromBitMap(rotatedImage)
		data = append(data, printData)
	}

	// 合并打印数据
	printCode := strings.Join(data, "")

	// 添加提示音
	if reminderSound {
		printCode += "\x1B\x42\x03\x02"
	}

	// 添加切纸命令
	printCode += "\x1d\x56\x00"

	// 如果需要打开钱箱
	if openMoneybox {
		printCode += "\x10\x14\x01\x00\x01"
	}

	// 转换为16进制字符串
	return hex.EncodeToString([]byte(printCode))
}

// ToRasterFormat 转换为光栅格式
func (i *ImgFont) ToRasterFormat(img image.Image) string {
	printCode := string([]byte{29, 118, 48, 0})
	printCode += i.GetBytesFromBitMap(img)
	printCode += "\x1d\x56\x00"
	return hex.EncodeToString([]byte(printCode))
}

// 旋转图像的辅助函数
func (i *ImgFont) rotateImage(src image.Image, angle float64) image.Image {
	// 对于Go语言，我们需要实现自定义旋转逻辑
	// 对于简单起见，我们只处理-90度旋转和0度旋转两种情况

	if angle == 0 {
		return src
	}

	// 如果是-90度旋转
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// 创建一个新图像，宽高互换
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	// 旋转图像
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// 对于-90度旋转，坐标转换为: (h-1-y, x)
			dst.Set(h-1-y+bounds.Min.Y, x, src.At(x, y))
		}
	}

	return dst
}

// GetBytesFromBitMap 将位图转换为打印机可用的光栅格式字节
func (i *ImgFont) GetBytesFromBitMap(bitmap image.Image) string {
	// 获取图像的宽度和高度
	bounds := bitmap.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	bw := (width-1)/8 + 1

	// 初始化返回的字节数组
	rv := make([]byte, height*bw+4)

	// 设置头部信息（宽度和高度）
	// xL
	rv[0] = byte(bw & 0xFF)
	// xH
	rv[1] = byte((bw >> 8) & 0xFF)
	// yL
	rv[2] = byte(height & 0xFF)
	// yH
	rv[3] = byte((height >> 8) & 0xFF)

	// 获取图像的像素数据
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			// 获取像素颜色
			r, g, b, _ := bitmap.At(j+bounds.Min.X, i+bounds.Min.Y).RGBA()

			// 转换为8位值
			red := uint8(r >> 8)
			green := uint8(g >> 8)
			blue := uint8(b >> 8)

			// 转换为灰度值
			gray := RGB2Gray(red, green, blue)

			// 设置相应位
			index := (width*i+j)/8 + 4
			shift := 7 - ((width*i + j) % 8)
			rv[index] |= (gray << shift)
		}
	}

	// 返回字符串
	return string(rv)
}

// RGB2Gray 将RGB转换为二值灰度值
func RGB2Gray(red, green, blue uint8) byte {
	// 计算加权灰度值
	gray := float64(red)*0.299 + float64(green)*0.587 + float64(blue)*0.114

	// 如果上过阈值则为0，否则为1
	if gray > 127 {
		return 0
	}
	return 1
}
