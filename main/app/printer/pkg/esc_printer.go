// Package pkg 提供打印机相关功能
package pkg

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// 常量定义
const (
	// 对齐方式
	AlignLeft   = 0
	AlignCenter = 1
	AlignRight  = 2

	// HRI 位置
	HriPosAbove = 1
	HriPosBelow = 2

	// 图像处理算法
	DiffuseDither   = 0 // 扩散抖动
	ThresholdDither = 1 // 阈值抖动

	// 列标志
	ColumnFlagBwReverse = 0x08 // 黑白反转
	ColumnFlagBold      = 0x01 // 粗体
	ColumnFlagDoubleH   = 0x02 // 双倍高度
	ColumnFlagDoubleW   = 0x04 // 双倍宽度
)

// SunmiCloudPrinter 实现与Sunmi云打印机通信的功能
type Printers struct {
	// 私有属性
	dotsPerLine    int
	charHSize      int
	asciiCharWidth int
	cjkCharWidth   int
	orderData      string
	strs           string
	columnSettings [][]int
	appID          string
	appKey         string
}

// NewPrinterWithAuth 创建一个新的Printer实例，并指定认证信息
func NewPrinterWithAuth(dotsPerLine int, appID, appKey string) *Printers {
	if dotsPerLine == 0 {
		dotsPerLine = 384
	}

	return &Printers{
		dotsPerLine:    dotsPerLine,
		charHSize:      1,
		asciiCharWidth: 12,
		cjkCharWidth:   24,
		orderData:      "",
		strs:           "",
		columnSettings: [][]int{},
		appID:          appID,
		appKey:         appKey,
	}
}

// NewPrinter 创建一个新的Printer实例（appID和appKey可选）
func NewPrinter(dotsPerLine int, auth ...string) *Printers {
	appID := ""
	appKey := ""

	if len(auth) > 0 {
		appID = auth[0]
	}

	if len(auth) > 1 {
		appKey = auth[1]
	}

	return NewPrinterWithAuth(dotsPerLine, appID, appKey)
}

// GetOrderData 获取订单数据
func (p *Printers) GetOrderData() string {
	return p.orderData
}

// generateSign 生成签名
func (p *Printers) generateSign(bodyData string, timestamp string, nonce string) string {
	msg := bodyData + p.appID + timestamp + nonce
	h := hmac.New(sha256.New, []byte(p.appKey))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

// httpPost 发送HTTP POST请求
func (p *Printers) httpPost(path string, body map[string]interface{}) (map[string]interface{}, error) {
	url := "https://openapi.sunmi.com" + path
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := fmt.Sprintf("%06d", rand.Intn(1000000))

	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Sunmi-Appid", p.appID)
	req.Header.Set("Sunmi-Timestamp", timestamp)
	req.Header.Set("Sunmi-Nonce", nonce)
	req.Header.Set("Sunmi-Sign", p.generateSign(string(bodyData), timestamp, nonce))
	req.Header.Set("Source", "openapi")
	req.Header.Set("Content-Type", "application/json")

	// 创建一个自定义的Transport，强制使用HTTP/1.1
	transport := &http.Transport{
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 禁用证书验证
		},
		TLSHandshakeTimeout:   10 * time.Second, // TLS握手超时
		ResponseHeaderTimeout: 15 * time.Second, // 响应头超时
		ExpectContinueTimeout: 1 * time.Second,  // Expect: 100-continue超时：1秒
	}
	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// BindShop 绑定打印机到店铺
func (p *Printers) BindShop(sn string, shopID string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"sn":      sn,
		"shop_id": shopID,
	}
	return p.httpPost("/v2/printer/open/open/device/bindShop", body)
}

// UnbindShop 解绑打印机
func (p *Printers) UnbindShop(sn string, shopID string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"sn":      sn,
		"shop_id": shopID,
	}
	return p.httpPost("/v2/printer/open/open/device/unbindShop", body)
}

// OnlineStatus 获取打印机在线状态
func (p *Printers) OnlineStatus(sn string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"sn": sn,
	}
	return p.httpPost("/v2/printer/open/open/device/onlineStatus", body)
}

// ClearPrintJob 清除打印任务
func (p *Printers) ClearPrintJob(sn string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"sn": sn,
	}
	return p.httpPost("/v2/printer/open/open/device/clearPrintJob", body)
}

