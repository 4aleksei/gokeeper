// Package service - application logic
package service

import (
	"context"
	"errors"

	"github.com/4aleksei/gokeeper/internal/common/store"
	"github.com/4aleksei/gokeeper/internal/common/utils"
	"github.com/4aleksei/gokeeper/internal/server/config"
	"github.com/4aleksei/gokeeper/internal/server/jwtauth"
	"go.uber.org/zap"
	// "github.com/4aleksei/gokeeper/internal/common/datacrypto"
)

type (
	serverStorage interface {
		AddUser(context.Context, string, string) (*store.User, error)
		GetUser(context.Context, string) (*store.User, error)
		AddData(context.Context, *store.UserDataCrypt) error
		GetData(context.Context, string) (*store.UserDataCrypt, error)
	}
	serverEncoder interface {
		Encrypt(*store.UserData) (*store.UserDataCrypt, error)
		Decrypt(*store.UserDataCrypt) (*store.UserData, error)
	}

	HandlerService struct {
		store   serverStorage
		l       *zap.Logger
		auth    *jwtauth.AuthService
		cfg     *config.Config
		encoder serverEncoder
	}
)

var (
	ErrPassIncorect   = errors.New("error, pass incorect")
	ErrIncorectUserId = errors.New("error, id user error")
)

func New(s serverStorage, enc serverEncoder, l *zap.Logger, c *config.Config) *HandlerService {
	return &HandlerService{
		store:   s,
		l:       l,
		cfg:     c,
		auth:    jwtauth.New(c),
		encoder: enc,
	}
}

func (serv *HandlerService) LoginUser(ctx context.Context, user string, password string) (*store.User, error) {
	passValue, err := serv.store.GetUser(ctx, user)
	pass := utils.HashPass([]byte(password), serv.cfg.Key)
	if err != nil {
		return nil, err
	}
	if passValue.HashPass == string(pass) {
		return passValue, nil
	}
	return nil, ErrPassIncorect
}

func (serv *HandlerService) RegisterUser(ctx context.Context, user string, password string) (*store.User, error) {
	pass := utils.HashPass([]byte(password), serv.cfg.Key)

	value, err := serv.store.AddUser(ctx, user, string(pass))
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (serv *HandlerService) BuildToken(ctx context.Context, user *store.User) (string, error) {
	return serv.auth.BuildJWT(user.Id)
}

func (serv *HandlerService) CheckToken(ctx context.Context, token string) (uint64, error) {
	return serv.auth.GetUserID(token)
}

func (serv *HandlerService) AddData(ctx context.Context, userId uint64, typeData int, data string, metadata string) (string, error) {
	dataUser := &store.UserData{
		Id:       userId,
		TypeData: typeData,
		UserData: data,
		MetaData: metadata,
	}
	encDataUser, err := serv.encoder.Encrypt(dataUser)
	if err != nil {
		return "", err
	}
	err = serv.store.AddData(ctx, encDataUser)
	if err != nil {
		return "", err
	}
	return encDataUser.Uuid, nil
}

func (serv *HandlerService) GetData(ctx context.Context, userId uint64, uuid string) (*store.UserData, error) {
	dataEnc, err := serv.store.GetData(ctx, uuid)
	if err != nil {
		return nil, err
	}
	if dataEnc.Id != userId {
		return nil, ErrIncorectUserId
	}
	dataUser, err := serv.encoder.Decrypt(dataEnc)
	if err != nil {
		return nil, err
	}
	return dataUser, err
}
