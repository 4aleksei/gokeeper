// Package grpcserver  gprc server
package grpcserver

import (
	"io"

	"github.com/4aleksei/gokeeper/internal/common/datafile"
	"github.com/4aleksei/gokeeper/internal/common/store"

	pb "github.com/4aleksei/gokeeper/pkg/api/proto"
	"google.golang.org/grpc/codes"

	"github.com/4aleksei/gokeeper/internal/server/grpcserver/interceptor"
	"google.golang.org/grpc/status"
)

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
