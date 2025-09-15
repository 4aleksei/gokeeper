package encoder

import (
	"github.com/4aleksei/gokeeper/internal/common/aescoder"
	"github.com/4aleksei/gokeeper/internal/common/store"
)

type (
	ServerEncoder interface {
		Encrypt(*store.UserData) (*store.UserDataCrypt, *aescoder.KeyAES, error)
		Decrypt(*store.UserDataCrypt) (*store.UserData, *aescoder.KeyAES, error)
	}
)
