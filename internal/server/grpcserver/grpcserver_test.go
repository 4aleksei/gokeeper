package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/4aleksei/gokeeper/internal/common/cryptocerts"
	"github.com/4aleksei/gokeeper/internal/common/datacrypto"
	"github.com/4aleksei/gokeeper/internal/common/logger"
	"github.com/4aleksei/gokeeper/internal/common/store"
	"github.com/4aleksei/gokeeper/internal/common/store/cache"
	"github.com/4aleksei/gokeeper/internal/server/config"
	"github.com/4aleksei/gokeeper/internal/server/service"
	pb "github.com/4aleksei/gokeeper/pkg/api/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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

	pr, pub, _ := cryptocerts.GenerateKey()
	crypto := datacrypto.New(pr, pub)

	st := service.New(cache.New(l.Logger), crypto, l.Logger, cfg)

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

func TestInsertData(t *testing.T) {
	initNew()
	l, _ := logger.New(logger.Config{Level: "debug"})
	cfg := &config.Config{Key: "Secret"}
	pr, pub, _ := cryptocerts.GenerateKey()
	crypto := datacrypto.New(pr, pub)

	st := service.New(cache.New(l.Logger), crypto, l.Logger, cfg)

	interceptor := NewAuthInterceptor(st)
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(

		grpc.UnaryServerInterceptor(interceptor.UnaryAuthMiddleware),
		//UnaryServerBlock(optsMy...),
	),
		grpc.ChainStreamInterceptor(

		//StreamServerBlock(optsMy...),
		))
	//cfg, _ := config.New()

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
	var login *pb.LoginResponse
	client := pb.NewKeeperServiceClient(conn)

	login, err = client.RegisterUser(context.Background(), &pb.LoginRequest{Name: "user1", Password: "abcd"})
	require.NoError(t, err)

	type dataST struct {
		typeData int
		data     string
		metadata string
	}
	dataTest1 := dataST{typeData: 0, data: "abc", metadata: "meta_abc"}

	tests := []struct {
		name    string
		oper    string
		data    dataST
		errcode codes.Code
	}{
		{name: "Test N1 AddData", oper: "AddData", data: dataTest1},
	}

	md := metadata.New(map[string]string{"authorization": login.GetToken()})
	ctxReq := metadata.NewOutgoingContext(context.Background(), md)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var val *pb.ResponseAddData

			var err error
			switch tt.oper {
			case "AddData":
				val, err = client.AddData(ctxReq, &pb.UserData{Type: pb.TypeData(tt.data.typeData), Data: tt.data.data, Metadata: tt.data.metadata})

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
				fmt.Println(val.GetUuid())

				valData, err := client.GetData(ctxReq, &pb.DownloadRequest{Uuid: val.GetUuid()})

				require.NoError(t, err)
				resp := dataST{
					typeData: int(valData.GetType()),
					data:     valData.GetData(),
					metadata: valData.GetMetadata(),
				}
				assert.Equal(t, tt.data, resp)
				//resp.getValue(val.Value)
				//id, errCh := st.CheckToken(context.Background(), val.GetToken())
				//require.NoError(t, errCh)
				//assert.Equal(t, tt.wantId, id)
			}
		})
	}

}
