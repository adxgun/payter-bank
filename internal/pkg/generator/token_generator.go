//go:generate mockgen -source=token_generator.go -destination=./mocks/mocks.go -package=generatormocks

package generator

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"payter-bank/config"
	"payter-bank/features/account"
	"time"
)

// TokenGenerator is an interface for generating and validating JWT tokens.
type TokenGenerator interface {
	Generate(profile account.Profile) (string, error)
	Validate(token string) (account.Profile, error)
}

type Claim struct {
	account.Profile
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

func (g *tokenGenerator) Generate(p account.Profile) (string, error) {
	claims := &Claim{
		Profile: p,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    g.cfg.Issuer,
			Subject:   p.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.cfg.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Audience:  []string{g.cfg.Audience},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.cfg.Secret))
}

func (g *tokenGenerator) Validate(token string) (account.Profile, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(g.cfg.Secret), nil
	})
	if err != nil {
		return account.Profile{}, err
	}

	claims, ok := parsedToken.Claims.(Claim)
	if !ok || !parsedToken.Valid {
		return account.Profile{}, errors.New("invalid token")
	}

	return claims.Profile, nil
}
