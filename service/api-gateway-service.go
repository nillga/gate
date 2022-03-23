package service

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/nillga/jwt-server/entity"
)

type GatewayService interface {
	BuildCooker(user *entity.User) (*http.Cookie, error)
	ReadBearer(authorizationHeader string) (string, error)
	ReadToken(token string) (*entity.User, error)
}

type service struct{}

func NewService() GatewayService {
	return &service{}
}

type Claims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Mail     string `json:"email"`
	IsAdmin  bool   `json:"admin"`
	jwt.StandardClaims
}

func (c *Claims) decodeJwt(token string) error {
	if _, err := jwt.ParseWithClaims(token, c, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	}); err != nil {
		return err
	}
	return nil
}

var secretKey = os.Getenv("SECRET_KEY")

func (s *service) ReadToken(token string) (*entity.User, error) {
	claims := &Claims{}

	if err := claims.decodeJwt(token); err != nil {
		return nil, err
	}

	return &entity.User{
		Id:       claims.Id,
		Username: claims.Username,
		Email:    claims.Mail,
		Admin:    claims.IsAdmin,
	}, nil
}

func (s *service) ReadBearer(authorizationHeader string) (string, error) {
	if authorizationHeader == "" {
		return "", errors.New("no auth provided")
	}

	authorizationParts := strings.Split(authorizationHeader, "Bearer")
	if len(authorizationParts) != 2 {
		return "", errors.New("invalid auth syntax")
	}
	token := strings.TrimSpace(authorizationParts[1])
	if len(token) < 1 {
		return "", errors.New("invalid token syntax")
	}

	return token, nil
}

func (s *service) BuildCooker(user *entity.User) (*http.Cookie, error) {
	claims := &Claims{
		Id:       user.Id,
		Username: user.Username,
		Mail:     user.Email,
		IsAdmin:  user.Admin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 2).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:    "jwt",
		Value:   tokenString,
		Expires: time.Now().Add(time.Hour * 2),
		HttpOnly: true,
	}, nil
}
