// Package app - Application client start
package app

import (
	"context"
	"errors"

	"github.com/4aleksei/gokeeper/internal/client/config"
	"github.com/4aleksei/gokeeper/internal/client/grpcclient"
	"github.com/4aleksei/gokeeper/internal/client/promt"
	"github.com/4aleksei/gokeeper/internal/client/promt/commands"
	"github.com/4aleksei/gokeeper/internal/client/promt/responses"
	"github.com/4aleksei/gokeeper/internal/client/service"
	"github.com/4aleksei/gokeeper/internal/common/logger"
	"github.com/4aleksei/gokeeper/internal/common/store"
)

var (
	ErrParamsNotEnough = errors.New("error parameters not enough")
)

func Run() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	l, err := logger.New(logger.Config{Level: cfg.Level})
	if err != nil {
		return err
	}

	client, err := grpcclient.New(cfg, l)
	if err != nil {
		return err
	}
	srv := service.New(client)

	commandLogin := func(ctx context.Context, s ...string) *responses.Respond {
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

	commandRegister := func(ctx context.Context, s ...string) *responses.Respond {
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

	commandData := func(ctx context.Context, s ...string) *responses.Respond {
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

	commandGetData := func(ctx context.Context, s ...string) *responses.Respond {
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

	pr := promt.New(
		promt.AddCommand(commands.New("Login", "Login name password ", commandLogin)),
		promt.AddCommand(commands.New("Register", "Register name password ", commandRegister)),
		promt.AddCommand(commands.New("AddData", "AddData type{'login','card','text','binary'} 'userdata' 'metadata'", commandData)),
		promt.AddCommand(commands.New("GetData", "GetData uuid", commandGetData)),
	)

	return pr.Loop(context.Background())
}
