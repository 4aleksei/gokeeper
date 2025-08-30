// Package cache - cache
package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/4aleksei/gokeeper/internal/common/store"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type (
	Store struct {
		users     sync.Map
		usersData cacheStore
		l         *zap.Logger
	}

	cacheStore struct {
		lock      sync.RWMutex
		uuidUsers map[uint64][]*store.UserData
		dataUsers map[string]*store.UserData
	}
)

var (
	ErrValueExists   = errors.New("error, value exists")
	ErrUserNotFound  = errors.New("error,no user")
	ErrValueNotFound = errors.New("error,no value")
	ErrNoDB          = errors.New("no db")
)

func (c *cacheStore) AddData(userdata *store.UserData) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, exis := c.dataUsers[userdata.Uuid]
	if exis {
		return ErrValueExists
	}
	c.dataUsers[userdata.Uuid] = userdata
	c.uuidUsers[userdata.Id] = append(c.uuidUsers[userdata.Id], userdata)
	return nil
}

func (c *cacheStore) GetData(uuid string) (*store.UserData, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	data, ok := c.dataUsers[uuid]
	if !ok {
		return nil, ErrValueNotFound
	}
	return data, nil
}

func (c *cacheStore) GetList(userID uint64) ([]*store.UserData, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	data, ok := c.uuidUsers[userID]
	if !ok {
		return nil, ErrValueNotFound
	}
	return data, nil
}

func New(l *zap.Logger) *Store {
	return &Store{l: l}
}

var idUsers atomic.Uint64

func getId() uint64 {
	idUsers.Add(1)
	return idUsers.Load()
}

func (s *Store) AddUser(ctx context.Context, user string, pass string) (*store.User, error) {
	userSt := &store.User{
		Name:     user,
		HashPass: pass,
		Id:       getId(),
	}
	_, ok := s.users.LoadOrStore(user, userSt)
	if ok {
		return nil, ErrUserNotFound
	}
	return userSt, nil
}

func (s *Store) GetUser(ctx context.Context, user string) (*store.User, error) {
	val, ok := s.users.Load(user)
	if ok {
		return val.(*store.User), nil
	}
	return nil, ErrValueNotFound
}

func (s *Store) AddData(ctx context.Context, userdata *store.UserData) error {
	uuid := uuid.New()
	userdata.Uuid = uuid.String()
	userdata.TimeStamp = time.Now()

	//userdata.EnKey= datacrypto.encrypt(userdata.UserData,userdata.MetaData)
	err := s.usersData.AddData(userdata)
	if err != nil {
		return ErrValueExists
	}
	return nil
}

func (s *Store) GetData(ctx context.Context, uuid string) (*store.UserData, error) {
	data, err := s.usersData.GetData(uuid)
	if err != nil {
		return nil, ErrValueNotFound
	}

	// userdata.UserData,userdata.MetaData=datacrypto.decrypt(userdata.UserData,userdata.MetaData,userdata.EnKey)

	return data, nil
}
