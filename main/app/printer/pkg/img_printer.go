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
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/printer/pkg/fonts"
	"ttpos-server-go/app/printer/pkg/images"
	"ttpos-server-go/app/printer/pkg/rabbit"
	"ttpos-server-go/config"

	qrcode "github.com/skip2/go-qrcode"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
)

// 图片创建控制器 - 控制NewImgFont的并发
var (
	imgFontSemaphore     chan struct{} // 信号量通道
	maxImgFontConcurrent = 200         // 最大并发数
	imgFontMutex         sync.RWMutex  // 读写锁
	activeImgFontCount   = 0           // 当前活跃数量
)

// 初始化图片创建控制器
func init() {
	// 创建信号量通道
	imgFontSemaphore = make(chan struct{}, maxImgFontConcurrent)
}

// 全局字体管理器
type GlobalFontManager struct {
	fonts       map[string]*truetype.Font // 字体缓存
	widths      map[string]int            // 字符宽度缓存
	widthAccess map[string]int64          // 宽度缓存访问时间（LRU）
	mutex       sync.RWMutex              // 读写锁
	initialized bool                      // 是否已初始化
	maxWidths   int                       // 最大宽度缓存数量
}

var globalFontManager = &GlobalFontManager{
	fonts:       make(map[string]*truetype.Font),
	widths:      make(map[string]int),
	widthAccess: make(map[string]int64),
	maxWidths:   3000, // 减少最大缓存数量
}

// GetFont 获取字体（线程安全）
func (gfm *GlobalFontManager) GetFont(fontPath string) (*truetype.Font, error) {
	gfm.mutex.RLock()
	if font, exists := gfm.fonts[fontPath]; exists {
		gfm.mutex.RUnlock()
		return font, nil
	}
	gfm.mutex.RUnlock()

	// 需要加载字体
	gfm.mutex.Lock()
	defer gfm.mutex.Unlock()

	// 双重检查
	if font, exists := gfm.fonts[fontPath]; exists {
		return font, nil
	}

	// 加载字体
	fontBytes, err := fonts.GetFontData(fontPath)
	if err != nil {
		return nil, err
	}

	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}

	gfm.fonts[fontPath] = font
	return font, nil
}

// GetFace 获取字体表面（线程安全，每次创建新实例）
func (gfm *GlobalFontManager) GetFace(fontPath string, fontSize int) (font.Face, error) {
	fonts, err := gfm.GetFont(fontPath)
	if err != nil {
		return nil, err
	}

	// 每次创建新的字体表面实例，确保线程安全
	face := truetype.NewFace(fonts, &truetype.Options{
		Size:    float64(fontSize),
		DPI:     95,
		Hinting: font.HintingNone,
	})

	return face, nil
}

// GetCharWidth 获取字符宽度（线程安全，带LRU）
func (gfm *GlobalFontManager) GetCharWidth(fontPath string, fontSize int, char string) (int, bool) {
	cacheKey := fmt.Sprintf("%s_%d_%s", fontPath, fontSize, char)

	gfm.mutex.Lock() // 使用写锁因为要更新访问时间
	defer gfm.mutex.Unlock()

	if width, exists := gfm.widths[cacheKey]; exists {
		// 更新访问时间
		gfm.widthAccess[cacheKey] = time.Now().UnixNano()
		return width, true
	}

	return 0, false
}

// SetCharWidth 设置字符宽度（线程安全，带LRU清理）
func (gfm *GlobalFontManager) SetCharWidth(fontPath string, fontSize int, char string, width int) {
	cacheKey := fmt.Sprintf("%s_%d_%s", fontPath, fontSize, char)

	gfm.mutex.Lock()
	defer gfm.mutex.Unlock()

	// 如果缓存已满，清理最少使用的条目
	if len(gfm.widths) >= gfm.maxWidths {
		gfm.cleanupLRUWidthCache()
	}

	// 添加新缓存
	gfm.widths[cacheKey] = width
	gfm.widthAccess[cacheKey] = time.Now().UnixNano()
}

