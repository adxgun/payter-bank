package auth

import (
	"context"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"net/http"
	"net/http/httptest"
	"payter-bank/internal/pkg/generator"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetTokenData_Success(t *testing.T) {
	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginCtx.Request = &http.Request{}
	userID, accountID := uuid.New(), uuid.New()

	expectedTokenData := generator.TokenData{
		UserID:    userID,
		AccountID: accountID,
	}

	claims := &validator.ValidatedClaims{
		CustomClaims: &generator.Claim{
			TokenData: expectedTokenData,
		},
	}

	ctxWithClaims := context.WithValue(ginCtx.Request.Context(), jwtmiddleware.ContextKey{}, claims)
	ginCtx.Request = ginCtx.Request.WithContext(ctxWithClaims)

	tokenData, err := GetTokenData(ginCtx)

	assert.NoError(t, err)
	assert.Equal(t, expectedTokenData, tokenData)
}

func TestGetTokenData_MissingClaims(t *testing.T) {
	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginCtx.Request = &http.Request{}

	tokenData, err := GetTokenData(ginCtx)

	assert.Error(t, err)
	assert.Equal(t, generator.TokenData{}, tokenData)
}

func TestGetTokenData_InvalidCustomClaims(t *testing.T) {
	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginCtx.Request = &http.Request{}

	ctxWithClaims := context.WithValue(ginCtx.Request.Context(), jwtmiddleware.ContextKey{}, struct{}{})
	ginCtx.Request = ginCtx.Request.WithContext(ctxWithClaims)

	tokenData, err := GetTokenData(ginCtx)

	assert.Error(t, err)
	assert.Equal(t, generator.TokenData{}, tokenData)
}

func TestGetCurrentProfile_Success(t *testing.T) {
	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginCtx.Request = &http.Request{}

	expectedProfile := Profile{
		AccountID:    uuid.New(),
		UserID:       uuid.New(),
		Email:        "test@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		AccountType:  "CURRENT",
		UserType:     "ADMIN",
		RegisteredAt: time.Now(),
	}

	ctxWithProfile := context.WithValue(ginCtx.Request.Context(), ProfileKey, expectedProfile)
	ginCtx.Request = ginCtx.Request.WithContext(ctxWithProfile)

	profile, err := GetCurrentProfile(ginCtx)

	assert.NoError(t, err)
	assert.Equal(t, expectedProfile, profile)
}

func TestGetCurrentProfile_MissingProfile(t *testing.T) {
	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginCtx.Request = &http.Request{}

	profile, err := GetCurrentProfile(ginCtx)

	assert.Error(t, err)
	assert.Equal(t, Profile{}, profile)
}