// PushVoice 推送语音通知
func (p *Printers) PushVoice(sn string, content string, cycle int, interval int, expireIn int) (map[string]interface{}, error) {
	if cycle <= 0 {
		cycle = 1
	}
	if interval <= 0 {
		interval = 2
	}
	if expireIn <= 0 {
		expireIn = 300
	}

	body := map[string]interface{}{
		"sn":        sn,
		"content":   content,
		"cycle":     cycle,
		"interval":  interval,
		"expire_in": expireIn,
	}
	return p.httpPost("/v2/printer/open/open/device/pushVoice", body)
}

// PushContent 推送打印内容
func (p *Printers) PushContent(sn string, tradeNo string, orderType int, count int, mediaText string, cycle int) (map[string]interface{}, error) {
	if orderType <= 0 {
		orderType = 1
	}
	if count <= 0 {
		count = 1
	}
	if cycle <= 0 {
		cycle = 1
	}

	body := map[string]interface{}{
		"sn":         sn,
		"trade_no":   tradeNo,
		"content":    p.orderData,
		"order_type": orderType,
		"count":      count,
		"media_text": mediaText,
		"cycle":      cycle,
	}
	return p.httpPost("/v2/printer/open/open/device/pushContent", body)
}

// PrintStatus 获取打印状态
func (p *Printers) PrintStatus(tradeNo string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"trade_no": tradeNo,
	}
	return p.httpPost("/v2/printer/open/open/ticket/printStatus", body)
}

// NewTicketNotify 新单提醒
func (p *Printers) NewTicketNotify(sn string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"sn": sn,
	}
	return p.httpPost("/v2/printer/open/open/ticket/newTicketNotify", body)
}

//////////////////////////////////////////////////
// Basic ESC/POS Commands
//////////////////////////////////////////////////

// Clear 清空打印数据
func (p *Printers) Clear() {
	p.orderData = ""
}

// AppendRawData 添加原始数据
func (p *Printers) AppendRawData(data string) {
	p.orderData += strings.ToLower(data)
}

// UnicodeToUTF8 将Unicode转换为UTF8字符串
func (p *Printers) UnicodeToUTF8(unicode int) string {
	if unicode <= 0x7f {
		n := unicode & 0x7f
		return fmt.Sprintf("%02x", n)
	}
	if unicode >= 0x80 && unicode <= 0x7ff {
		n := ((unicode>>6)&0x1f|0xc0)<<8 | ((unicode)&0x3f | 0x80)
		return fmt.Sprintf("%04x", n)
	}
	if unicode >= 0x800 && unicode <= 0xffff {
		n := ((unicode>>12)&0x0f|0xe0)<<16 | ((unicode>>6)&0x3f|0x80)<<8 | ((unicode)&0x3f | 0x80)
		return fmt.Sprintf("%06x", n)
	}
	if unicode >= 0x10000 && unicode <= 0x10ffff {
		n := ((unicode>>18)&0x07|0xf0)<<24 | ((unicode>>12)&0x3f|0x80)<<16 | ((unicode>>6)&0x3f|0x80)<<8 | ((unicode)&0x3f | 0x80)
		return fmt.Sprintf("%08x", n)
	}
	return ""
}

// UTF8ToUnicode 将UTF8字符转换为Unicode码点
func (p *Printers) UTF8ToUnicode(str string, size int) (int, int) {
	unicode := 0

	if size < 1 {
		return 0, 0
	}
	v0 := int(str[0])
	if (v0 & 0x80) == 0x00 {
		return v0, 1
	}

	if size < 2 {
		return 0, size
	}
	v1 := int(str[1])
	if (v0&0xe0) == 0xc0 && (v1&0xc0) == 0x80 {
		unicode = ((v0 & 0x1f) << 6) + (v1 & 0x3f)
		return unicode, 2
	}

	if size < 3 {
		return 0, size
	}
	v2 := int(str[2])
	if (v0&0xf0) == 0xe0 && (v1&0xc0) == 0x80 && (v2&0xc0) == 0x80 {
		unicode = ((v0 & 0x0f) << 12) + ((v1 & 0x3f) << 6) + (v2 & 0x3f)
		return unicode, 3
	}

	if size < 4 {
		return 0, size
	}
	v3 := int(str[3])
	if (v0&0xF8) == 0xf0 && (v1&0xc0) == 0x80 && (v2&0xc0) == 0x80 && (v3&0xc0) == 0x80 {
		unicode = ((v0 & 0x07) << 18) + ((v1 & 0x3f) << 12) + ((v2 & 0x3f) << 6) + (v3 & 0x3f)
		return unicode, 4
	}

	return 0, 1
}

