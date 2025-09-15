package storage

import (
	"context"

	"github.com/4aleksei/gokeeper/internal/common/store"
)

type (
	ServerStorage interface {
		AddUser(context.Context, string, string) (*store.User, error)
		GetUser(context.Context, string) (*store.User, error)
		AddData(context.Context, *store.UserDataCrypt) error
		GetData(context.Context, string) (*store.UserDataCrypt, error)
	}
)
