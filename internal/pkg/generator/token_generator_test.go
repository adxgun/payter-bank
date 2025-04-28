package generator

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"payter-bank/internal/config"
)

func setupTokenGenerator() TokenGenerator {
	cfg := config.JWTConfig{
		Secret:   "testsecret",
		Issuer:   "test-issuer",
		Audience: "test-audience",
		Expiry:   time.Hour,
	}
	return NewTokenGenerator(cfg)
}

func TestGenerateAndValidate_Success(t *testing.T) {
	gen := setupTokenGenerator()
	data := TokenData{
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	token, err := gen.Generate(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidate_InvalidTokenFormat(t *testing.T) {
	gen := setupTokenGenerator()

	token := "invalid.token.here"

	parsedData, err := gen.Validate(token)
	assert.Error(t, err)
	assert.Equal(t, TokenData{}, parsedData)
}

func TestValidate_InvalidSignature(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:   "differentsecret",
		Issuer:   "test-issuer",
		Audience: "test-audience",
		Expiry:   time.Hour,
	}
	wrongGen := NewTokenGenerator(cfg)

	data := TokenData{
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	token, err := wrongGen.Generate(data)
	assert.NoError(t, err)

	gen := setupTokenGenerator()

	parsedData, err := gen.Validate(token)
	assert.Error(t, err)
	assert.Equal(t, TokenData{}, parsedData)
}

func TestValidate_ExpiredToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:   "testsecret",
		Issuer:   "test-issuer",
		Audience: "test-audience",
		Expiry:   -time.Minute, // Already expired
	}
	gen := NewTokenGenerator(cfg)

	data := TokenData{
		UserID:    uuid.New(),
		AccountID: uuid.New(),
		ExpiresAt: time.Now().Add(-time.Minute),
	}

	token, err := gen.Generate(data)
	assert.NoError(t, err)

	//
	parsedData, err := gen.Validate(token)
	assert.Error(t, err) // Should error due to expiry
	assert.Equal(t, TokenData{}, parsedData)
}