// AppendUnicode 添加Unicode字符
func (p *Printers) AppendUnicode(unicode int, count int) {
	utf8 := p.UnicodeToUTF8(unicode)
	for i := 0; i < count; i++ {
		p.orderData += utf8
	}
}

// AppendTextWithReturn 添加文本并指定是否返回文本内容而不添加到打印队列
func (p *Printers) AppendTextWithReturn(text string, isReturn bool) string {
	if isReturn {
		p.strs = ""
		for i := 0; i < len(text); i++ {
			p.strs += fmt.Sprintf("%02x", text[i])
		}
		return p.strs
	} else {
		for i := 0; i < len(text); i++ {
			p.orderData += fmt.Sprintf("%02x", text[i])
		}
		return ""
	}
}

// AppendText 添加文本（可选参数isReturn，默认为false）
func (p *Printers) AppendText(text string, isReturn ...bool) string {
	text = NewPrintTextHelper().FilterCharacter(text)
	//
	returnResult := false
	if len(isReturn) > 0 {
		returnResult = isReturn[0]
	}
	return p.AppendTextWithReturn(text, returnResult)
}

// LineFeed 打印缓冲区数据并换行（可选参数n，默认为1）
func (p *Printers) LineFeed(n ...int) {
	lines := 1
	if len(n) > 0 && n[0] > 0 {
		lines = n[0]
	}
	for i := 0; i < lines; i++ {
		p.orderData += "0a"
	}
}

// RestoreDefaultSettings 恢复默认设置
func (p *Printers) RestoreDefaultSettings() {
	p.charHSize = 1
	p.orderData += "1b40"
}

// RestoreDefaultLineSpacing 恢复默认行间距
func (p *Printers) RestoreDefaultLineSpacing() {
	p.orderData += "1b32"
}

// SetLineSpacing 设置行间距
func (p *Printers) SetLineSpacing(n int) {
	if n >= 0 && n <= 255 {
		p.orderData += fmt.Sprintf("1b33%02x", n)
	}
}

// SetPrintModes 设置打印模式 [ESC !]
func (p *Printers) SetPrintModes(bold, doubleH, doubleW bool) {
	n := 0
	if bold {
		n |= 8
	}
	if doubleH {
		n |= 16
	}
	if doubleW {
		n |= 32
	}
	p.charHSize = 1
	if doubleW {
		p.charHSize = 2
	}
	p.orderData += fmt.Sprintf("1b21%02x", n)
}

// SetCharacterSizeWithApply 设置字符大小并指定是否应用 [GS !]
func (p *Printers) SetCharacterSizeWithApply(h, w int, apply bool) string {
	n := 0
	if h >= 1 && h <= 8 {
		n |= (h - 1)
	}
	if w >= 1 && w <= 8 {
		n |= (w - 1) << 4
		p.charHSize = w
	}
	d := fmt.Sprintf("1d21%02x", n)
	if apply {
		p.orderData += d
	}
	return d
}

// SetCharacterSize 设置字符大小（可选参数apply，默认为true）[GS !]
func (p *Printers) SetCharacterSize(h, w int, apply ...bool) string {
	applyChange := true
	if len(apply) > 0 {
		applyChange = apply[0]
	}
	return p.SetCharacterSizeWithApply(h, w, applyChange)
}

// HorizontalTab 跳到下一个制表符位置 [HT]
func (p *Printers) HorizontalTab(n int) {
	for i := 0; i < n; i++ {
		p.orderData += "09"
	}
}

// SetAbsolutePrintPosition 设置绝对打印位置 [ESC $]
func (p *Printers) SetAbsolutePrintPosition(n int) {
	if n >= 0 && n <= 65535 {
		p.orderData += fmt.Sprintf("1b24%02x%02x", (n & 0xff), ((n >> 8) & 0xff))
	}
}

// SetRelativePrintPosition 设置相对打印位置 [ESC \]
func (p *Printers) SetRelativePrintPosition(n int) {
	if n >= -32768 && n <= 32767 {
		p.orderData += fmt.Sprintf("1b5c%02x%02x", (n & 0xff), ((n >> 8) & 0xff))
	}
}

// SetAlignment 设置对齐方式 [ESC a]
func (p *Printers) SetAlignment(n int) {
	if n >= 0 && n <= 2 {
		p.orderData += fmt.Sprintf("1b61%02x", n)
	}
}

// SetUnderlineMode 设置下划线模式 [ESC -]
func (p *Printers) SetUnderlineMode(n int) {
	if n >= 0 && n <= 2 {
		p.orderData += fmt.Sprintf("1b2d%02x", n)
	}
}

