// Package service - application logic
package service

import (
	"context"

	"github.com/4aleksei/gokeeper/internal/client/grpcclient"
	"github.com/4aleksei/gokeeper/internal/client/transaction"
)

type (
	HandleService struct {
		client *grpcclient.KeeperServiceService
	}
)

func New(c *grpcclient.KeeperServiceService) *HandleService {

	return &HandleService{
		client: c,
	}

}

func (s *HandleService) SendRegister(ctx context.Context, name string, pass string) (string, error) {
	req := &transaction.Request{
		Command: transaction.UserRegister{User: transaction.User{Name: name, Password: pass}},
	}
	resp, err := s.client.SendSingleCommand(ctx, req)
	if err != nil {
		return "", err
	}
	str, ok := resp.Resp.(transaction.TokenUser)
	if !ok {
		return "", transaction.ErrBadTypeResponse
	}
	return str.Token, nil
}

func (s *HandleService) SendLogin(ctx context.Context, name string, pass string) (string, error) {
	req := &transaction.Request{
		Command: transaction.UserLogin{User: transaction.User{Name: name, Password: pass}},
	}
	resp, err := s.client.SendSingleCommand(ctx, req)
	if err != nil {
		return "", err
	}
	str, ok := resp.Resp.(transaction.TokenUser)
	if !ok {
		return "", transaction.ErrBadTypeResponse
	}
	return str.Token, nil
}

func (s *HandleService) SendData(ctx context.Context, token string, typdata int, data string, metadata string) (string, error) {
	req := &transaction.Request{
		Command: transaction.UserData{Token: transaction.TokenUser{Token: token}, TypeData: typdata, Data: data, MetaData: metadata},
	}
	resp, err := s.client.SendSingleCommand(ctx, req)
	if err != nil {
		return "", err
	}
	str, ok := resp.Resp.(transaction.UUIDData)
	if !ok {
		return "", transaction.ErrBadTypeResponse
	}
	return str.UUID, nil
}

func (s *HandleService) GetData(ctx context.Context, token string, uuid string) (*transaction.UserData, error) {
	req := &transaction.Request{
		Command: transaction.GetUserData{Token: transaction.TokenUser{Token: token}, UUID: transaction.UUIDData{UUID: uuid}},
	}
	resp, err := s.client.SendSingleCommand(ctx, req)
	if err != nil {
		return nil, err
	}
	str, ok := resp.Resp.(transaction.UserData)
	if !ok {
		return nil, transaction.ErrBadTypeResponse
	}
	return &str, nil
}
