package mq

type AsyncSellingMsg struct {
	RecordId int64  `json:"record_id"` //记录ID
	MsgType  MsgTyp `json:"msg_type"`  //消息类型
}
