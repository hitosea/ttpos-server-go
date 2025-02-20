package service

import (
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/pkg/database"
)

// ICallSrv 定义沽清服务接口
type ICallSrv interface {
	Unprocessed() resp.UnprocessedResp
	CallList(companyUuid uint64, soldOutReq req.CallListReq) (resp.CallList, error)                           // 呼叫列表
	PrintExceptionList(companyUuid uint64, soldOutReq req.PrintExceptionReq) (resp.PrintExceptionList, error) // 打印异常列表
}

// callSrv 沽清服务结构体
type callSrv struct {
	dbm *database.DBManager // 数据库管理

}

// NewCallSrv 创建新的收银产品类别服务
func NewCallSrv(dbm *database.DBManager) ICallSrv {
	return NewCallSrvImpl(dbm)
}

// NewCallSrvImpl 创建新的收银服务实现
func NewCallSrvImpl(dbm *database.DBManager) ICallSrv {
	return &callSrv{
		dbm: dbm,
	}
}

// CallList 获取呼叫列表
func (s *callSrv) CallList(companyUuid uint64, soldOutReq req.CallListReq) (resp.CallList, error) {
	return resp.CallList{}, nil
}

// PrintExceptionList 获取打印异常列表
func (s *callSrv) PrintExceptionList(companyUuid uint64, soldOutReq req.PrintExceptionReq) (resp.PrintExceptionList, error) {
	return resp.PrintExceptionList{}, nil
}

// Unprocessed 获取未处理消息数量
func (s *callSrv) Unprocessed() resp.UnprocessedResp {
	return resp.UnprocessedResp{}
}
