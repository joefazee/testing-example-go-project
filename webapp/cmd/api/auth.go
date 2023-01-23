package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"webapp/pkg/data"

	"github.com/golang-jwt/jwt/v4"
)

var jwtTokenExpiry = time.Minute * 30
var refreshTokenExpiry = time.Hour * 24

type TokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	UserName string `json:"name"`
	jwt.RegisteredClaims
}

func (app *application) getTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {

	w.Header().Add("Vary", "Authorization")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", nil, errors.New("no auth header")
	}

	headersParts := strings.Split(authHeader, " ")
	if len(headersParts) != 2 {
		return "", nil, errors.New("invalid auth header")
	}

	if headersParts[0] != "Bearer" {
		return "", nil, errors.New("unauthorized: no Bearer")
	}

	token := headersParts[1]

	claims := &Claims{}

	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unable signing method %v", t.Header["alg"])
		}

		return []byte(app.JWTSecret), nil
	})

	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("expired token")
		}

		return "", nil, err
	}

	if claims.Issuer != app.Domain {
		return "", nil, errors.New("incorrect issuer")
	}

	return token, claims, nil
}

func (app *application) generateTokenPair(user *data.User) (TokenPairs, error) {

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprintf("%d", user.ID)
	claims["aud"] = app.Domain
	claims["iss"] = app.Domain
	claims["admin"] = user.IsAdmin == 1

	claims["exp"] = time.Now().Add(jwtTokenExpiry).Unix()

	signedAccessToken, err := token.SignedString([]byte(app.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create refresh token
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprintf("%d", user.ID)
	refreshTokenClaims["exp"] = time.Now().Add(refreshTokenExpiry).Unix()

	signedRefreshToken, err := refreshToken.SignedString([]byte(app.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}

	var tokenPairs = TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}

	return tokenPairs, nil

}