// SetBlackWhiteReverseMode 设置黑白倒转模式 [GS B]
func (p *Printers) SetBlackWhiteReverseMode(enabled bool) {
	value := 0
	if enabled {
		value = 1
	}
	p.orderData += fmt.Sprintf("1d42%02x", value)
}

// SetUpsideDownMode 设置倒立模式 [ESC {]
func (p *Printers) SetUpsideDownMode(enabled bool) {
	value := 0
	if enabled {
		value = 1
	}
	p.orderData += fmt.Sprintf("1b7b%02x", value)
}

// CutPaper 切纸 [GS V m]
func (p *Printers) CutPaper(reminderSound bool) {
	// 添加提示音
	if reminderSound {
		p.AppendText("\x1B\x42\x03\x02", false)
	}
	p.orderData += fmt.Sprintf("1d56%02x", 0)
}

// PostponedCutPaper 延迟切纸 [GS V m n]
// 打印机在收到该命令后不会立即切纸，而是在读入(d + n)点行后才进行切纸，
// 其中d是打印位置和切纸位置之间的距离。
func (p *Printers) PostponedCutPaper(fullCut bool, n int) {
	if n >= 0 && n <= 255 {
		mode := 98 // 部分切纸
		if fullCut {
			mode = 97 // 全切纸
		}
		p.orderData += fmt.Sprintf("1d56%02x%02x", mode, n)
	}
}

// SetOrderData 设置对齐 [ESC a]
func (p *Printers) SetOrderData(n int) {
	if n >= 0 && n <= 2 {
		p.orderData += fmt.Sprintf("1b61%02x", n)
	}
}

//////////////////////////////////////////////////
// Sunmi Proprietary Commands
//////////////////////////////////////////////////

// SetCjkEncoding 设置CJK编码（当UTF-8模式禁用时有效）
//
//	n  编码
//
// ---  --------
//
//	 0  GB18030
//	 1  BIG5
//	11  Shift_JIS
//	12  JIS 0208
//	21  KS C 5601
//
// 128  禁用CJK模式
// 255  恢复默认值
func (p *Printers) SetCjkEncoding(n int) {
	if n >= 0 && n <= 255 {
		p.orderData += fmt.Sprintf("1d284503000601%02x", n)
	}
}

// SetUtf8Mode 设置UTF-8模式
//
//	n  模式
//
// ---  ----
//
//	0  禁用
//	1  启用
//
// 255  恢复默认值
func (p *Printers) SetUtf8Mode(n int) {
	if n >= 0 && n <= 255 {
		p.orderData += fmt.Sprintf("1d284503000603%02x", n)
	}
}

// SetHarfBuzzAsciiCharSize 设置矢量字体的拉丁字符大小
func (p *Printers) SetHarfBuzzAsciiCharSize(n int) {
	if n >= 0 && n <= 255 {
		p.asciiCharWidth = n
		p.orderData += fmt.Sprintf("1d28450300060a%02x", n)
	}
}

// SetHarfBuzzCjkCharSize 设置矢量字体的CJK字符大小
func (p *Printers) SetHarfBuzzCjkCharSize(n int) {
	if n >= 0 && n <= 255 {
		p.cjkCharWidth = n
		p.orderData += fmt.Sprintf("1d28450300060b%02x", n)
	}
}

// SetHarfBuzzOtherCharSize 设置矢量字体的其他字符大小
func (p *Printers) SetHarfBuzzOtherCharSize(n int) {
	if n >= 0 && n <= 255 {
		p.orderData += fmt.Sprintf("1d28450300060c%02x", n)
	}
}

// SelectAsciiCharFont 选择拉丁字符的字体
//
//	n  字体
//
// -----  ----
//
//	0  内置点阵字体
//	1  内置矢量字体
//
// >=128  第(n-128)个自定义矢量字体
func (p *Printers) SelectAsciiCharFont(n int) {
	if n >= 0 && n <= 255 {
		p.orderData += fmt.Sprintf("1d284503000614%02x", n)
	}
}

// SelectCjkCharFont 选择CJK字符的字体
//
//	n  字体
//
// -----  ----
//
//	0  内置点阵字体
//	1  内置矢量字体
//
// >=128  第(n-128)个自定义矢量字体
func (p *Printers) SelectCjkCharFont(n int) {
	if n >= 0 && n <= 255 {
		p.orderData += fmt.Sprintf("1d284503000615%02x", n)
	}
}