// cleanupLRUWidthCache 清理最少使用的宽度缓存（内部方法，调用时已加锁）
func (gfm *GlobalFontManager) cleanupLRUWidthCache() {
	if len(gfm.widths) == 0 {
		return
	}

	// 找到最旧的25%缓存项并删除
	type accessItem struct {
		key        string
		accessTime int64
	}

	items := make([]accessItem, 0, len(gfm.widthAccess))
	for key, accessTime := range gfm.widthAccess {
		items = append(items, accessItem{key, accessTime})
	}

	// 按访问时间排序
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].accessTime > items[j].accessTime {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// 删除最旧的25%
	removeCount := len(items) / 4
	for i := 0; i < removeCount; i++ {
		key := items[i].key
		delete(gfm.widths, key)
		delete(gfm.widthAccess, key)
	}
}

// PreloadCommonFonts 预加载常用字体
func (gfm *GlobalFontManager) PreloadCommonFonts() {
	commonFonts := []string{
		fonts.FontEN, fonts.FontZH, fonts.FontJA, fonts.FontKO,
	}

	// 只预加载字体对象，不预加载Face（因为Face不缓存）
	for _, fontPath := range commonFonts {
		gfm.GetFont(fontPath) // 预加载字体
	}

	gfm.mutex.Lock()
	gfm.initialized = true
	gfm.mutex.Unlock()
}

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
	// 本地缓存已移除，使用全局字体管理器
	// 分割高度
	SegmentationHeight int
	// 信号量控制（用于并发控制）
	hasSemaphore bool
}

