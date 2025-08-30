package grpcserver

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/4aleksei/gokeeper/internal/common/logger"
	"github.com/4aleksei/gokeeper/internal/common/store"
	"github.com/4aleksei/gokeeper/internal/common/store/cache"
	"github.com/4aleksei/gokeeper/internal/server/config"
	"github.com/4aleksei/gokeeper/internal/server/service"
	pb "github.com/4aleksei/gokeeper/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func initNew() {
	lis = bufconn.Listen(bufSize)
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestServerSingle(t *testing.T) {
	initNew()
	l, _ := logger.New(logger.Config{Level: "debug"})
	grpcServer := grpc.NewServer()
	cfg, _ := config.New()
	st := service.New(cache.New(l.Logger), l.Logger, cfg)

	pb.RegisterKeeperServiceServer(grpcServer, KeeperServiceService{serv: st,
		srv: grpcServer,
		l:   l,
	})

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	defer grpcServer.Stop() // Ensure server is stopped after test

	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewKeeperServiceClient(conn)

	tests := []struct {
		name    string
		oper    string
		user    store.User
		wantId  uint64
		errcode codes.Code
	}{
		{name: "Test N1 RegisterUser", oper: "RegisterUser", user: store.User{Name: "aa", HashPass: "bbb"}, wantId: 1},

		{name: "Test N2 LoginUser", oper: "LoginUser", user: store.User{Name: "aa", HashPass: "bbb"}, wantId: 1},

		{name: "Test N3 Error LoginUser", oper: "LoginUser", user: store.User{Name: "cc", HashPass: "bbb"}, wantId: 1, errcode: codes.NotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var val *pb.LoginResponse

			var err error
			switch tt.oper {
			case "LoginUser":
				val, err = client.LoginUser(context.Background(), &pb.LoginRequest{Name: tt.user.Name, Password: tt.user.HashPass})
			case "RegisterUser":
				val, err = client.RegisterUser(context.Background(), &pb.LoginRequest{Name: tt.user.Name, Password: tt.user.HashPass})
			default:
				log.Fatal("error operation")
			}
			if err != nil {
				if e, ok := status.FromError(err); ok {
					switch e.Code() {
					case codes.NotFound, codes.DeadlineExceeded:
						log.Println(e.Message())
					default:
						log.Println(e.Code(), e.Message())
					}
					assert.Equal(t, tt.errcode, e.Code())
				} else {
					log.Printf("не получилось распарсить ошибку %v", err)
					require.NoError(t, err)
				}
			} else {
				require.NoError(t, err)
				//resp.getValue(val.Value)
				id, errCh := st.CheckToken(context.Background(), val.GetToken())
				require.NoError(t, errCh)

				assert.Equal(t, tt.wantId, id)
			}
		})
	}
}