// SelectOtherCharFont 选择其他字符的字体
//
//	n  字体
//
// -----  ----
//
//	0,1  内置矢量字体
//
// >=128  第(n-128)个自定义矢量字体
func (p *Printers) SelectOtherCharFont(n int) {
	if n >= 0 && n <= 255 {
		p.orderData += fmt.Sprintf("1d284503000616%02x", n)
	}
}

// SetPrintDensity 设置打印密度
func (p *Printers) SetPrintDensity(n int) {
	if n >= 0 && n <= 255 {
		p.orderData += fmt.Sprintf("1d2845020007%02x", n)
	}
}

// SetPrintSpeed 设置打印速度
func (p *Printers) SetPrintSpeed(n int) {
	if n >= 0 && n <= 255 {
		p.orderData += fmt.Sprintf("1d2845020008%02x", n)
	}
}

// SetCutterMode 设置切纸模式
// n  模式
// -  ----
// 0  根据切纸命令执行全切或部分切纸
// 1  始终在任何切纸命令上执行部分切纸
// 2  始终在任何切纸命令上执行全切
// 3  不在任何切纸命令上切纸
func (p *Printers) SetCutterMode(n int) {
	if n >= 0 && n <= 255 {
		p.orderData += fmt.Sprintf("1d2845020010%02x", n)
	}
}

// ClearPaperNotTakenAlarm 清除纸张未取警报
func (p *Printers) ClearPaperNotTakenAlarm() {
	p.orderData += "1d2854010004"
}

//////////////////////////////////////////////////
// Print in Columns
//////////////////////////////////////////////////

// WidthOfChar 计算字符的宽度
func (p *Printers) WidthOfChar(c int) int {
	if c >= 0x00020 && c <= 0x0036f {
		return p.asciiCharWidth
	}
	if c >= 0x0ff61 && c <= 0x0ff9f {
		return p.cjkCharWidth / 2
	}
	if c == 0x02010 ||
		(c >= 0x02013 && c <= 0x02016) ||
		(c >= 0x02018 && c <= 0x02019) ||
		(c >= 0x0201c && c <= 0x0201d) ||
		(c >= 0x02025 && c <= 0x02026) ||
		(c >= 0x02030 && c <= 0x02033) ||
		c == 0x02035 ||
		c == 0x0203b {
		return p.cjkCharWidth
	}
	if (c >= 0x01100 && c <= 0x011ff) ||
		(c >= 0x02460 && c <= 0x024ff) ||
		(c >= 0x025a0 && c <= 0x027bf) ||
		(c >= 0x02e80 && c <= 0x02fdf) ||
		(c >= 0x03000 && c <= 0x0318f) ||
		(c >= 0x031a0 && c <= 0x031ef) ||
		(c >= 0x03200 && c <= 0x09fff) ||
		(c >= 0x0ac00 && c <= 0x0d7ff) ||
		(c >= 0x0f900 && c <= 0x0faff) ||
		(c >= 0x0fe30 && c <= 0x0fe4f) ||
		(c >= 0x1f000 && c <= 0x1f9ff) {
		return p.cjkCharWidth
	}
	if (c >= 0x0ff01 && c <= 0x0ff5e) ||
		(c >= 0x0ffe0 && c <= 0x0ffe5) {
		return p.cjkCharWidth
	}
	return p.asciiCharWidth
}

// WidthOfString 计算字符串的宽度
func (p *Printers) WidthOfString(str string) int {
	w := 0
	i := 0
	for i < len(str) {
		s := str[i:]
		c, size := p.UTF8ToUnicode(s, len(s))
		i += size
		w += p.WidthOfChar(c) * p.charHSize
	}
	return w
}

// SetupColumns 设置列
func (p *Printers) SetupColumns(columnWidths ...[]int) {
	p.columnSettings = nil
	remain := p.dotsPerLine
	for _, s := range columnWidths {
		if s[0] == 0 || s[0] > remain {
			s[0] = remain
		}
		p.columnSettings = append(p.columnSettings, s)
		remain -= s[0]
		if remain == 0 {
			return
		}
	}
}

