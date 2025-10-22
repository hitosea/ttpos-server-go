// Package message 实现消息服务的业务逻辑
package message

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"ttpos-bmp/app/ttpos-message/internal/consts"
	"ttpos-bmp/app/ttpos-message/internal/model/dto"
	"ttpos-bmp/app/ttpos-message/internal/service"
)

// sMessage 消息服务实现
type sMessage struct{}

// Message 消息服务单例
var Message = new(sMessage)

func init() {
	service.RegisterMessage(Message)
}

// SendMessage 发送消息
// 接收消息发送请求，创建消息记录并提交到消息队列
// 参数：
//   - ctx: 上下文对象
//   - in: 发送消息输入参数
//
// 返回：
//   - out: 发送结果
//   - err: 错误信息
func (s *sMessage) SendMessage(ctx context.Context, in *dto.SendMessageInput) (out *dto.SendMessageOutput, err error) {
	out = &dto.SendMessageOutput{}

	// 参数验证
	if err = s.validateSendMessageInput(ctx, in); err != nil {
		out.Success = false
		out.Message = err.Error()
		return out, err
	}
	// 检查消息是否已存在（幂等性）
	var existingRecord *dto.MessageRecordDTO
	existingRecord, err = s.GetMessageByUuid(ctx, in.MessageUuid)
	if err != nil && !gerror.Equal(err, gerror.New(consts.ErrMsgMessageNotFound)) {
		out.Success = false
		out.Message = "检查消息是否存在失败"
		return out, err
	}
	if existingRecord != nil {
		out.Success = true
		out.Message = "消息已存在，无需重复提交"
		out.MessageUuid = in.MessageUuid
		return out, nil
	}

	// 查询模板信息
	var template *dto.MessageTemplateDTO
	template, err = s.GetTemplateById(ctx, in.TemplateId)
	if err != nil {
		out.Success = false
		out.Message = consts.ErrMsgTemplateNotFound
		return out, err
	}

	// 检查模板状态
	if template.Status != consts.TemplateStatusEnabled {
		err = gerror.New(consts.ErrMsgTemplateDisabled)
		out.Success = false
		out.Message = err.Error()
		return out, err
	}

	// 检查模板类型与消息类型是否匹配
	if template.TemplateType != in.MessageType {
		err = gerror.Newf("模板类型(%s)与消息类型(%s)不匹配", template.TemplateType, in.MessageType)
		out.Success = false
		out.Message = err.Error()
		return out, err
	}

	// 渲染模板内容
	var content string
	content, err = s.renderTemplate(ctx, template.TemplateContent, in.MessageArgs)
	if err != nil {
		out.Success = false
		out.Message = consts.ErrMsgTemplateRender
		return out, err
	}

	// 渲染主题（如果有）
	subject := in.Subject
	if subject == "" && template.TemplateSubject != "" {
		subject, err = s.renderTemplate(ctx, template.TemplateSubject, in.MessageArgs)
		if err != nil {
			g.Log().Warning(ctx, "渲染主题失败", err)
			subject = template.TemplateSubject
		}
	}

	// 创建消息记录
	now := gtime.Now().Unix()
	record := &dto.MessageRecordDTO{
		Uuid:         in.MessageUuid,
		TemplateId:   in.TemplateId,
		MessageType:  in.MessageType,
		Recipient:    in.Recipient,
		Subject:      subject,
		Content:      content,
		MessageArgs:  in.MessageArgs,
		Status:       consts.MessageStatusPending,
		CompanyUuid:  in.CompanyUuid,
		OperatorUuid: in.OperatorUuid,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 保存到数据库
	err = s.CreateMessageRecord(ctx, record)
	if err != nil {
		out.Success = false
		out.Message = "创建消息记录失败"
		return out, err
	}

	// 提交到消息队列（只发送 UUID 和类型，详细内容从数据库获取）
	err = service.Queue().PublishMessage(ctx, &dto.RocketMQMessage{
		MessageUuid: in.MessageUuid,
		MessageType: in.MessageType,
	})
	if err != nil {
		g.Log().Error(ctx, "提交消息到队列失败", err)
		// 更新消息状态为失败
		_ = s.UpdateMessageStatus(ctx, in.MessageUuid, consts.MessageStatusFailed, err.Error())
		out.Success = false
		out.Message = consts.ErrMsgRocketMQError
		return out, err
	}

	out.Success = true
	out.Message = "消息提交成功"
	out.MessageUuid = in.MessageUuid
	g.Log().Info(ctx, "消息提交成功", "uuid", in.MessageUuid)

	return out, nil
}

// GetMessageStatus 查询消息状态
// 根据消息UUID查询消息的发送状态和详细信息
// 参数：
//   - ctx: 上下文对象
//   - in: 查询消息状态输入参数
//
// 返回：
//   - out: 消息状态信息
//   - err: 错误信息
func (s *sMessage) GetMessageStatus(ctx context.Context, in *dto.GetMessageStatusInput) (out *dto.GetMessageStatusOutput, err error) {
	out = &dto.GetMessageStatusOutput{}

	// 查询消息记录
	var record *dto.MessageRecordDTO
	record, err = s.GetMessageByUuid(ctx, in.MessageUuid)
	if err != nil {
		out.Success = false
		out.Message = consts.ErrMsgMessageNotFound
		return out, err
	}

	out.Success = true
	out.Message = "查询成功"
	out.MessageInfo = record

	return out, nil
}

// ResendMessage 重发消息
// 针对发送失败的消息，支持重新发送
// 参数：
//   - ctx: 上下文对象
//   - in: 重发消息输入参数
//
// 返回：
//   - out: 重发结果
//   - err: 错误信息
func (s *sMessage) ResendMessage(ctx context.Context, in *dto.ResendMessageInput) (out *dto.ResendMessageOutput, err error) {
	out = &dto.ResendMessageOutput{}

	// 查询消息记录
	var record *dto.MessageRecordDTO
	record, err = s.GetMessageByUuid(ctx, in.MessageUuid)
	if err != nil {
		out.Success = false
		out.Message = consts.ErrMsgMessageNotFound
		return out, err
	}

	// 检查消息状态是否允许重发
	if record.Status == consts.MessageStatusSuccess {
		err = gerror.New("消息已发送成功，无需重发")
		out.Success = false
		out.Message = err.Error()
		return out, err
	}

	// 检查重试次数
	if record.RetryCount >= consts.MaxRetryCount {
		err = gerror.Newf("消息已达到最大重试次数(%d)", consts.MaxRetryCount)
		out.Success = false
		out.Message = err.Error()
		return out, err
	}

	// 更新消息状态为待发送
	err = s.UpdateMessageStatus(ctx, in.MessageUuid, consts.MessageStatusPending, "")
	if err != nil {
		out.Success = false
		out.Message = "更新消息状态失败"
		return out, err
	}

	// 重新提交到消息队列（只发送 UUID 和类型，详细内容从数据库获取）
	err = service.Queue().PublishMessage(ctx, &dto.RocketMQMessage{
		MessageUuid: record.Uuid,
		MessageType: record.MessageType,
	})
	if err != nil {
		g.Log().Error(ctx, "重新提交消息到队列失败", err)
		out.Success = false
		out.Message = consts.ErrMsgRocketMQError
		return out, err
	}

	out.Success = true
	out.Message = "消息重发成功"
	g.Log().Info(ctx, "消息重发成功", "uuid", in.MessageUuid)

	return out, nil
}

// validateSendMessageInput 验证发送消息输入参数
func (s *sMessage) validateSendMessageInput(ctx context.Context, in *dto.SendMessageInput) error {
	// 验证接收人格式
	if in.MessageType == consts.MessageTypeEmail {
		if !isValidEmail(in.Recipient) {
			return gerror.New(consts.ErrMsgInvalidRecipient + ": 邮箱格式不正确")
		}
	} else if in.MessageType == consts.MessageTypeSMS {
		if !isValidPhone(in.Recipient) {
			return gerror.New(consts.ErrMsgInvalidRecipient + ": 手机号格式不正确")
		}
	}

	// 验证主题长度
	if len(in.Subject) > consts.MaxSubjectLength {
		return gerror.Newf("消息主题长度不能超过%d个字符", consts.MaxSubjectLength)
	}

	// 验证消息参数是否为有效的JSON
	if in.MessageArgs != "" {
		var tempMap map[string]interface{}
		if err := json.Unmarshal([]byte(in.MessageArgs), &tempMap); err != nil {
			return gerror.New("消息参数必须是有效的JSON格式")
		}
	}

	return nil
}

// renderTemplate 渲染模板
// 将模板中的变量占位符替换为实际值
func (s *sMessage) renderTemplate(ctx context.Context, template string, argsJSON string) (string, error) {
	if argsJSON == "" {
		return template, nil
	}

	// 解析参数
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", gerror.Wrap(err, "解析消息参数失败")
	}

	// 替换模板变量
	result := template
	for key, value := range args {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, gconv.String(value))
	}

	return result, nil
}

