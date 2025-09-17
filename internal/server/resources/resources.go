// Package resources - Open resources , Files,Store , DB ..etc
package resources

import (
	"context"

	"github.com/4aleksei/gokeeper/internal/common/aescoder"
	"github.com/4aleksei/gokeeper/internal/common/cryptocerts"
	"github.com/4aleksei/gokeeper/internal/common/datacrypto"
	"github.com/4aleksei/gokeeper/internal/common/logger"
	"github.com/4aleksei/gokeeper/internal/common/store"
	"github.com/4aleksei/gokeeper/internal/common/store/cache"
	"github.com/4aleksei/gokeeper/internal/server/config"
)

type (
	resoucesStorage interface {
		AddUser(context.Context, string, string) (*store.User, error)
		GetUser(context.Context, string) (*store.User, error)
		AddData(context.Context, *store.UserDataCrypt) error
		GetData(context.Context, string) (*store.UserDataCrypt, error)
	}
	resourceEncoder interface {
		Encrypt(*store.UserData) (*store.UserDataCrypt, *aescoder.KeyAES, error)
		Decrypt(*store.UserDataCrypt) (*store.UserData, *aescoder.KeyAES, error)
	}

	handleResources struct {
		Store resoucesStorage
		Enc   resourceEncoder
	}
)

func New(cfg *config.Config, l *logger.ZapLogger) (*handleResources, error) {
	pr, pub, _ := cryptocerts.GenerateKey()
	crypto := datacrypto.New(pr, pub)
	return &handleResources{
		Store: cache.New(l.Logger),
		Enc:   crypto,
	}, nil
}

func (r *handleResources) Close(ctx context.Context) error {
	return nil
}
