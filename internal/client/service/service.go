// Package service - application logic
package service

import (
	"context"
	"io"
	"os"

	"github.com/4aleksei/gokeeper/internal/client/grpcclient"
	"github.com/4aleksei/gokeeper/internal/client/transaction"
	"github.com/google/uuid"
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

func openReadFile(filename string) (chan []byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	ch := make(chan []byte)
	var errR error
	go func() {
		buffer := make([]byte, 4096)
		defer close(ch)
		defer file.Close()
		for {
			var n int
			n, errR = file.Read(buffer)
			if errR == io.EOF {
				break
			}
			if errR != nil {
				break
			}
			ch <- buffer[:n]
		}
	}()
	return ch, nil
}

func (s *HandleService) UploadData(ctx context.Context, token string, typdata int, metadata string, filename string) (string, error) {
	ch, err := openReadFile(filename)
	if err != nil {
		return "", err
	}
	req := &transaction.Request{
		Command: transaction.StreamData{Token: transaction.TokenUser{Token: token}, TypeData: typdata, MetaData: metadata, Output: ch},
	}

	resp, err := s.client.SendStreamCommand(ctx, req)
	if err != nil {
		return "", err
	}
	str, ok := resp.Resp.(transaction.UUIDData)
	if !ok {
		return "", transaction.ErrBadTypeResponse
	}
	return str.UUID, nil
}
func (s *HandleService) genFileName() string {
	name := uuid.New().String() + ".data"
	return name
}

func openWriteFile(ctx context.Context, filename string) (chan []byte, error) {

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	ch := make(chan []byte)
	var errW error
	go func() {

		defer file.Close()
		for res := range ch {

			select {
			case <-ctx.Done():
				return

			default:
				_, errW = file.Write(res)
				if errW != nil {
					return
				}
			}
		}
	}()
	return ch, nil
}

func (s *HandleService) DownloadData(ctx context.Context, token string, uuid string) (*transaction.UserData, error) {
	filename := s.genFileName()
	ch, err := openWriteFile(ctx, filename)
	if err != nil {
		return nil, err
	}

	req := &transaction.Request{
		Command: transaction.GetStreamData{Token: transaction.TokenUser{Token: token}, UUID: transaction.UUIDData{UUID: uuid}, Input: ch},
	}
	resp, err := s.client.SendStreamCommand(ctx, req)
	if err != nil {
		return nil, err
	}
	str, ok := resp.Resp.(transaction.UserData)
	if !ok {
		return nil, transaction.ErrBadTypeResponse
	}
	str.Data = filename
	return &str, nil

}