// GetTemplateById 根据ID获取模板
func (s *sMessage) GetTemplateById(ctx context.Context, templateId uint64) (*dto.MessageTemplateDTO, error) {
	var template *dto.MessageTemplateDTO
	err := g.DB().Model("message_template").
		Where("id", templateId).
		Where("deleted_at", 0).
		Scan(&template)
	if err != nil {
		return nil, gerror.Wrap(err, "查询模板失败")
	}
	if template == nil {
		return nil, gerror.New(consts.ErrMsgTemplateNotFound)
	}
	return template, nil
}

// GetMessageByUuid 根据UUID获取消息记录
func (s *sMessage) GetMessageByUuid(ctx context.Context, uuid string) (*dto.MessageRecordDTO, error) {
	var record *dto.MessageRecordDTO
	err := g.DB().Model("message_record").
		Where("uuid", uuid).
		Where("deleted_at", 0).
		Scan(&record)
	if err != nil {
		return nil, gerror.Wrap(err, "查询消息记录失败")
	}
	if record == nil {
		return nil, gerror.New(consts.ErrMsgMessageNotFound)
	}
	return record, nil
}

// CreateMessageRecord 创建消息记录
func (s *sMessage) CreateMessageRecord(ctx context.Context, record *dto.MessageRecordDTO) error {
	_, err := g.DB().Model("message_record").Data(record).Insert()
	if err != nil {
		return gerror.Wrap(err, "创建消息记录失败")
	}
	return nil
}

