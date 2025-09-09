// Package store - store structs
package store

import (
	"errors"
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

var (
	ErrBadType = errors.New("error type id_text")

	typesMAP = map[string]int{
		"login":  0,
		"card":   1,
		"text":   2,
		"binary": 3,
	}

	typesTab = []string{"login", "card", "text", "binary"}
)

func GetType(str string) (int, error) {
	v, ok := typesMAP[str]
	if !ok {
		return -1, ErrBadType
	}
	return v, nil
}

func GetStringType(t int) string {
	if t >= len(typesTab) {
		return "nan"
	}
	return typesTab[t]
}
