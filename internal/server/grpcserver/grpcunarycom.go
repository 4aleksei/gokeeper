// Package grpcserver  gprc server
package grpcserver

import (
	"context"

	"github.com/4aleksei/gokeeper/internal/common/store"

	pb "github.com/4aleksei/gokeeper/pkg/api/proto"
	"google.golang.org/grpc/codes"

	"github.com/4aleksei/gokeeper/internal/server/grpcserver/interceptor"
	"google.golang.org/grpc/status"
)

func (s KeeperServiceService) LoginUser(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	var response pb.LoginResponse
	name := in.GetName()
	pass := in.GetPassword()
	user, err := s.serv.LoginUser(ctx, name, pass)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, `%s`, err.Error())
	}
	token, err := s.serv.BuildToken(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `%s`, err.Error())
	}

	response.Token = token
	return &response, nil
}

func (s KeeperServiceService) RegisterUser(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	var response pb.LoginResponse

	name := in.GetName()
	pass := in.GetPassword()
	user, err := s.serv.RegisterUser(ctx, name, pass)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `%s`, err.Error())
	}
	token, err := s.serv.BuildToken(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `%s`, err.Error())
	}

	response.Token = token
	return &response, nil
}

func (s KeeperServiceService) AddData(ctx context.Context, in *pb.UserData) (*pb.ResponseAddData, error) {
	var response pb.ResponseAddData
	userID, ok := ctx.Value(interceptor.UserIdValue{}).(uint64)
	if !ok {
		return nil, status.Errorf(codes.Internal, `%s`, "no USERID")
	}

	data := &store.UserData{
		Id:       userID,
		TypeData: int(in.GetType()),
		UserData: in.GetData(),
		MetaData: in.GetMetadata(),
	}

	uuid, err := s.serv.AddData(ctx, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `%s`, err.Error())
	}
	response.Uuid = uuid
	return &response, nil
}

func (s KeeperServiceService) GetData(ctx context.Context, in *pb.DownloadRequest) (*pb.UserData, error) {
	var response pb.UserData
	userID, ok := ctx.Value(interceptor.UserIdValue{}).(uint64)
	if !ok {
		return nil, status.Errorf(codes.Internal, `%s`, "no USERID")
	}

	data, err := s.serv.GetData(ctx, userID, in.GetUuid())
	if err != nil {
		return nil, status.Errorf(codes.Internal, `%v`, err)
	}
	response.Data = data.UserData
	response.Metadata = data.MetaData
	response.Type = pb.TypeData(data.TypeData)
	return &response, nil
}
