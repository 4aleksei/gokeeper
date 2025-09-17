package transaction

import (
	"errors"
)

var (
	ErrBadTypeCommand  = errors.New("unk  type command")
	ErrBadTypeResponse = errors.New("unk  type response")
)

type (
	User struct {
		Name     string
		Password string
	}

	TokenUser struct {
		Token string
	}

	UUIDData struct {
		UUID string
	}

	UserLogin struct {
		User User
	}

	UserRegister struct {
		User User
	}

	GetUserData struct {
		Token TokenUser
		UUID  UUIDData
	}

	UserData struct {
		Token    TokenUser
		TypeData int
		Data     string
		MetaData string
	}

	StreamData struct {
		Token    TokenUser
		TypeData int
		MetaData string
		Output   chan []byte
	}

	GetStreamData struct {
		Token TokenUser
		UUID  UUIDData
		Input chan []byte
	}

	Request struct {
		Command any
	}

	Response struct {
		Resp any
		Err  error
	}
)
