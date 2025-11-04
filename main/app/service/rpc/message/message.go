package message

import (
	"context"
	"errors"
	v1 "ttpos-bmp/app/ttpos-message/api/message"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/pkg/database"

	"google.golang.org/grpc"
)

type IMessageSrv interface {
	SendMessage(ctx context.Context, sendMessageReq *v1.SendMessageReq) (*v1.SendMessageResp, error)
}

type messageSrv struct {
}

func NewIMessageSrv(dbm *database.DBManager) IMessageSrv {
	return &messageSrv{}
}

func NewMessageClient() (v1.MessageClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.MessageServiceName)
	if err != nil {
		return nil, nil, err
	}
	return v1.NewMessageClient(conn), conn, nil
}

func (s *messageSrv) SendMessage(ctx context.Context, sendMessageReq *v1.SendMessageReq) (*v1.SendMessageResp, error) {
	client, conn, err := NewMessageClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	result, err := client.SendMessage(ctx, sendMessageReq)
	if err != nil {
		return nil, err
	}
	if result.Success {
		return result, nil
	}
	return nil, errors.New(result.Message)
}
