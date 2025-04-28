package auth

import (
	"errors"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payter-bank/internal/pkg/generator"
	"time"
)

var ProfileKey = "current_profile"

type Profile struct {
	AccountID    uuid.UUID `json:"account_id"`
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	AccountType  string    `json:"account_type"`
	UserType     string    `json:"user_type"`
	RegisteredAt time.Time `json:"registered_at"`
}

func GetTokenData(ctx *gin.Context) (generator.TokenData, error) {
	claims, ok := ctx.Request.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok {
		return generator.TokenData{}, errors.New("could not get claims from context")
	}

	customClaims, ok := claims.CustomClaims.(*generator.Claim)
	if !ok {
		return generator.TokenData{}, errors.New("could not get custom claims from context")
	}

	return customClaims.TokenData, nil
}

func GetCurrentProfile(ctx *gin.Context) (Profile, error) {
	profile, ok := ctx.Request.Context().Value(ProfileKey).(Profile)
	if !ok {
		return Profile{}, errors.New("could not get profile from context")
	}
	return profile, nil
}