// PrintInColumns 列式打印内容
func (p *Printers) PrintInColumns(columns ...string) {
	if len(p.columnSettings) == 0 {
		return
	}

	strcur := make([]string, 0)
	strrem := make([]string, 0)
	strwidth := make([]int, 0)

	numOfColumns := 0
	for i := 0; i < len(columns); i++ {
		if i == len(p.columnSettings) {
			break
		}
		strcur = append(strcur, "")
		strrem = append(strrem, columns[i])
		strwidth = append(strwidth, 0)
		numOfColumns++
	}

	for {
		done := true
		pos := 0

		for i := 0; i < numOfColumns; i++ {
			width := p.columnSettings[i][0]
			alignment := p.columnSettings[i][1]
			flag := p.columnSettings[i][2]

			if len(strrem[i]) == 0 {
				pos += width
				continue
			}

			done = false
			strcur[i] = ""
			strwidth[i] = 0

			for len(strrem[i]) > 0 {
				c, bytes := p.UTF8ToUnicode(strrem[i], len(strrem[i]))
				if c == 0x0a {
					strrem[i] = strrem[i][1:]
					break
				} else {
					w := p.WidthOfChar(c) * p.charHSize
					if (flag & ColumnFlagDoubleW) != 0 {
						w *= 2
					}
					if strwidth[i]+w > width {
						break
					} else {
						strcur[i] += strrem[i][:bytes]
						strwidth[i] += w
						strrem[i] = strrem[i][bytes:]
					}
				}
			}

			switch alignment {
			case AlignCenter:
				p.SetAbsolutePrintPosition(pos + (width-strwidth[i])/2)
			case AlignRight:
				p.SetAbsolutePrintPosition(pos + (width - strwidth[i]))
			default:
				p.SetAbsolutePrintPosition(pos)
			}

			if (flag & ColumnFlagBwReverse) != 0 {
				p.SetBlackWhiteReverseMode(true)
			}
			if (flag & (ColumnFlagBold | ColumnFlagDoubleH | ColumnFlagDoubleW)) != 0 {
				p.SetPrintModes((flag&ColumnFlagBold) != 0, (flag&ColumnFlagDoubleH) != 0, (flag&ColumnFlagDoubleW) != 0)
			}

			p.AppendText(strcur[i], false)

			if (flag & (ColumnFlagBold | ColumnFlagDoubleH | ColumnFlagDoubleW)) != 0 {
				p.SetPrintModes(false, false, false)
			}
			if (flag & ColumnFlagBwReverse) != 0 {
				p.SetBlackWhiteReverseMode(false)
			}
			pos += width
		}

		if !done {
			p.LineFeed(1)
		} else {
			break
		}
	}
}

//////////////////////////////////////////////////
// Barcode & QR Code Printing
//////////////////////////////////////////////////

// AppendBarcode 添加条形码
func (p *Printers) AppendBarcode(hriPos, height, moduleSize, barcodeType int, text string) {
	textLength := len(text)
	if textLength == 0 {
		return
	}
	if textLength > 255 {
		textLength = 255
	}
	if height < 1 {
		height = 1
	} else if height > 255 {
		height = 255
	}
	if moduleSize < 1 {
		moduleSize = 1
	} else if moduleSize > 6 {
		moduleSize = 6
	}

	p.orderData += fmt.Sprintf("1d48%02x", (hriPos & 3))
	p.orderData += "1d6600"
	p.orderData += fmt.Sprintf("1d68%02x", height)
	p.orderData += fmt.Sprintf("1d77%02x", moduleSize)
	p.orderData += fmt.Sprintf("1d6b%02x%02x", barcodeType, textLength)

	for i := 0; i < textLength; i++ {
		p.orderData += fmt.Sprintf("%02x", text[i])
	}
}

// AppendQRcode 添加二维码
func (p *Printers) AppendQRcode(moduleSize, ecLevel int, text string) {
	textLength := len(text)
	if textLength == 0 {
		return
	}
	if textLength > 65535 {
		textLength = 65535
	}
	if moduleSize < 1 {
		moduleSize = 1
	} else if moduleSize > 16 {
		moduleSize = 16
	}
	if ecLevel < 0 {
		ecLevel = 0
	} else if ecLevel > 3 {
		ecLevel = 3
	}

	p.orderData += "1d286b040031410000"
	p.orderData += fmt.Sprintf("1d286b03003143%02x", moduleSize)
	p.orderData += fmt.Sprintf("1d286b03003145%02x", ecLevel+48)
	p.orderData += fmt.Sprintf("1d286b%02x%02x315030", ((textLength + 3) & 0xFF), (((textLength + 3) >> 8) & 0xFF))

	for i := 0; i < textLength; i++ {
		p.orderData += fmt.Sprintf("%02x", text[i])
	}

	p.orderData += "1d286b0300315130"
}

//////////////////////////////////////////////////
// Image Printing
//////////////////////////////////////////////////