// NewImgFont 创建一个新的ImgFont实例
func NewImgFont(imageWidth int, defaultTextLineHeight int, direction int) *ImgFont {
	// 获取信号量，控制并发数量
	imgFontSemaphore <- struct{}{}

	// 增加活跃计数
	imgFontMutex.Lock()
	activeImgFontCount++
	imgFontMutex.Unlock()

	img := &ImgFont{
		ImageWidth:            567, // 默认宽度
		ImageHeight:           0,
		ImagePadding:          24,
		DefaultTextLineHeight: 45,
		TextLineHeight:        45,
		TextSpacing:           1.0,
		TextTotalHeight:       0,
		TextLastLineUsedWidth: 0,
		Alignment:             AlignLeft,
		FontSize:              20,
		FontWeight:            1,
		Direction:             0,
		MySpecialFonts:        make(map[string]string),
		SegmentationHeight:    2200,
		hasSemaphore:          true, // 标记已获取信号量
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
	if direction == 1 {
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

	// 确保全局字体管理器已初始化
	if !globalFontManager.initialized {
		globalFontManager.PreloadCommonFonts()
	}

	// 创建图像
	img.createImg()

	// gc的时候Cleanup()
	runtime.SetFinalizer(img, func(i *ImgFont) {
		i.Cleanup()
	})

	return img
}

// ReleaseSemaphore 释放信号量（在ImgFont使用完毕后调用）
func (i *ImgFont) ReleaseSemaphore() {
	if i.hasSemaphore {
		// 释放信号量
		<-imgFontSemaphore

		// 减少活跃计数
		imgFontMutex.Lock()
		if activeImgFontCount > 0 {
			activeImgFontCount--
		}
		imgFontMutex.Unlock()

		// 标记已释放
		i.hasSemaphore = false
	}
}

// Cleanup 清理ImgFont实例内存（包含信号量释放）
func (i *ImgFont) Cleanup() {
	// 释放信号量
	i.ReleaseSemaphore()

	// 清理图像资源
	i.Image = nil
	i.MySpecialFonts = nil

	// 只在内存压力大时强制GC
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if m.Alloc > 1000*1024*1024 { // 超过1000MB时
		runtime.GC()
	}
}

// GetImgFontStatus 获取ImgFont控制器状态
func GetImgFontStatus() map[string]interface{} {
	imgFontMutex.RLock()
	defer imgFontMutex.RUnlock()

	return map[string]interface{}{
		"max_concurrent":   maxImgFontConcurrent,
		"active_count":     activeImgFontCount,
		"available":        maxImgFontConcurrent - activeImgFontCount,
		"semaphore_length": len(imgFontSemaphore),
	}
}

// SetMaxImgFontConcurrent 设置最大ImgFont并发数
func SetMaxImgFontConcurrent(max int) {
	if max > 0 && max <= 200 {
		imgFontMutex.Lock()
		maxImgFontConcurrent = max
		imgFontMutex.Unlock()

		// 重新创建信号量通道
		imgFontSemaphore = make(chan struct{}, max)
	}
}

// PrintImgFontStats 打印ImgFont统计信息
func PrintImgFontStats() {
	status := GetImgFontStatus()
	fmt.Printf("ImgFont状态: 最大并发 %d, 活跃数量 %d, 可用 %d, 信号量长度 %d\n",
		status["max_concurrent"],
		status["active_count"],
		status["available"],
		status["semaphore_length"])
}

// createImg 创建一个新的图像
func (i *ImgFont) createImg() {
	// 创建一个空白画布
	i.Image = image.NewRGBA(image.Rect(0, 0, i.ImageWidth, 1500))

	// 设置背景颜色为白色
	draw.Draw(i.Image, i.Image.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
}

// GetFontPath 根据字符获取对应的字体路径
func (i *ImgFont) GetFontPath(char string) string {
	// 中文
	chiPattern := regexp.MustCompile(`[\x{4E00}-\x{9FFF}\x{FF01}\x{FF0C}\x{FF08}\x{FF09}\x{FF1A}\x{FF5E}\x{2014}]`)
	if char != "ー" && (chiPattern.MatchString(char) || char == "【" || char == "】" || char == "。" || char == "、" || char == "-" || char == "￥" || char == "；" || char == "？" || char == "＋" || char == "：") {
		return fonts.FontZH
	}

	// 泰语
	thaiPattern := regexp.MustCompile(`[\p{Thai}]`)
	if thaiPattern.MatchString(char) || strings.Contains(char, "฿") {
		return fonts.FontTHExt
	}

	// 韩语
	hangulPattern := regexp.MustCompile(`[\p{Hangul}]`)
	if hangulPattern.MatchString(char) {
		return fonts.FontKO
	}

	// 日文
	japPattern := regexp.MustCompile(`[\p{Hiragana}\p{Katakana}\p{Han}★]+`)
	if char == "ー" || japPattern.MatchString(char) {
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

	// 老挝语
	laoPattern := regexp.MustCompile(`[\x{0E80}-\x{0EFF}]`)
	if laoPattern.MatchString(char) {
		return fonts.FontLO
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
// GetFontWeight 获取字体宽度（优化版）
func (i *ImgFont) GetFontWeight(fontSize int, char string) int {
	// 特殊字符'ြ'的处理
	if char == "ြ" {
		return 5
	}

	// 获取字体路径
	fontPath := i.GetFontPath(char)

	// 先检查全局缓存
	if width, exists := globalFontManager.GetCharWidth(fontPath, fontSize, char); exists {
		return width
	}

	// 需要计算宽度
	contextWidth := i.fastGetCharWidth(fontPath, fontSize, char)

	// 缅甸语特殊处理
	if fontPath == fonts.FontMY || fontPath == fonts.FontMY2 {
		if contextWidth < 0 {
			contextWidth = 0
		}
	}

	// 存入全局缓存
	globalFontManager.SetCharWidth(fontPath, fontSize, char, contextWidth)

	return contextWidth
}

// fastGetCharWidth 高性能字符宽度计算（改进UTF-8处理）
func (i *ImgFont) fastGetCharWidth(fontPath string, fontSize int, char string) int {
	// 如果文本为空，返回默认值
	if len(char) == 0 {
		return fontSize
	}

	// 使用全局字体管理器获取字体表面
	face, err := globalFontManager.GetFace(fontPath, fontSize)
	if err != nil {
		return fontSize
	}

	// 正确处理UTF-8字符
	runes := []rune(char)
	if len(runes) == 0 {
		return fontSize
	}

	// 计算第一个Unicode字符的宽度
	advance, ok := face.GlyphAdvance(runes[0])
	if !ok {
		// 如果无法获取字形宽度，尝试使用字体度量
		metrics := face.Metrics()
		return metrics.Height.Round() / 2 // 使用高度的一半作为默认宽度
	}

	// 应用缩放系数
	width := advance.Round()
	scaleFactor := 0.99
	if fontPath == fonts.FontTH || fontPath == fonts.FontTHExt {
		scaleFactor = 1.1
	}

	return int(float64(width) * scaleFactor)
}

// getCharWidth 使用freetype获取字符宽度 - 模拟PHP的imagettfbbox的行为（保留兼容性）
func (i *ImgFont) getCharWidth(fontPath string, fontSize int, text string) int {
	// 如果文本为空，返回默认值
	if len(text) == 0 {
		return fontSize
	}

	// 使用全局字体管理器获取字体
	f, err := globalFontManager.GetFont(fontPath)
	if err != nil {
		fmt.Println("获取字体失败:", err)
		return fontSize
	}

	// 创建字体选项
	opt := truetype.Options{
		Size:    float64(fontSize),
		DPI:     95,
		Hinting: font.HintingNone,
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

	if fontPath == fonts.FontTH || fontPath == fonts.FontTHExt {
		scaleFactor = 1.1
	}

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
	// 智能估算所需高度
	estimatedHeight := int(height) + i.TextLineHeight*2 + 100 // 减少缓冲空间
	currentImageHeight := i.Image.Bounds().Dy()
	if estimatedHeight > currentImageHeight {
		// 智能计算新高度：按需增长，避免过度分配
		requiredGrowth := estimatedHeight - currentImageHeight
		growthIncrement := 500 // 基础增长单位
		// 根据需要动态调整增长量
		if requiredGrowth > 1000 {
			growthIncrement = 1000
		}
		newHeight := currentImageHeight + ((requiredGrowth/growthIncrement)+1)*growthIncrement
		i.expandImageHeight(newHeight)
	}

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
			tmpw = i.GetTextWidth(fontSize, tmpStr)
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

		// 获取当前字符的宽度 - 直接使用GetFontWeight，已经有缓存机制
		charWidth := float64(i.GetFontWeight(fontSize, char)) * i.TextSpacing

		// 非居左对齐的处理
		if i.Alignment != AlignLeft {
			// 获取当前到最后的所有字符
			lastText := segments[key:]

			// 计算最后一行的宽度 - 使用优化的GetTextWidth函数
			lastTextWidth := i.GetTextWidth(fontSize, strings.Join(lastText, ""))

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

// GetTextWidth 批量计算文本宽度（优化版）
func (i *ImgFont) GetTextWidth(fontSize int, text string) float64 {
	if text == "" {
		return 0
	}

	// 计算宽度 - 直接计算，不缓存文本宽度
	totalWidth := 0.0
	for _, char := range i.GetFontArrays(text) {
		totalWidth += float64(i.GetFontWeight(fontSize, char))
	}

	return totalWidth * i.TextSpacing
}

// drawText 使用freetype绘制文本（优化版）
func (i *ImgFont) drawText(text, fontPath string, fontSize, fontWeight int, x, y int, textColor color.RGBA) {
	// 从全局字体管理器获取字体
	f, err := globalFontManager.GetFont(fontPath)
	if err != nil {
		fmt.Println("获取字体失败:", err)
		return
	}

	// 先绘制描边（黑色或深色）
	outlineColor := color.RGBA{0, 0, 0, 255} // 黑色描边

	// 创建上下文
	ctx := freetype.NewContext()
	ctx.SetDPI(95)
	ctx.SetFont(f)
	ctx.SetClip(i.Image.Bounds())
	ctx.SetDst(i.Image)
	ctx.SetFontSize(float64(fontSize))
	ctx.SetSrc(image.NewUniform(outlineColor))
	ctx.SetHinting(0) // font.HintingNone

	// 后绘制原始文本
	ctx.SetSrc(image.NewUniform(textColor))
	pt := freetype.Pt(x, y+fontSize)

	// 对于58mm打印纸，减少重复绘制避免重影
	// 只有在初始ImageWidth为58mm时才应用此逻辑，避免PrintInColumns中临时ImageWidth的误判
	maxWeight := fontWeight
	if fontWeight > 2 {
		// 检查是否为58mm纸张：初始ImageWidth应该是384像素
		// 通过TextSpacing来判断更准确，因为58mm模板设置了TextSpacing=1.1
		if i.TextSpacing > 1.05 { // 58mm模板的TextSpacing是1.1
			maxWeight = 2 // 限制最大重复次数
		}
	}

	for j := 0; j < maxWeight; j++ {
		_, err := ctx.DrawString(text, pt)
		if err != nil {
			fmt.Println("绘制原始文本失败:", err)
		}
		// 对于多次绘制，稍微偏移位置避免完全重叠
		if j > 0 && maxWeight > 1 {
			pt.X += 64 // 1像素偏移 (64 = 1 << 6)
		}
	}

}

// SetSegmentationHeight 设置分割高度
func (i *ImgFont) SetSegmentationHeight(height int) *ImgFont {
	// i.SegmentationHeight = height
	return i
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
	imgData, err := os.ReadFile(imgPath)
	if err != nil {
		fmt.Println("读取图片文件失败:", err)
		return i
	}

	// 解码图片
	srcImg, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		fmt.Println("解码图片失败:", err)
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

	// 计算图片位置 (居右显示)
	if i.Alignment == AlignRight {
		x = i.ImageWidth - width - i.ImagePadding
	}

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

// AppendEmbeddedImg 添加嵌入式图片
func (i *ImgFont) AppendEmbeddedImg(imageName string, size int, isRoundness bool, topHeight int) *ImgFont {
	// 从嵌入式资源读取图片数据
	imgData, err := images.GetImageData(imageName)
	if err != nil {
		fmt.Println("读取嵌入式图片失败:", err)
		return i
	}

	// 解码图片
	srcImg, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		fmt.Println("解码图片失败:", err)
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

	// 计算图片位置 (居右显示)
	if i.Alignment == AlignRight {
		x = i.ImageWidth - width - i.ImagePadding
	}

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
		// 检查并去除可能的前缀
		base64Data := data
		// 查找常见的Base64前缀
		prefixIndex := strings.Index(data, ";base64,")
		if prefixIndex != -1 {
			base64Data = data[prefixIndex+8:] // 跳过";base64,"及其前面的内容
		}

		// 如果是Base64编码的图片数据
		decoded, decodeErr := base64.StdEncoding.DecodeString(base64Data)
		if decodeErr != nil {
			fmt.Println("解码Base64图片错误:", decodeErr)
			// 尝试使用URL安全的Base64解码
			decoded, decodeErr = base64.URLEncoding.DecodeString(base64Data)
			if decodeErr != nil {
				fmt.Println("URL安全Base64解码也失败:", decodeErr)
				return i
			}
		}

		// 解码图片
		qrImg, _, err = image.Decode(bytes.NewReader(decoded))
		if err != nil {
			fmt.Println("解码图片错误", err)
			return i
		}

		// 调整总高度
		i.TextTotalHeight = i.TextTotalHeight - (margin * 2)
	} else {
		// 创建二维码
		qr, err := qrcode.New(data, qrcode.Medium)
		if err != nil {
			fmt.Println("Error generating QR code:", err)
			return i
		}
		// 设置二维码属性
		qr.DisableBorder = false
		// 生成指定大小的二维码图片
		qrImg = qr.Image(size)
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
	i.TextTotalHeight += height - 40
	i.TextLastLineUsedWidth += width

	// 添加换行
	i.LineFeed(1)

	return i
}

// 添加条形码
func (i *ImgFont) AppendBarcode(data string, w, h int) *ImgFont {
	// Create the barcode
	qrCode, _ := code128.Encode(data)

	// Scale the barcode to 200x200 pixels
	barcodeImg, err := barcode.Scale(qrCode, w, h)
	if err != nil {
		fmt.Println("生成条形码错误", err)
		return i
	}

	// 获取条形码尺寸
	width := barcodeImg.Bounds().Dx()
	height := barcodeImg.Bounds().Dy()

	// 计算条形码位置 (水平居中)
	x := (i.ImageWidth - width) / 2

	// 将条形码绘制到目标图像上
	draw.Draw(
		i.Image,
		image.Rect(x, i.TextTotalHeight, x+width, i.TextTotalHeight+height),
		barcodeImg,
		image.Point{0, 0},
		draw.Over,
	)

	// 更新文本总高度和最后一行已使用宽度
	i.TextTotalHeight += height - 40
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
		i.AppendText("----------------------------------------------------------------------")
	} else if i.ImageWidth == 384 {
		// 58mm 打印纸适配的分割线 (384像素) - 减少到44个短横线避免换行
		i.AppendText("----------------------------------------------")
	} else if i.ImageWidth == 410 {
		// 58mm 打印纸适配的分割线 (410像素)
		i.AppendText("--------------------------------------------------")
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
	Text         string // 文本内容
	Width        int    // 列宽度
	ContentWidth int    // 内容宽度
	Align        int    // 对齐方式
	FontWeight   int    // 字体粗细
	FontSize     int    // 字体大小
	LineHeight   int    // 行高
	RightPadding int    // 内边距
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
		contentWidth := column.ContentWidth
		if column.ContentWidth == 0 {
			contentWidth = columnWidth
		}

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
		if align == AlignCenter || (align == AlignRight && key != len(columns)-1) {
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
		result := i.AddText(text, float64(height), float64(contentWidth-i.ImagePadding*2), deviationWidth)
		results = append(results, result)

		// 更新偏移量并恢复原始设置
		deviationWidth += float64(columnWidth) + float64(column.RightPadding)
		i.ImageWidth = imageWidth

		// 恢复字体设置
		i.SetFontWeight(oldFontWeight)
		i.SetFontSize(oldFontSize)
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

// SetTextTotalHeight 设置文本总高度
func (i *ImgFont) SetTextTotalHeight(height int) *ImgFont {
	i.TextTotalHeight = i.TextTotalHeight - height
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
		i.FontWeight = fontWeight
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

// expandImageHeight 扩大图片高度
func (i *ImgFont) expandImageHeight(newHeight int) {
	// 如果新高度不大于当前高度，则不需要扩大
	currentHeight := i.Image.Bounds().Dy()
	if newHeight <= currentHeight {
		return
	}

	// 保存原图片引用用于复制
	oldImage := i.Image

	// 创建新的更大的图片
	newImage := image.NewRGBA(image.Rect(0, 0, i.ImageWidth, newHeight))

	// 设置背景颜色为白色
	draw.Draw(newImage, newImage.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// 将原图片内容复制到新图片中
	draw.Draw(newImage, oldImage.Bounds(), oldImage, image.Point{}, draw.Src)

	// 替换原图片
	i.Image = newImage

	// 显式清空原图片引用，帮助GC回收
	oldImage = nil

	// 使用全局字体管理器，无需本地清理
}

// Save 保存图像并返回打印数据
func (i *ImgFont) Save(imageSrc string, reminderSound bool, openMoneybox int) string {
	// 清理ImgFont实例内存
	defer i.Cleanup()

	if config.Server.Mode == constant.ServerModeDebug && imageSrc == "" {
		imageSrc = "./tmp/printer/test_img.png"
	}

	maxHeight := i.SegmentationHeight
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
	var data []string
	data = append(data, string([]byte{0x1B, 0x40}))
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
	if openMoneybox == 1 {
		printCode += "\x10\x14\x01\x00\x01"
	} else if openMoneybox == 2 {
		// 使用字节表示方式代替PHP的chr()函数
		printCode += "\x1B\x70\x00\x19\xFA"
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
	// 只处理标准旋转角度：0度、-90度、90度

	// 标准化角度到 0, -90, 90
	normalizedAngle := int(angle) % 360
	if normalizedAngle < 0 {
		normalizedAngle += 360
	}

	// 只有90度的倍数才旋转，其他情况直接返回原图
	switch normalizedAngle {
	case 0:
		return src
	case 90, -270:
		// 90度顺时针旋转 = -90度逆时针旋转
		return i.rotate90(src, true)
	case 270, -90:
		// 270度顺时针旋转 = -90度逆时针旋转
		return i.rotate90(src, false)
	case 180, -180:
		// 180度旋转
		return i.rotate180(src)
	default:
		// 非标准角度，直接返回原图避免错误
		return src
	}
}

// rotate90 90度旋转
func (i *ImgFont) rotate90(src image.Image, clockwise bool) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// 创建一个新图像，宽高互换
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	// 旋转图像
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if clockwise {
				// 顺时针90度: (y, w-1-x)
				dst.Set(y-bounds.Min.Y, w-1-(x-bounds.Min.X), src.At(x, y))
			} else {
				// 逆时针90度: (h-1-y, x)
				dst.Set(h-1-(y-bounds.Min.Y), x-bounds.Min.X, src.At(x, y))
			}
		}
	}

	return dst
}

// rotate180 180度旋转
func (i *ImgFont) rotate180(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// 创建一个新图像，尺寸不变
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	// 旋转图像
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// 180度旋转: (w-1-x, h-1-y)
			dst.Set(w-1-(x-bounds.Min.X), h-1-(y-bounds.Min.Y), src.At(x, y))
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

	// 计算字节宽度，确保8字节对齐
	bw := (width + 7) / 8 // 向上取整到8的倍数

	// 预分配足够的空间，并初始化为0
	rv := make([]byte, height*bw+4)

	// 设置头部信息（宽度和高度）
	rv[0] = byte(bw & 0xFF)            // xL
	rv[1] = byte((bw >> 8) & 0xFF)     // xH
	rv[2] = byte(height & 0xFF)        // yL
	rv[3] = byte((height >> 8) & 0xFF) // yH

	// 优化：使用常量数组避免重复计算
	var shifts = [8]byte{7, 6, 5, 4, 3, 2, 1, 0}

	// 优化：使用类型断言加速常见图像类型的访问
	switch img := bitmap.(type) {
	case *image.RGBA:
		// 直接访问 RGBA 图像的像素数据 - 高度优化版本
		pix := img.Pix
		stride := img.Stride
		for y := 0; y < height; y++ {
			rowOffset := y * stride
			baseIndex := y*bw + 4 // 预计算行基础索引

			for x := 0; x < width; x++ {
				pixelOffset := rowOffset + x*4

				// 使用位运算快速灰度计算
				r, g, b := pix[pixelOffset], pix[pixelOffset+1], pix[pixelOffset+2]
				// 快速灰度计算：(77*R + 151*G + 28*B) >> 8
				gray := (77*uint32(r) + 151*uint32(g) + 28*uint32(b)) >> 8

				// 快速二值化
				var bit byte
				if gray > 127 {
					bit = 0
				} else {
					bit = 1
				}

				// 优化的位设置
				byteIndex := baseIndex + x/8
				bitPos := x % 8
				rv[byteIndex] |= bit << shifts[bitPos]
			}
		}
	default:
		// 对于其他类型的图像，使用通用方法
		for y := 0; y < height; y++ {
			baseIndex := y*bw + 4 // 预计算行基础索引

			for x := 0; x < width; x++ {
				// 获取像素颜色
				r, g, b, _ := bitmap.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()

				// 转换为8位值并快速灰度计算
				red := uint8(r >> 8)
				green := uint8(g >> 8)
				blue := uint8(b >> 8)

				// 快速灰度计算
				gray := (77*uint32(red) + 151*uint32(green) + 28*uint32(blue)) >> 8

				// 快速二值化
				var bit byte
				if gray > 127 {
					bit = 0
				} else {
					bit = 1
				}

				// 优化的位设置
				byteIndex := baseIndex + x/8
				bitPos := x % 8
				rv[byteIndex] |= bit << shifts[bitPos]
			}
		}
	}

	// 返回字符串
	return string(rv)
}

// RGB2Gray 将RGB转换为二值灰度值
func RGB2Gray(red, green, blue uint8) byte {
	// 使用整数运算替代浮点运算，提高性能
	// 使用标准的灰度转换权重：R:0.299, G:0.587, B:0.114
	// 转换为整数计算: 30% R + 59% G + 11% B
	gray := (30*uint32(red) + 59*uint32(green) + 11*uint32(blue)) / 100

	// 如果上过阈值则为0，否则为1
	if gray > 127 {
		return 0
	}
	return 1
}
