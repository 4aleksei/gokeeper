// Package grpcclient - gRPC realization
package grpcclient

import (
	"context"
	"crypto/x509"
	"io"
	"net"
	"os"
	"time"

	"github.com/4aleksei/gokeeper/internal/client/config"
	"github.com/4aleksei/gokeeper/internal/client/transaction"
	"github.com/4aleksei/gokeeper/internal/common/logger"
	pb "github.com/4aleksei/gokeeper/pkg/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/metadata"
)

type (
	KeeperServiceService struct {
		client *agentClient
		l      *logger.ZapLogger
	}

	agentClient struct {
		client     pb.KeeperServiceClient
		connection *grpc.ClientConn
		localAddr  string
	}
)

func loadCert(cfg *config.Config) (grpc.DialOption, error) {

	if cfg.CertKeyFile != "" {
		caCert, err := os.ReadFile(cfg.CertKeyFile)
		if err != nil {
			return grpc.WithTransportCredentials(insecure.NewCredentials()), err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		// Create TLS configuration
		tlsCredentials := credentials.NewClientTLSFromCert(caCertPool, "") // "localhost" must match server certificate's CN or SAN

		return grpc.WithTransportCredentials(tlsCredentials), nil
	}

	return grpc.WithTransportCredentials(insecure.NewCredentials()), nil
}

func newClient(cfg *config.Config) (*agentClient, error) {
	agclient := &agentClient{}
	myDialer := net.Dialer{Timeout: 30 * time.Second,
		KeepAlive: 30 * time.Second}

	dialOpt, err := loadCert(cfg)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.NewClient(cfg.Address, dialOpt,
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			conn, err := myDialer.DialContext(ctx, "tcp", addr)
			if err == nil {
				agclient.localAddr = conn.LocalAddr().(*net.TCPAddr).IP.String()
			}
			return conn, err
		}))

	if err != nil {
		return nil, err
	}
	c := pb.NewKeeperServiceClient(conn)
	agclient.client = c
	agclient.connection = conn
	return agclient, nil
}

func New(cfg *config.Config, l *logger.ZapLogger) (*KeeperServiceService, error) {
	c, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return &KeeperServiceService{
		client: c,
		l:      l,
	}, nil
}

func (c *KeeperServiceService) SendSingleCommand(ctx context.Context, req *transaction.Request) (*transaction.Response, error) {
	resp, err := sendTransaction(ctx, c.client, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *KeeperServiceService) SendStreamCommand(ctx context.Context, req *transaction.Request) (*transaction.Response, error) {
	resp, err := sendStream(ctx, c.client, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func sendStream(ctx context.Context, client *agentClient, req *transaction.Request) (*transaction.Response, error) {
	md := metadata.New(map[string]string{"X-Real-IP": client.localAddr})
	ctxReq := metadata.NewOutgoingContext(ctx, md)

	switch v := req.Command.(type) {
	case transaction.StreamData:
		md := metadata.New(map[string]string{"authorization": v.Token.Token})
		ctxReqMd := metadata.NewOutgoingContext(ctxReq, md)
		stream, err := client.client.UploadData(ctxReqMd)
		if err != nil {
			return nil, err
		}
		var fsend bool
		for res := range v.Output {
			select {
			case <-ctxReq.Done():
				return nil, ctxReq.Err()

			default:

				if !fsend {
					err = stream.Send(&pb.DataChunk{Data: res, Type: pb.TypeData(v.TypeData), Metadata: v.MetaData})
					fsend = true
				} else {
					err = stream.Send(&pb.DataChunk{Data: res})
				}
				if err != nil {
					return nil, err
				}
			}
		}

		resp, err := stream.CloseAndRecv()
		if err != nil {
			return nil, err
		}

		return &transaction.Response{Resp: transaction.UUIDData{UUID: resp.GetUuid()}}, nil

	case transaction.GetStreamData:

		md := metadata.New(map[string]string{"authorization": v.Token.Token})
		ctxReqMd := metadata.NewOutgoingContext(ctxReq, md)
		stream, err := client.client.DownloadData(ctxReqMd, &pb.DownloadRequest{Uuid: v.UUID.UUID})
		if err != nil {
			return nil, err
		}
		var tx transaction.UserData
		var firstP bool
		defer close(v.Input)
		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				break // End of stream
			}
			if err != nil {
				return nil, err
			}
			v.Input <- chunk.GetData()

			if !firstP {
				tx.MetaData = chunk.Metadata
				tx.TypeData = int(chunk.GetType())
				firstP = true
			}
		}

		return &transaction.Response{Resp: tx}, nil
	}

	return nil, transaction.ErrBadTypeCommand
}

func sendTransaction(ctx context.Context, client *agentClient, req *transaction.Request) (*transaction.Response, error) {

	md := metadata.New(map[string]string{"X-Real-IP": client.localAddr})
	ctxReq := metadata.NewOutgoingContext(ctx, md)

	switch v := req.Command.(type) {
	case transaction.UserLogin:
		resp, err := client.client.LoginUser(ctxReq, &pb.LoginRequest{Name: v.User.Name, Password: v.User.Password})
		if err != nil {
			return nil, err
		}
		return &transaction.Response{Resp: transaction.TokenUser{Token: resp.GetToken()}}, nil

	case transaction.UserRegister:

		resp, err := client.client.RegisterUser(ctxReq, &pb.LoginRequest{Name: v.User.Name, Password: v.User.Password})
		if err != nil {
			return nil, err
		}

		return &transaction.Response{Resp: transaction.TokenUser{Token: resp.GetToken()}}, nil

	case transaction.UserData:
		md := metadata.New(map[string]string{"authorization": v.Token.Token})
		ctxReqMd := metadata.NewOutgoingContext(ctxReq, md)
		resp, err := client.client.AddData(ctxReqMd, &pb.UserData{Data: v.Data, Metadata: v.MetaData, Type: pb.TypeData(v.TypeData)})
		if err != nil {
			return nil, err
		}
		return &transaction.Response{Resp: transaction.UUIDData{UUID: resp.GetUuid()}}, nil

	case transaction.GetUserData:
		md := metadata.New(map[string]string{"authorization": v.Token.Token})
		ctxReqMd := metadata.NewOutgoingContext(ctxReq, md)
		resp, err := client.client.GetData(ctxReqMd, &pb.DownloadRequest{Uuid: v.UUID.UUID})
		if err != nil {
			return nil, err
		}
		return &transaction.Response{Resp: transaction.UserData{Data: resp.GetData(), MetaData: resp.Metadata, TypeData: int(resp.GetType())}}, nil

	}

	return nil, transaction.ErrBadTypeCommand
}
