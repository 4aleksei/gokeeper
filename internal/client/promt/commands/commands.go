// Package commands - commands realization
package commands

import (
	"context"
	"errors"

	"github.com/4aleksei/gokeeper/internal/client/promt/responses"
	"github.com/4aleksei/gokeeper/internal/client/service"
	"github.com/4aleksei/gokeeper/internal/common/store"
)

var (
	ErrParamsNotEnough = errors.New("error parameters not enough")
)

func CommandLogin(ctx context.Context, srv *service.HandleService, s ...string) *responses.Respond {
	if len(s) < 3 {
		return responses.New(
			responses.AddError(ErrParamsNotEnough),
		)
	}

	token, err := srv.SendLogin(ctx, s[1], s[2])
	if err != nil {
		return responses.New(
			responses.AddError(err),
		)
	}
	return responses.New(
		responses.AddUserName(s[1]),
		responses.AddToken(token),
	)
}

func CommandRegister(ctx context.Context, srv *service.HandleService, s ...string) *responses.Respond {
	if len(s) < 3 {
		return responses.New(
			responses.AddError(ErrParamsNotEnough),
		)
	}
	token, err := srv.SendRegister(ctx, s[1], s[2])
	if err != nil {
		return responses.New(
			responses.AddError(err),
		)
	}
	return responses.New(
		responses.AddUserName(s[1]),
		responses.AddToken(token),
	)
}

func CommandData(ctx context.Context, srv *service.HandleService, s ...string) *responses.Respond {
	if len(s) < 4 {
		return responses.New(
			responses.AddError(ErrParamsNotEnough),
		)
	}
	t, err := store.GetType(s[1])
	if err != nil {
		return responses.New(
			responses.AddError(err),
		)
	}

	uuid, err := srv.SendData(ctx, s[0], t, s[2], s[3])
	if err != nil {
		return responses.New(
			responses.AddError(err),
		)
	}
	return responses.New(
		responses.AddUUID(uuid),
	)
}

func CommandGetData(ctx context.Context, srv *service.HandleService, s ...string) *responses.Respond {
	if len(s) < 2 {
		return responses.New(
			responses.AddError(ErrParamsNotEnough),
		)
	}
	data, err := srv.GetData(ctx, s[0], s[1])
	if err != nil {
		return responses.New(
			responses.AddError(err),
		)
	}
	return responses.New(
		responses.AddData(data.TypeData, data.Data, data.MetaData),
	)
}

func CommandUploadData(ctx context.Context, srv *service.HandleService, s ...string) *responses.Respond {
	if len(s) < 4 {
		return responses.New(
			responses.AddError(ErrParamsNotEnough),
		)
	}
	t, err := store.GetType(s[1])
	if err != nil {
		return responses.New(
			responses.AddError(err),
		)
	}

	uuid, err := srv.UploadData(ctx, s[0], t, s[2], s[3])
	if err != nil {
		return responses.New(
			responses.AddError(err),
		)
	}
	return responses.New(
		responses.AddUUID(uuid),
	)
}

func CommandDownloadData(ctx context.Context, srv *service.HandleService, s ...string) *responses.Respond {
	if len(s) < 2 {
		return responses.New(
			responses.AddError(ErrParamsNotEnough),
		)
	}

	data, err := srv.DownloadData(ctx, s[0], s[1])
	if err != nil {
		return responses.New(
			responses.AddError(err),
		)
	}
	return responses.New(
		responses.AddData(data.TypeData, data.Data, data.MetaData),
	)
}
