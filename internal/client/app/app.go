// Package app - Application client start
package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/4aleksei/gokeeper/internal/client/config"
	"github.com/4aleksei/gokeeper/internal/client/grpcclient"
	"github.com/4aleksei/gokeeper/internal/client/prompt"
	"github.com/4aleksei/gokeeper/internal/client/prompt/command"
	"github.com/4aleksei/gokeeper/internal/client/prompt/commands"
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

	pr := prompt.New(
		prompt.AddCommand(command.New(srvV, "Login", "Login name password ", commands.CommandLogin)),
		prompt.AddCommand(command.New(srvV, "Register", "Register name password ", commands.CommandRegister)),
		prompt.AddCommand(command.New(srvV, "AddData", "AddData type{'login','card'} 'userdata' 'metadata'", commands.CommandData)),
		prompt.AddCommand(command.New(srvV, "GetData", "GetData uuid", commands.CommandGetData)),
		prompt.AddCommand(command.New(srvV, "UploadData", "UploadData type{'text','binary'} 'metadata' 'filename of data'", commands.CommandUploadData)),
		prompt.AddCommand(command.New(srvV, "DownloadData", "DownloadData uuid", commands.CommandDownloadData)),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go func() {
		err = pr.Loop(ctx)
	}()

	<-ctx.Done()

	stop()
	return err
}
