// Package service - application logic
package service

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/4aleksei/gokeeper/internal/common/aescoder"
	"github.com/4aleksei/gokeeper/internal/common/datafile"
	"github.com/4aleksei/gokeeper/internal/common/store"
	"github.com/4aleksei/gokeeper/internal/common/utils/random"
	"github.com/4aleksei/gokeeper/internal/server/config"
	"github.com/4aleksei/gokeeper/internal/server/jwtauth"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type (
	serverStorage interface {
		AddUser(context.Context, string, string) (*store.User, error)
		GetUser(context.Context, string) (*store.User, error)
		AddData(context.Context, *store.UserDataCrypt) error
		GetData(context.Context, string) (*store.UserDataCrypt, error)
	}
	serverEncoder interface {
		Encrypt(*store.UserData) (*store.UserDataCrypt, *aescoder.KeyAES, error)
		Decrypt(*store.UserDataCrypt) (*store.UserData, *aescoder.KeyAES, error)
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
	if err != nil {
		return nil, err
	}
	pass := random.HashPass([]byte(password), serv.cfg.Key)
	if passValue.HashPass == hex.EncodeToString(pass) {
		return passValue, nil
	}
	return nil, ErrPassIncorect
}

func (serv *HandlerService) RegisterUser(ctx context.Context, user string, password string) (*store.User, error) {
	pass := random.HashPass([]byte(password), serv.cfg.Key)
	value, err := serv.store.AddUser(ctx, user, hex.EncodeToString(pass))
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

func (serv *HandlerService) AddData(ctx context.Context, dataUser *store.UserData) (string, error) {
	encDataUser, _, err := serv.encoder.Encrypt(dataUser)
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
	dataUser, _, err := serv.encoder.Decrypt(dataEnc)
	if err != nil {
		return nil, err
	}
	return dataUser, err
}

func (serv *HandlerService) genFileName() string {
	name := serv.cfg.FilePath + uuid.New().String() + ".data"
	return name
}

func (serv *HandlerService) CreateDataStream(ctx context.Context, dataUser *store.UserData) (*datafile.LongtermfileWrite, *store.UserDataCrypt, error) {
	nameFile := serv.genFileName()
	dataUser.UserData = nameFile

	encDataUser, key, err := serv.encoder.Encrypt(dataUser)
	if err != nil {
		return nil, nil, err
	}
	f := datafile.NewWrite(nameFile, key)
	err = f.OpenWriter()

	if err != nil {
		return nil, nil, err
	}
	return f, encDataUser, nil
}

func (serv *HandlerService) AddDataStream(ctx context.Context, encDataUser *store.UserDataCrypt) (string, error) {
	err := serv.store.AddData(ctx, encDataUser)
	if err != nil {
		return "", err
	}
	return encDataUser.Uuid, nil
}

func (serv *HandlerService) GetDataStream(ctx context.Context, userId uint64, uuid string) (*store.UserData, *datafile.LongtermfileRead, error) {
	dataEnc, err := serv.store.GetData(ctx, uuid)
	if err != nil {
		return nil, nil, err
	}
	if dataEnc.Id != userId {
		return nil, nil, ErrIncorectUserId
	}
	dataUser, key, err := serv.encoder.Decrypt(dataEnc)
	if err != nil {
		return nil, nil, err
	}
	f := datafile.NewRead(dataUser.UserData, key)
	err = f.OpenReader()
	return dataUser, f, err
}