// DiffuseDither 灰度图像转单色图像 - 拋散抖动算法
func (p *Printers) DiffuseDither(srcData []byte, width, height int) {
	line1 := 0
	line2 := 1
	bmwidth := (width + 7) / 8

	// 初始化目标数据和行缓冲区
	dstData := make([]byte, bmwidth*height)
	linebuf := make([][]int, 2)
	linebuf[0] = make([]int, width*height)
	linebuf[1] = make([]int, width*height)

	// 复制第一行数据到行缓冲区
	for x := 0; x < width; x++ {
		linebuf[1][x] = int(srcData[x])
	}

	for y := 0; y < height; y++ {
		// 交换行缓冲区
		tmp := line1
		line1 = line2
		line2 = tmp
		notLastLine := y < height-1

		// 如果不是最后一行，读取下一行数据
		if notLastLine {
			p := (y + 1) * width
			for x := 0; x < width; x++ {
				linebuf[line2][x] = int(srcData[p+x])
			}
		}

		// 初始化当前行目标数据
		q := bmwidth * y
		for i := 0; i < bmwidth; i++ {
			dstData[q+i] = 0
		}

		b1 := 0
		b2 := 0
		mask := byte(0x80)

		// 处理每个像素
		for x := 1; x <= width; x++ {
			var err int
			if linebuf[line1][b1] < 128 { // 黑色像素
				err = linebuf[line1][b1]
				dstData[q] |= mask
			} else { // 白色像素
				err = linebuf[line1][b1] - 255
			}
			b1++

			// 移动或重置掩码
			if mask == 1 {
				q++
				mask = 0x80
			} else {
				mask >>= 1
			}

			// 计算误差扩散
			e7 := ((err * 7) + 8) >> 4
			e5 := ((err * 5) + 8) >> 4
			e3 := ((err * 3) + 8) >> 4
			e1 := err - (e7 + e5 + e3)

			// 将误差扩散到相邻像素
			if x < width {
				linebuf[line1][b1] += e7 // 向右像素扩散误差
			}
			if notLastLine {
				linebuf[line2][b2] += e5 // 下方像素
				if x > 1 {
					linebuf[line2][b2-1] += e3 // 左下方像素
				}
				if x < width {
					linebuf[line2][b2+1] += e1 // 右下方像素
				}
			}
			b2++
		}
	}

	// 添加打印图像的头部信息
	finalData := make([]byte, len(dstData)+8)
	finalData[0] = 0x1d
	finalData[1] = 0x76
	finalData[2] = 0x30
	finalData[3] = 0x00
	finalData[4] = byte(bmwidth & 0xff)
	finalData[5] = byte((bmwidth >> 8) & 0xff)
	finalData[6] = byte(height & 0xff)
	finalData[7] = byte((height >> 8) & 0xff)

	// 复制图像数据
	copy(finalData[8:], dstData)

	// 将数据转换为十六进制字符串添加到orderData
	for i := 0; i < len(finalData); i++ {
		p.orderData += fmt.Sprintf("%02x", finalData[i])
	}
}

// ThresholdDither 灰度图像转单色图像 - 阈值抖动算法
func (p *Printers) ThresholdDither(srcData []byte, width, height int) {
	bmwidth := (width + 7) / 8

	// 初始化空白目标数据
	dstData := make([]byte, bmwidth*height)

	pos := 0
	q := 0
	for y := 0; y < height; y++ {
		mask := byte(0x80)
		k := 0
		for x := 0; x < width; x++ {
			// 如果是黑色像素(低于128的灰度)
			if srcData[pos+x] < 128 {
				dstData[q+k] |= mask
			}

			// 移动或重置掩码
			if mask == 1 {
				k++
				mask = 0x80
			} else {
				mask >>= 1
			}
		}
		pos += width
		q += bmwidth
	}

	// 准备打印图像的头部信息
	finalData := make([]byte, len(dstData)+8)
	finalData[0] = 0x1d
	finalData[1] = 0x76
	finalData[2] = 0x30
	finalData[3] = 0x00
	finalData[4] = byte(bmwidth & 0xff)
	finalData[5] = byte((bmwidth >> 8) & 0xff)
	finalData[6] = byte(height & 0xff)
	finalData[7] = byte((height >> 8) & 0xff)

	// 复制图像数据
	copy(finalData[8:], dstData)

	// 将数据转换为十六进制字符串添加到orderData
	for i := 0; i < len(finalData); i++ {
		p.orderData += fmt.Sprintf("%02x", finalData[i])
	}
}

// 图像处理相关常量已在文件头部定义

