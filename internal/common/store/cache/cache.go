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
	StoreCache struct {
		users     sync.Map
		usersData cacheStore
		l         *zap.Logger
	}

	cacheStore struct {
		lock      sync.RWMutex
		uuidUsers map[uint64][]*store.UserDataCrypt
		dataUsers map[string]*store.UserDataCrypt
	}
)

var (
	ErrValueExists   = errors.New("error, value exists")
	ErrUserExists    = errors.New("error, user exists")
	ErrUserNotFound  = errors.New("error,no user")
	ErrValueNotFound = errors.New("error,no value")
	ErrNoDB          = errors.New("no db")
)

func New(l *zap.Logger) *StoreCache {
	stor := &StoreCache{l: l}

	stor.usersData.uuidUsers = make(map[uint64][]*store.UserDataCrypt)
	stor.usersData.dataUsers = make(map[string]*store.UserDataCrypt)
	return stor
}

func (c *cacheStore) AddData(userdata *store.UserDataCrypt) error {
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

func (c *cacheStore) GetData(uuid string) (*store.UserDataCrypt, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	data, ok := c.dataUsers[uuid]
	if !ok {
		return nil, ErrValueNotFound
	}
	return data, nil
}

func (c *cacheStore) GetList(userID uint64) ([]*store.UserDataCrypt, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	data, ok := c.uuidUsers[userID]
	if !ok {
		return nil, ErrValueNotFound
	}
	return data, nil
}

var idUsers atomic.Uint64

func getId() uint64 {
	idUsers.Add(1)
	return idUsers.Load()
}

func (s *StoreCache) AddUser(ctx context.Context, user string, pass string) (*store.User, error) {
	userSt := &store.User{
		Name:     user,
		HashPass: pass,
		Id:       getId(),
	}
	s.l.Debug("Add user", zap.String("Name", user))
	_, ok := s.users.LoadOrStore(user, userSt)
	if ok {
		return nil, ErrUserExists
	}
	return userSt, nil
}

func (s *StoreCache) GetUser(ctx context.Context, user string) (*store.User, error) {
	val, ok := s.users.Load(user)
	s.l.Debug("Get user", zap.String("Name", user), zap.Bool("GETED", ok))
	if ok {
		return val.(*store.User), nil
	}
	return nil, ErrUserNotFound
}

func (s *StoreCache) AddData(ctx context.Context, userdata *store.UserDataCrypt) error {
	uuid := uuid.New()
	userdata.Uuid = uuid.String()
	userdata.TimeStamp = time.Now()
	err := s.usersData.AddData(userdata)
	if err != nil {
		return ErrValueExists
	}
	return nil
}

func (s *StoreCache) GetData(ctx context.Context, uuid string) (*store.UserDataCrypt, error) {
	data, err := s.usersData.GetData(uuid)
	if err != nil {
		return nil, ErrValueNotFound
	}
	return data, nil
}
