package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

type TokenMaker interface {
	GenerateToken(isRefreshToken bool, userID string) (string, *TokenClaims, error)
	ValidateAccessToken(tokenStr string) (isValid bool, claims *TokenClaims, err error)
	ValidateRefreshToken(tokenStr string) (isValid bool, claims *TokenClaims, err error)
	RefreshTokens(refreshToken string) (newAccessToken string, newRefreshToken string, err error)
}

type TokenManager struct {
	AccessTokenSecret        []byte
	RefreshTokenSecret       []byte
	AccessTokenExpiryInSecs  int64
	RefreshTokenExpiryInSecs int64
}

func NewTokenManager(accessTokenSecret, refreshTokenSecret string,
	accessTokenExpiryInSecs, refreshTokenExpiryInSecs int64) *TokenManager {
	return &TokenManager{
		AccessTokenSecret:        []byte(accessTokenSecret),
		RefreshTokenSecret:       []byte(refreshTokenSecret),
		AccessTokenExpiryInSecs:  accessTokenExpiryInSecs,
		RefreshTokenExpiryInSecs: refreshTokenExpiryInSecs,
	}
}

func (tm *TokenManager) GenerateToken(isRefreshToken bool, userID string) (string, *TokenClaims, error) {
	var (
		tokenID string
		secret  []byte
		expiry  time.Duration
	)

	secret = tm.AccessTokenSecret
	expiry = time.Second * time.Duration(tm.AccessTokenExpiryInSecs)

	if isRefreshToken {
		tokenID = uuid.New().String()
		secret = tm.RefreshTokenSecret
		expiry = time.Second * time.Duration(tm.RefreshTokenExpiryInSecs)
	}

	// claims := jwt.MapClaims{
	// 	"userID": userID,
	// 	"iat":    time.Now().Unix(),
	// 	"exp":    time.Now().Add(expiry).Unix(),
	// }

	// claims := jwt.RegisteredClaims{
	// 	ID:        uuid.New().String(),
	// 	ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
	// 	IssuedAt:  jwt.NewNumericDate(time.Now()),
	// 	NotBefore: jwt.NewNumericDate(time.Now()),
	// }

	claims := TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Issuer:    "aa_backend", // todo: correct this
			Subject:   "app",        // todo correct this to token type
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString(secret)
	if err != nil {
		return "", nil, err
	}

	return tokenStr, &claims, nil
}

func (tm *TokenManager) ValidateAccessToken(tokenStr string) (isValid bool, claims *TokenClaims, err error) {
	return tm.validateToken(tokenStr, tm.AccessTokenSecret)
}

func (tm *TokenManager) ValidateRefreshToken(tokenStr string) (isValid bool, claims *TokenClaims, err error) {
	return tm.validateToken(tokenStr, tm.RefreshTokenSecret)
}

func (tm *TokenManager) RefreshTokens(refreshToken string) (newAccessToken string, newRefreshToken string, err error) {
	_, claims, err := tm.validateToken(refreshToken, tm.RefreshTokenSecret)
	if err != nil {
		return "", "", err
	}

	newAccessToken, _, err = tm.GenerateToken(false, claims.UserID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, _, err = tm.GenerateToken(true, claims.UserID)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil

}

func (tm *TokenManager) validateToken(tokenStr string, secret []byte) (isValid bool, claims *TokenClaims, err error) {

	token, err := jwt.ParseWithClaims(
		tokenStr,
		TokenClaims{},
		func(token *jwt.Token) (any, error) {
			return secret, nil
		},
	)

	if err != nil {
		return false, nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return true, claims, nil
	}

	return false, nil, err
}
