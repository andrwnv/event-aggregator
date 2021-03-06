package services

import (
	"fmt"
	"github.com/andrwnv/event-aggregator/core/dto"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/joho/godotenv"
	"os"
	"time"
)

type JWTService interface {
	GenerateToken(email string, user dto.BaseUserInfo) string
	ValidateToken(token string) (*jwt.Token, error)
}

type authCustomClaims struct {
	Name string           `json:"name"`
	User dto.BaseUserInfo `json:"user"`
	jwt.StandardClaims
}

type _jwtServices struct {
	secretKey string
	issue     string
}

func JWTAuthService() JWTService {
	return &_jwtServices{
		secretKey: getSecretKey(),
		issue:     "take-place",
	}
}

func getSecretKey() string {
	godotenv.Load(".env")
	secret := os.Getenv("SECRET")
	if secret == "" {
		secret = "secret"
	}
	return secret
}

func (service *_jwtServices) GenerateToken(email string, user dto.BaseUserInfo) string {
	claims := &authCustomClaims{
		email,
		user,
		jwt.StandardClaims{
			ExpiresAt: jwt.At(time.Now().Add(time.Hour * 48)),
			Issuer:    service.issue,
			IssuedAt:  jwt.Now(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(service.secretKey))
	if err != nil {
		panic(err)
	}
	return t
}

func (service *_jwtServices) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, isValid := token.Method.(*jwt.SigningMethodHMAC); !isValid {
			return nil, fmt.Errorf("invalid token")
		}
		return []byte(service.secretKey), nil
	})
}
