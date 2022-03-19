package service

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/nillga/jwt-server/entity"
	"net/http"
	"os"
	"time"
)

type GatewayService interface {
	BuildCooker(user *entity.User) (*http.Cookie, error)
	ReadCooker(cookie *http.Cookie) (*entity.User, error)
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

func (c *Claims) decodeJwt(cookie *http.Cookie) error {
	if _, err := jwt.ParseWithClaims(cookie.Value, c, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	}); err != nil {
		return err
	}
	return nil
}

var secretKey = os.Getenv("SECRET_KEY")

func (s *service) ReadCooker(cookie *http.Cookie) (*entity.User, error) {
	claims := &Claims{}

	if err := claims.decodeJwt(cookie); err != nil {
		return nil, err
	}

	return &entity.User{
		Id:       claims.Id,
		Username: claims.Username,
		Email:    claims.Mail,
		Admin:    claims.IsAdmin,
	}, nil
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
		Path:    "/",
	}, nil
}
