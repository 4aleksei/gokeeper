// Package grpcserver  gprc server
package grpcserver

import (
	"context"
	"fmt"
	"net"

	/*
		"crypto/tls"

			"net/netip"

		"google.golang.org/grpc/credentials"*/
	"github.com/4aleksei/gokeeper/internal/common/logger"
	"github.com/4aleksei/gokeeper/internal/server/config"
	"github.com/4aleksei/gokeeper/internal/server/service"
	pb "github.com/4aleksei/gokeeper/pkg/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	/*"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/4aleksei/metricscum/internal/common/utils"
	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/service"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"*//*

		"google.golang.org/grpc/codes"
		_ "google.golang.org/grpc/encoding/gzip"
		"google.golang.org/grpc/metadata"
		"google.golang.org/grpc/peer"
		"google.golang.org/grpc/status"*/)

type (
	KeeperServiceService struct {
		pb.UnimplementedKeeperServiceServer
		//store *service.HandlerStore
		srv  *grpc.Server
		l    *logger.ZapLogger
		serv *service.HandlerService
	}

	authInterceptor struct {
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

func (s KeeperServiceService) AddData(ctx context.Context, in *pb.UserData) (*pb.ResponseAddData, error) {
	var response pb.ResponseAddData
	userID, ok := ctx.Value("user_id").(uint64)
	if !ok {
		return nil, status.Errorf(codes.Internal, `%s`, "no USERID")
	}

	uuid, err := s.serv.AddData(ctx, userID, int(in.GetType()), in.GetData(), in.GetMetadata())
	if err != nil {
		return nil, status.Errorf(codes.Internal, `%s`, err.Error())
	}
	response.Uuid = uuid
	return &response, nil
}

func (s KeeperServiceService) GetData(ctx context.Context, in *pb.DownloadRequest) (*pb.UserData, error) {
	var response pb.UserData
	userID, ok := ctx.Value("user_id").(uint64)
	if !ok {
		return nil, status.Errorf(codes.Internal, `%s`, "no USERID")
	}

	data, err := s.serv.GetData(ctx, userID, in.GetUuid())
	if err != nil {
		return nil, status.Errorf(codes.Internal, `%v`, err)
	}
	response.Data = data.UserData
	response.Metadata = data.MetaData
	response.Type = pb.UserData_Type(data.TypeData)
	return &response, nil
}

func NewAuthInterceptor(s *service.HandlerService) *authInterceptor {
	return &authInterceptor{serv: s}
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

	interceptor := NewAuthInterceptor(s)

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		logging.UnaryServerInterceptor(InterceptorLogger(l.Logger), opts...),
		grpc.UnaryServerInterceptor(interceptor.UnaryAuthMiddleware),
		//UnaryServerBlock(optsMy...),
	),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(InterceptorLogger(l.Logger), opts...),
			//StreamServerBlock(optsMy...),
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

func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
		}
	})
}

func isMethodLoginRegister(info *grpc.UnaryServerInfo) bool {
	switch info.FullMethod {
	case "/grpcgokeeper.KeeperService/LoginUser":
		return true
	case "/grpcgokeeper.KeeperService/RegisterUser":
		return true
	default:
		return false
	}
}

func (a *authInterceptor) UnaryAuthMiddleware(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// get metadata object
	fmt.Printf("Intercepting call to: %s\n", info.FullMethod)
	if isMethodLoginRegister(info) {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// extract token from authorization header
	token := md["authorization"]
	if len(token) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	// validate token and retrieve the userID
	userID, err := a.serv.CheckToken(ctx, token[0])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	// add our user ID to the context, so we can use it in our RPC handler
	ctx = context.WithValue(ctx, "user_id", userID)

	// call our handler
	return handler(ctx, req)
}
