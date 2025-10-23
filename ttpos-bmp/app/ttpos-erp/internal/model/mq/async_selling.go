package mq

type AsyncSellingMsg struct {
	RecordId         int64  `json:"record_id,omitempty"`           //记录ID
	MsgType          MsgTyp `json:"msg_type"`                      //消息类型
	PosOpenEntryName string `json:"pos_open_entry_name,omitempty"` //开单名称
}
