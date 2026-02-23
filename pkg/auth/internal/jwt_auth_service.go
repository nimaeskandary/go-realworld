package internal

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthService struct {
	secretKey     []byte
	tokenDuration time.Duration
	userService   user_types.UserService
}

func NewJWTAuthService(cfg auth_types.JwtAuthServiceConfig, userService user_types.UserService) (auth_types.AuthService, error) {
	secret, err := base64.StdEncoding.DecodeString(string(cfg.SecretBase64))
	if err != nil {
		return nil, fmt.Errorf("failed to decode jwt secret: %v", err)
	}

	return &JWTAuthService{
		secretKey:     secret,
		tokenDuration: time.Duration(cfg.TokenDurationSeconds) * time.Second,
		userService:   userService,
	}, nil
}

func (s *JWTAuthService) GenerateToken(ctx context.Context, username string) (auth_types.AuthToken, auth_types.DomainError) {
	var err error
	user, err := s.userService.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, auth_types.AsDomainError(fmt.Errorf("error getting user for token: %w", err))
	}

	claims := &jwtClaims{
		UserId:   user.Id,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(s.secretKey)
	if err != nil {
		return nil, auth_types.AsDomainError(err)
	}

	return NewJwtToken(ss, claims, user), nil
}

func (s *JWTAuthService) ParseToken(ctx context.Context, tokenStr string) (auth_types.AuthToken, auth_types.DomainError) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		return s.secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, auth_types.InvalidTokenError{Reason: "token is invalid"}
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, auth_types.InvalidTokenError{Reason: "cannot parse token claims"}
	}

	user, err := s.userService.GetUserByUsername(ctx, claims.Username)
	if err != nil {
		return nil, auth_types.AsDomainError(fmt.Errorf("error getting user for token: %w", err))
	}

	authToken := NewJwtToken(token.Raw, claims, user)
	if authToken.IsExpired() {
		return nil, auth_types.ExpiredTokenError{}
	}

	return authToken, nil
}
