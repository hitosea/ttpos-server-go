package repository

import (
	"websocket/model"
	"websocket/pkg/database"
)

type WebSocketMsgRepository struct {
	dbm *database.DBManager
}

func NewWebSocketMsgRepository(dbm *database.DBManager) *WebSocketMsgRepository {
	return &WebSocketMsgRepository{dbm: dbm}
}

func (r *WebSocketMsgRepository) Update(id uint, vars map[string]interface{}) error {
	return r.dbm.GetDB(0).Model(&model.WebSocketMsg{}).Where("id = ?", id).Updates(vars).Error
}

func (r *WebSocketMsgRepository) Create(webSocketMsg model.WebSocketMsg) (uint, error) {
	result := r.dbm.GetDB(0).Create(&webSocketMsg)
	return webSocketMsg.ID, result.Error
}

func (r *WebSocketMsgRepository) DeleteByTypeAndCompanyId(msgType string, companyId uint) error {
	return r.dbm.GetDB(0).Where("type = ? AND company_uuid = ?", msgType, companyId).Delete(&model.WebSocketMsg{}).Error
}

func (r *WebSocketMsgRepository) DeleteByTypeAndId(id uint) error {
	return r.dbm.GetDB(0).Where("id = ?", id).Delete(&model.WebSocketMsg{}).Error
}