// AppendImage 添加图像
func (p *Printers) AppendImage(imageFile string, mode int, maxWidth int) error {
	// 打开图像文件
	file, err := os.Open(imageFile)
	if err != nil {
		return fmt.Errorf("无法打开图像文件: %v", err)
	}
	defer file.Close()

	// 解码图像
	org_image, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("无法解码图像: %v", err)
	}

	// 获取原图像尺寸
	orgBounds := org_image.Bounds()
	orgWidth := orgBounds.Dx()
	orgHeight := orgBounds.Dy()

	// 如果最大宽度不合适，使用默认宽度
	if maxWidth <= 0 || maxWidth > p.dotsPerLine {
		maxWidth = p.dotsPerLine
	}

	// 计算缩放后的尺寸
	w := orgWidth
	h := orgHeight

	if w > maxWidth {
		h = maxWidth * h / w
		w = maxWidth
	}

	// 创建灰度图像数据
	grayscale := make([]byte, w*h)
	i := 0

	// 对图像进行采样和转换为灰度
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// 计算原图像对应点的位置
			srcX := x * orgWidth / w
			srcY := y * orgHeight / h

			// 获取像素颜色
			r, g, b, _ := org_image.At(srcX, srcY).RGBA()

			// 转换为8位颜色值
			r8 := byte(r >> 8)
			g8 := byte(g >> 8)
			b8 := byte(b >> 8)

			// 计算灰度值（与原算法相同的权重）
			grayscale[i] = byte((int(r8)*11 + int(g8)*16 + int(b8)*5) / 32)
			i++
		}
	}

	// 根据模式选择报错算法
	switch mode {
	case DiffuseDither:
		p.DiffuseDither(grayscale, w, h)
	case ThresholdDither:
		p.ThresholdDither(grayscale, w, h)
	default:
		// 默认使用阈值抖动
		p.ThresholdDither(grayscale, w, h)
	}

	return nil
}

//////////////////////////////////////////////////
// Page Mode Commands
//////////////////////////////////////////////////

// EnterPageMode [ESC L] 进入页面模式
func (p *Printers) EnterPageMode() {
	p.orderData += "1b4c"
}

// SetPrintAreaInPageMode [ESC W] 设置页面模式下的打印区域
// x, y: 打印区域原点
// w, h: 打印区域的宽度和高度
func (p *Printers) SetPrintAreaInPageMode(x, y, w, h int) {
	p.orderData += "1b57"
	p.orderData += fmt.Sprintf("%02x%02x", (x & 0xff), ((x >> 8) & 0xff))
	p.orderData += fmt.Sprintf("%02x%02x", (y & 0xff), ((y >> 8) & 0xff))
	p.orderData += fmt.Sprintf("%02x%02x", (w & 0xff), ((w >> 8) & 0xff))
	p.orderData += fmt.Sprintf("%02x%02x", (h & 0xff), ((h >> 8) & 0xff))
}

// SetPrintDirectionInPageMode [ESC T] 选择页面模式下的打印方向
// dir: 0:正常; 1:顺时钟旋转90度; 2:顺时钟旋转180度; 3:顺时钟旋转270度
func (p *Printers) SetPrintDirectionInPageMode(dir int) {
	if dir >= 0 && dir <= 3 {
		p.orderData += fmt.Sprintf("1b54%02x", dir)
	}
}

// SetAbsoluteVerticalPrintPositionInPageMode [GS $] 设置页面模式下的绝对垂直打印位置
func (p *Printers) SetAbsoluteVerticalPrintPositionInPageMode(n int) {
	if n >= 0 && n <= 65535 {
		p.orderData += fmt.Sprintf("1d24%02x%02x", (n & 0xff), ((n >> 8) & 0xff))
	}
}

// SetRelativeVerticalPrintPositionInPageMode [GS \] 设置页面模式下的相对垂直打印位置
func (p *Printers) SetRelativeVerticalPrintPositionInPageMode(n int) {
	if n >= -32768 && n <= 32767 {
		p.orderData += fmt.Sprintf("1d5c%02x%02x", (n & 0xff), ((n >> 8) & 0xff))
	}
}

// PrintAndExitPageMode [FF] 打印缓冲区数据并退出页面模式
func (p *Printers) PrintAndExitPageMode() {
	p.orderData += "0c"
}

// PrintInPageMode [ESC FF] 打印缓冲区数据（保持在页面模式）
func (p *Printers) PrintInPageMode() {
	p.orderData += "1b0c"
}

// ClearInPageMode [CAN] 清除缓冲区数据（保持在页面模式）
func (p *Printers) ClearInPageMode() {
	p.orderData += "18"
}

// ExitPageMode [ESC S] 退出页面模式并不打印地清除缓冲区数据
func (p *Printers) ExitPageMode() {
	p.orderData += "1b53"
}
