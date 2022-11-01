package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type JwtWrapper struct {
	TokenKey            string
	TokenExpires        int64
	AccessTokenKey      string
	AccessTokenExpires  int64
	RefreshTokenKey     string
	RefreshTokenExpires int64
	Issuer              string
	ExpirationHours     int64
}

type jwtClaims struct {
	jwt.StandardClaims
	Id    uuid.UUID
	Email string
}

type JwtData struct {
	Id    uuid.UUID
	Email string
}

func (w *JwtWrapper) GenerateToken(data JwtData, tokenType string) (signedToken string, err error) {
	claims := &jwtClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(w.ExpirationHours)).Unix(),
			Issuer:    w.Issuer,
		},
		Id:    data.Id,
		Email: data.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	switch tokenType {
	case "ACCESS_TOKEN":
		signedToken, err = token.SignedString([]byte(w.AccessTokenKey))
	case "REFRESH_TOKEN":
		signedToken, err = token.SignedString([]byte(w.RefreshTokenKey))
	default:
		signedToken, err = token.SignedString([]byte(w.TokenKey))
	}
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (w *JwtWrapper) ValidateToken(signedToken string, tokenType string) (claims *jwtClaims, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&jwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			switch tokenType {
			case "ACCESS_TOKEN":
				return []byte(w.AccessTokenKey), nil
			case "REFRESH_TOKEN":
				return []byte(w.RefreshTokenKey), nil
			default:
				return []byte(w.TokenKey), nil
			}
		},
	)

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*jwtClaims)

	if !ok {
		return nil, errors.New("Couldn't parse claims")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, errors.New("JWT is expired")
	}

	return claims, nil

}
