package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenClaims struct {
	EntityID   string `json:"entityId,omitempty"`
	EntityType string `json:"entityType,omitempty"`
	jwt.RegisteredClaims
}

type RefreshTokens struct {
	NewAccessToken        string       `json:"accessToken"`
	NewRefreshToken       string       `json:"refreshToken"`
	NewAccessTokenClaims  *TokenClaims `json:"accessTokenClaims"`
	NewRefreshTokenClaims *TokenClaims `json:"refreshTokenClaims"`
}

// type TokenServicer interface {
// 	GenerateToken(isRefreshToken bool, entityID string, entityType string) (string, *TokenClaims, error)
// 	ValidateAccessToken(tokenStr string) (isValid bool, claims *TokenClaims, err error)
// 	ValidateRefreshToken(tokenStr string) (isValid bool, claims *TokenClaims, err error)
// 	RefreshTokens(entityID string, entityType string) (*RefreshTokens, error)
// }

type TokenService struct {
	AccessTokenSecret        []byte
	RefreshTokenSecret       []byte
	AccessTokenExpiryInSecs  int64
	RefreshTokenExpiryInSecs int64
}

func NewTokenService(accessTokenSecret, refreshTokenSecret string,
	accessTokenExpiryInSecs, refreshTokenExpiryInSecs int64) *TokenService {
	return &TokenService{
		AccessTokenSecret:        []byte(accessTokenSecret),
		RefreshTokenSecret:       []byte(refreshTokenSecret),
		AccessTokenExpiryInSecs:  accessTokenExpiryInSecs,
		RefreshTokenExpiryInSecs: refreshTokenExpiryInSecs,
	}
}

func (tm *TokenService) GenerateToken(isRefreshToken bool, entityID string, entityType string) (tokenStr string, claims *TokenClaims, err error) {
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

	claims = &TokenClaims{
		EntityID:   entityID,
		EntityType: entityType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Issuer:    "aa_backend", // todo: correct this
			Subject:   fmt.Sprintf("%s_%s", entityType, entityID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err = token.SignedString(secret)
	if err != nil {
		return "", nil, err
	}

	return tokenStr, claims, nil
}

func (tm *TokenService) ValidateAccessToken(tokenStr string) (isValid bool, claims *TokenClaims, err error) {
	return tm.validateToken(tokenStr, tm.AccessTokenSecret)
}

func (tm *TokenService) ValidateRefreshToken(tokenStr string) (isValid bool, claims *TokenClaims, err error) {
	return tm.validateToken(tokenStr, tm.RefreshTokenSecret)
}

func (tm *TokenService) RefreshTokens(entityID string, entityType string) (*RefreshTokens, error) {
	// _, claims, err := tm.validateToken(refreshToken, tm.RefreshTokenSecret)
	// if err != nil {
	// 	return "", "", err
	// }

	newAccessToken, newAccessTokenClaims, err := tm.GenerateToken(false, entityID, entityType)
	if err != nil {
		return nil, err
	}

	newRefreshToken, newRefreshTokenClaims, err := tm.GenerateToken(true, entityID, entityType)
	if err != nil {
		return nil, err
	}

	return &RefreshTokens{
		NewAccessToken:        newAccessToken,
		NewRefreshToken:       newRefreshToken,
		NewAccessTokenClaims:  newAccessTokenClaims,
		NewRefreshTokenClaims: newRefreshTokenClaims,
	}, nil

}

func (tm *TokenService) validateToken(tokenStr string, secret []byte) (isValid bool, claims *TokenClaims, err error) {

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&TokenClaims{},
		func(token *jwt.Token) (any, error) {
			return secret, nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) { // todo: revisit this
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("error parsing token: %s", err)
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return true, claims, nil
	}

	return false, nil, err
}
