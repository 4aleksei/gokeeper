// Package app - Application client start
package app

import (
	"context"

	"github.com/4aleksei/gokeeper/internal/client/config"
	"github.com/4aleksei/gokeeper/internal/client/grpcclient"
	"github.com/4aleksei/gokeeper/internal/client/promt"
	"github.com/4aleksei/gokeeper/internal/client/promt/command"
	"github.com/4aleksei/gokeeper/internal/client/promt/commands"
	"github.com/4aleksei/gokeeper/internal/client/service"
	"github.com/4aleksei/gokeeper/internal/common/logger"
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
	srvV := service.New(client)

	pr := promt.New(
		promt.AddCommand(command.New(srvV, "Login", "Login name password ", commands.CommandLogin)),
		promt.AddCommand(command.New(srvV, "Register", "Register name password ", commands.CommandRegister)),
		promt.AddCommand(command.New(srvV, "AddData", "AddData type{'login','card'} 'userdata' 'metadata'", commands.CommandData)),
		promt.AddCommand(command.New(srvV, "GetData", "GetData uuid", commands.CommandGetData)),
		promt.AddCommand(command.New(srvV, "UploadData", "UploadData type{'text','binary'} 'metadata' 'filename of data'", commands.CommandUploadData)),
	)

	return pr.Loop(context.Background())
}
