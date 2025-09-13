// Package app - Application start and stop
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/4aleksei/gokeeper/internal/common/logger"
	"github.com/4aleksei/gokeeper/internal/server/config"
	"github.com/4aleksei/gokeeper/internal/server/grpcserver"
	"github.com/4aleksei/gokeeper/internal/server/resources"
	"github.com/4aleksei/gokeeper/internal/server/service"
	"go.uber.org/zap"
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

	storageRes, errC := resources.New(cfg, l)
	if errC != nil {
		l.Logger.Error("Error create resources :", zap.Error(errC))
		return errC
	}

	gService := service.New(storageRes.Store, storageRes.Enc, l.Logger, cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	grpcServ, errG := grpcserver.New(gService, l, cfg)
	if errG != nil {
		l.Logger.Error("Error server grpc construct:", zap.Error(errG))
		return errG
	}

	<-ctx.Done()

	stop()
	fmt.Println("signal received")

	//sigs := make(chan os.Signal, 1)
	//signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	//sig := <-sigs

	l.Logger.Info("Server is shutting down...")

	grpcServ.StopServ()

	errClose := storageRes.Close(context.Background())
	if errClose != nil {
		l.Logger.Error("Resources close error :", zap.Error(errClose))
	} else {
		l.Logger.Info("Resources close complete")
	}

	_ = l.Logger.Sync()
	return nil
}