// UpdateMessageStatus 更新消息状态
func (s *sMessage) UpdateMessageStatus(ctx context.Context, uuid string, status int, errorMsg string) error {
	data := g.Map{
		"status":     status,
		"updated_at": gtime.Now().Unix(),
	}

	if status == consts.MessageStatusSuccess {
		data["send_time"] = gtime.Now().Unix()
	}

	if errorMsg != "" {
		if len(errorMsg) > consts.MaxErrorMessageLength {
			errorMsg = errorMsg[:consts.MaxErrorMessageLength]
		}
		data["error_message"] = errorMsg
	}

	// 如果是发送中或失败状态，增加重试次数
	if status == consts.MessageStatusSending || status == consts.MessageStatusFailed {
		_, err := g.DB().Model("message_record").
			Where("uuid", uuid).
			Increment("retry_count", 1)
		if err != nil {
			g.Log().Warning(ctx, "增加重试次数失败", err)
		}
	}

	_, err := g.DB().Model("message_record").
		Where("uuid", uuid).
		Data(data).
		Update()
	if err != nil {
		return gerror.Wrap(err, "更新消息状态失败")
	}

	return nil
}

// CreateSendLog 创建发送日志
func (s *sMessage) CreateSendLog(ctx context.Context, log *dto.MessageSendLogDTO) error {
	log.CreatedAt = gtime.Now().Unix()
	_, err := g.DB().Model("message_send_log").Data(log).Insert()
	if err != nil {
		return gerror.Wrap(err, "创建发送日志失败")
	}
	return nil
}

// isValidEmail 验证邮箱格式
func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// isValidPhone 验证手机号格式
func isValidPhone(phone string) bool {
	// 简单验证：11位数字
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}
