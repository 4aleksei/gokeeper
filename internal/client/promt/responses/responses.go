// Package responses
package responses

const (
	TokenData TypeData = iota
	UserData
	StreamUserData
	UserDataUUID
)

type (
	TypeData int
	Respond  struct {
		typeData     TypeData
		userTypeData int
		data         string
		metadata     string
		err          error
	}
)

func AddUserName(name string) func(*Respond) {
	return func(r *Respond) {
		r.metadata = name
	}
}

func AddToken(token string) func(*Respond) {
	return func(r *Respond) {
		r.data = token
		r.typeData = TokenData
	}
}

func AddUUID(uuid string) func(*Respond) {
	return func(r *Respond) {
		r.data = uuid
		r.typeData = UserDataUUID
	}
}

func AddData(t int, data string, metadata string) func(*Respond) {
	return func(r *Respond) {
		r.data = data
		r.metadata = metadata
		r.userTypeData = t
		r.typeData = UserData
	}
}

func AddError(err error) func(*Respond) {
	return func(r *Respond) {
		r.err = err
	}
}

func New(options ...func(*Respond)) *Respond {
	resp := &Respond{}
	for _, o := range options {
		o(resp)
	}
	return resp
}

func (r *Respond) GetToken() (string, bool) {
	if r.typeData == TokenData {
		return r.data, true
	}
	return "", false
}

func (r *Respond) GetData() (string, bool) {
	if r.typeData == UserData {
		return r.data, true
	}
	return "", false
}

func (r *Respond) GetUUID() (string, bool) {
	if r.typeData == UserDataUUID {
		return r.data, true
	}
	return "", false
}

func (r *Respond) GetMetaData() string {
	return r.metadata
}

func (r *Respond) GetError() error {
	return r.err
}

func (r *Respond) GetType() int {
	return int(r.userTypeData)
}
