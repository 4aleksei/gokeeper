// Package grpcserver  gprc server
package grpcserver

import (
	"net"

	"github.com/4aleksei/gokeeper/internal/common/logger"

	"github.com/4aleksei/gokeeper/internal/server/config"
	"github.com/4aleksei/gokeeper/internal/server/service"
	pb "github.com/4aleksei/gokeeper/pkg/api/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/4aleksei/gokeeper/internal/server/grpcserver/interceptor"
)

type (
	KeeperServiceService struct {
		pb.UnimplementedKeeperServiceServer
		srv  *grpc.Server
		l    *logger.ZapLogger
		serv *service.HandlerService
	}
)

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
