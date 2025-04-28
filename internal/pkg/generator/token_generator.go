//go:generate mockgen -source=token_generator.go -destination=./mocks/mocks.go -package=generatormocks

package generator

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"payter-bank/internal/config"
	"time"
)

// TokenGenerator is an interface for generating and validating JWT tokens.
type TokenGenerator interface {
	Generate(data TokenData) (string, error)
	Validate(token string) (TokenData, error)
}

type TokenData struct {
	UserID    uuid.UUID
	AccountID uuid.UUID
	ExpiresAt time.Time
}

type Claim struct {
	TokenData
	jwt.RegisteredClaims
}

func (a Claim) Validate(ctx context.Context) error {
	return nil
}

type tokenGenerator struct {
	cfg config.JWTConfig
}

func NewTokenGenerator(cfg config.JWTConfig) TokenGenerator {
	return &tokenGenerator{
		cfg: cfg,
	}
}

func (g *tokenGenerator) Generate(data TokenData) (string, error) {
	claims := &Claim{
		TokenData: data,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    g.cfg.Issuer,
			Subject:   data.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.cfg.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Audience:  []string{g.cfg.Audience},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.cfg.Secret))
}

func (g *tokenGenerator) Validate(tk string) (TokenData, error) {
	parsedToken, err := jwt.Parse(tk, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(g.cfg.Secret), nil
	})
	if err != nil {
		return TokenData{}, err
	}

	claims, ok := parsedToken.Claims.(Claim)
	if !ok || !parsedToken.Valid {
		return TokenData{}, errors.New("invalid token")
	}

	data := claims.TokenData
	return TokenData{
		UserID:    data.UserID,
		AccountID: data.AccountID,
		ExpiresAt: data.ExpiresAt,
	}, nil
}
