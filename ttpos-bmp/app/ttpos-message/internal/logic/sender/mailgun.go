// Package sender 实现消息发送服务
package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"ttpos-bmp/app/ttpos-message/internal/consts"
	"ttpos-bmp/app/ttpos-message/internal/model/dto"
	"ttpos-bmp/app/ttpos-message/internal/service"
)

// sMailgun Mailgun 邮件发送服务实现
type sMailgun struct {
	domain    string
	apiKey    string
	fromEmail string
	fromName  string
	apiBase   string
	timeout   int
	client    *http.Client
}

// Mailgun Mailgun 服务单例
var Mailgun = &sMailgun{}

func init() {
	service.RegisterMailgun(Mailgun)
}

// Init 初始化 Mailgun 服务
// 从配置文件中加载 Mailgun 相关配置
func (s *sMailgun) Init(ctx context.Context) error {
	s.domain = g.Cfg().MustGet(ctx, "mailgun.domain").String()
	s.apiKey = g.Cfg().MustGet(ctx, "mailgun.apiKey").String()
	s.fromEmail = g.Cfg().MustGet(ctx, "mailgun.fromEmail").String()
	s.fromName = g.Cfg().MustGet(ctx, "mailgun.fromName", "TTPOS System").String()
	s.apiBase = g.Cfg().MustGet(ctx, "mailgun.apiBase", "https://api.mailgun.net/v3").String()
	s.timeout = g.Cfg().MustGet(ctx, "mailgun.timeout", 30).Int()

	// 创建 HTTP 客户端
	s.client = &http.Client{
		Timeout: time.Duration(s.timeout) * time.Second,
	}

	// 验证必要配置
	if s.domain == "" {
		return gerror.New("Mailgun domain 未配置")
	}
	if s.apiKey == "" {
		return gerror.New("Mailgun apiKey 未配置")
	}
	if s.fromEmail == "" {
		return gerror.New("Mailgun fromEmail 未配置")
	}

	g.Log().Info(ctx, "Mailgun 服务初始化成功", "domain", s.domain)
	return nil
}

// SendEmail 发送邮件
// 调用 Mailgun API 发送邮件
// 参数：
//   - ctx: 上下文对象
//   - messageUuid: 消息UUID
//   - recipient: 收件人邮箱地址
//   - subject: 邮件主题
//   - content: 邮件内容（HTML格式）
//
// 返回：
//   - err: 错误信息
func (s *sMailgun) SendEmail(ctx context.Context, messageUuid, recipient, subject, content string) error {
	// 构建 API URL
	apiURL := fmt.Sprintf("%s/%s/messages", s.apiBase, s.domain)

	// 构建表单数据
	formData := url.Values{}
	formData.Set("from", fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail))
	formData.Set("to", recipient)
	formData.Set("subject", subject)
	formData.Set("html", content)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return gerror.Wrap(err, "创建HTTP请求失败")
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("api", s.apiKey)

	// 记录请求数据
	requestData, _ := json.Marshal(map[string]interface{}{
		"from":    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		"to":      recipient,
		"subject": subject,
	})

	// 发送请求
	startTime := time.Now()
	resp, err := s.client.Do(req)
	duration := time.Since(startTime)

	g.Log().Info(ctx, "Mailgun API 请求完成",
		"uuid", messageUuid,
		"duration", duration,
		"status", func() int {
			if resp != nil {
				return resp.StatusCode
			}
			return 0
		}(),
	)

	if err != nil {
		// 创建发送失败日志
		_ = service.Message().CreateSendLog(ctx, &dto.MessageSendLogDTO{
			MessageUuid:  messageUuid,
			SendTime:     gtime.Now().Unix(),
			SendResult:   consts.SendResultFailed,
			ErrorMessage: err.Error(),
			RequestData:  string(requestData),
		})
		return gerror.Wrap(err, "发送HTTP请求失败")
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return gerror.Wrap(err, "读取响应失败")
	}

	// 记录响应数据
	responseData := string(body)

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("Mailgun API 返回错误: %s (HTTP %d)", responseData, resp.StatusCode)

		// 创建发送失败日志
		_ = service.Message().CreateSendLog(ctx, &dto.MessageSendLogDTO{
			MessageUuid:  messageUuid,
			SendTime:     gtime.Now().Unix(),
			SendResult:   consts.SendResultFailed,
			ErrorMessage: errorMsg,
			RequestData:  string(requestData),
			ResponseData: responseData,
		})

		return gerror.New(errorMsg)
	}

	// 解析响应
	var mailgunResp dto.MailgunSendResponse
	if err := json.Unmarshal(body, &mailgunResp); err != nil {
		g.Log().Warning(ctx, "解析 Mailgun 响应失败", err)
	}

	// 创建发送成功日志
	_ = service.Message().CreateSendLog(ctx, &dto.MessageSendLogDTO{
		MessageUuid:  messageUuid,
		SendTime:     gtime.Now().Unix(),
		SendResult:   consts.SendResultSuccess,
		RequestData:  string(requestData),
		ResponseData: responseData,
	})

	g.Log().Info(ctx, "邮件发送成功",
		"uuid", messageUuid,
		"recipient", recipient,
		"mailgun_id", mailgunResp.Id,
	)

	return nil
}

// ValidateConfig 验证配置
func (s *sMailgun) ValidateConfig(ctx context.Context) error {
	if s.domain == "" || s.apiKey == "" || s.fromEmail == "" {
		return gerror.New("Mailgun 配置不完整")
	}
	return nil
}

// GetConfig 获取配置信息（用于调试）
func (s *sMailgun) GetConfig() map[string]string {
	return map[string]string{
		"domain":    s.domain,
		"fromEmail": s.fromEmail,
		"fromName":  s.fromName,
		"apiBase":   s.apiBase,
		"timeout":   fmt.Sprintf("%d", s.timeout),
	}
}
