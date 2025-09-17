package jwtauth

import (
	"errors"
	"time"

	"github.com/4aleksei/gokeeper/internal/server/config"
	"github.com/golang-jwt/jwt/v4"
)

type (
	AuthService struct {
		secretKey string
	}

	Claims struct {
		jwt.RegisteredClaims
		UserID uint64
	}
)

var (
	ErrTokenError            = errors.New("error, value exists")
	ErrTokenErrorUnAlgMethod = errors.New("error,no user")
	ErrNoDB                  = errors.New("no db")
)

const tokenExp = time.Hour * 3

func New(cfg *config.Config) *AuthService {
	return &AuthService{
		secretKey: cfg.Key,
	}
}

func (a *AuthService) BuildJWT(userID uint64) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		// собственное утверждение
		UserID: userID,
	})
	// создаём строку токена
	tokenString, err := token.SignedString([]byte(a.secretKey))
	if err != nil {
		return "", err
	}
	// возвращаем строку токена
	return tokenString, nil
}

func (a *AuthService) GetUserID(tokenString string) (uint64, error) {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, ErrTokenErrorUnAlgMethod //fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(a.secretKey), nil
	})

	if err != nil {
		return 0, ErrTokenError
	}

	if !token.Valid {
		return 0, ErrTokenError
	}
	return claims.UserID, nil
}
