// Package store - store structs
package store

import (
	"time"
)

type (
	User struct {
		Id       uint64
		Name     string
		HashPass string
	}

	UserData struct {
		Id        uint64
		Uuid      string
		TypeData  int
		UserData  string
		MetaData  string
		TimeStamp time.Time
	}

	UserDataCrypt struct {
		Id         uint64
		Uuid       string
		TypeData   int
		UserDataEn []byte
		MetaDataEn []byte
		EnKey      string
		TimeStamp  time.Time
	}
)
