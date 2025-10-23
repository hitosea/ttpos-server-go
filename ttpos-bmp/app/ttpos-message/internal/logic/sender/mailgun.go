// Package sender 实现消息发送服务
package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/mailgun/mailgun-go/v4"

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
	timeout   int
	mg        *mailgun.MailgunImpl
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
	s.timeout = g.Cfg().MustGet(ctx, "mailgun.timeout", 30).Int()

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

	// 创建 Mailgun 客户端
	s.mg = mailgun.NewMailgun(s.domain, s.apiKey)

	// 设置超时时间
	s.mg.SetAPIBase(mailgun.APIBase)

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
	// 创建邮件消息（使用独立构造函数）
	message := mailgun.NewMessage(
		fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail), // 发件人
		subject, // 主题
		"",      // 纯文本内容（留空）
	)

	// 添加收件人
	message.AddRecipient(recipient)

	// 设置 HTML 内容
	message.SetHTML(content)

	// 记录请求数据
	requestData, _ := json.Marshal(map[string]any{
		"from":    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		"to":      recipient,
		"subject": subject,
	})

	// 发送邮件
	startTime := time.Now()
	respMsg, respId, err := s.mg.Send(ctx, message)
	duration := time.Since(startTime)

	g.Log().Info(ctx, "Mailgun API 请求完成",
		"uuid", messageUuid,
		"duration", duration,
		"message_id", respId,
		"response", respMsg,
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
		return gerror.Wrap(err, "发送邮件失败")
	}

	// 构建响应数据
	responseData, _ := json.Marshal(map[string]any{
		"id":      respId,
		"message": respMsg,
	})

	// 创建发送成功日志
	_ = service.Message().CreateSendLog(ctx, &dto.MessageSendLogDTO{
		MessageUuid:  messageUuid,
		SendTime:     gtime.Now().Unix(),
		SendResult:   consts.SendResultSuccess,
		RequestData:  string(requestData),
		ResponseData: string(responseData),
	})

	g.Log().Info(ctx, "邮件发送成功",
		"uuid", messageUuid,
		"recipient", recipient,
		"mailgun_id", respId,
		"message", respMsg,
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
		"timeout":   fmt.Sprintf("%d", s.timeout),
	}
}
