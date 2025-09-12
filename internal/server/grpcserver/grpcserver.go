// Package grpcserver  gprc server
package grpcserver

import (
	"context"

	"io"
	"net"

	"github.com/4aleksei/gokeeper/internal/common/datafile"
	"github.com/4aleksei/gokeeper/internal/common/logger"
	"github.com/4aleksei/gokeeper/internal/common/store"

	"github.com/4aleksei/gokeeper/internal/server/config"
	"github.com/4aleksei/gokeeper/internal/server/service"
	pb "github.com/4aleksei/gokeeper/pkg/api/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/4aleksei/gokeeper/internal/server/grpcserver/interceptor"
	"google.golang.org/grpc/status"
)

type (
	KeeperServiceService struct {
		pb.UnimplementedKeeperServiceServer
		srv  *grpc.Server
		l    *logger.ZapLogger
		serv *service.HandlerService
	}
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

func (s KeeperServiceService) UploadData(stream pb.KeeperService_UploadDataServer) error {
	var blockData *datafile.LongtermfileWrite
	var uuid string
	var encData *store.UserDataCrypt
	userID, ok := stream.Context().Value(interceptor.UserIdValue{}).(uint64)
	if !ok {
		return status.Errorf(codes.Internal, `%s`, "no USERID")
	}

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			if blockData != nil && encData != nil {
				var errAdd error
				blockData.Success()
				blockData.CloseWrite()
				uuid, errAdd = s.serv.AddDataStream(stream.Context(), encData)
				if errAdd != nil {
					return errAdd
				}
			}
			if uuid == "" {
				return status.Errorf(codes.Internal, `%s`, "error: null stream")
			}
			return stream.SendAndClose(&pb.ResponseAddData{
				Uuid: uuid,
			})
		}

		if err != nil {
			return err
		}

		if blockData == nil {
			data := &store.UserData{
				Id:       userID,
				TypeData: int(req.GetType()),
				MetaData: req.GetMetadata(),
			}
			var errAdd error
			blockData, encData, errAdd = s.serv.CreateDataStream(stream.Context(), data)

			if errAdd != nil {
				return errAdd
			}
		}
		defer blockData.CloseWrite()
		_, err = blockData.WriteData(req.GetData())
		if err != nil {
			return err
		}

	}
}

func (s KeeperServiceService) DownloadData(req *pb.DownloadRequest, stream pb.KeeperService_DownloadDataServer) error {

	userID, ok := stream.Context().Value(interceptor.UserIdValue{}).(uint64)
	if !ok {
		return status.Errorf(codes.Internal, `%s`, "no USERID")
	}

	data, blockData, err := s.serv.GetDataStream(stream.Context(), userID, req.GetUuid())
	if err != nil {
		return err
	}

	buffer := make([]byte, 4096) // Chunk size
	var sendMetaData bool
	var chunk *pb.DataChunk
	for {
		n, err := blockData.ReadData(buffer)
		if err == io.EOF {
			break // End of file
		}
		if err != nil {
			return status.Errorf(codes.Internal, "error reading : %v", err)
		}

		if !sendMetaData {
			sendMetaData = true
			chunk = &pb.DataChunk{
				Data:     buffer[:n],
				Metadata: data.MetaData,
				Type:     pb.TypeData(data.TypeData),
			}
		} else {
			chunk = &pb.DataChunk{
				Data: buffer[:n],
			}
		}
		if err := stream.Send(chunk); err != nil {
			return status.Errorf(codes.Internal, "error sending chunk: %v", err)
		}
	}

	return nil
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

func New(s *service.HandlerService, l *logger.ZapLogger, c *config.Config) (*KeeperServiceService, error) {
	listen, err := net.Listen("tcp", c.GrcpAddress)
	if err != nil {
		l.Logger.Debug("gRCP Listen Error: ", zap.Error(err))
		return nil, err
	}

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	auth := interceptor.NewAuthInterceptor(s)

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		logging.UnaryServerInterceptor(interceptor.InterceptorLogger(l.Logger), opts...),
		grpc.UnaryServerInterceptor(auth.UnaryAuthMiddleware),
	),
		grpc.ChainStreamInterceptor(
			//logging.StreamServerInterceptor(interceptor.InterceptorLogger(l.Logger), opts...),
			grpc.StreamServerInterceptor(auth.StreamAuthMiddleware),
		),
	)

	server := &KeeperServiceService{serv: s,
		srv: grpcServer,
		l:   l,
	}

	pb.RegisterKeeperServiceServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			l.Logger.Debug("gRCP Server Start Error: ", zap.Error(err))
		}
	}()
	return server, nil
}

func (s KeeperServiceService) StopServ() {
	s.srv.GracefulStop()
	s.srv.Stop()
}
