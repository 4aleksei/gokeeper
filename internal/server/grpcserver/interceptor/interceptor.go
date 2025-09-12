package interceptor

import (
	"context"
	"fmt"

	"github.com/4aleksei/gokeeper/internal/server/service"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type (
	authInterceptor struct {
		serv *service.HandlerService
	}

	UserIdValue struct {
	}
)

func NewAuthInterceptor(s *service.HandlerService) *authInterceptor {
	return &authInterceptor{serv: s}
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
	ctx = context.WithValue(ctx, UserIdValue{}, userID)

	// call our handler
	return handler(ctx, req)
}

func (a *authInterceptor) StreamAuthMiddleware(req any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// extract token from authorization header
	token := md["authorization"]
	if len(token) == 0 {
		return status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	// validate token and retrieve the userID
	userID, err := a.serv.CheckToken(ctx, token[0])
	if err != nil {
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	serverStream := &grpcmiddleware.WrappedServerStream{
		ServerStream:   stream,
		WrappedContext: context.WithValue(ctx, UserIdValue{}, userID),
	}
	// add our user ID to the context, so we can use it in our RPC handler
	//ctx = context.WithValue(ctx, UserIdValue{}, userID)
	//stream.Context().WithValue()

	return handler(req, serverStream)
}
