package account

import (
	"bytes"
	"context"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"payter-bank/internal/api"
	platformerrors "payter-bank/internal/errors"
	"payter-bank/internal/pkg/generator"
	"testing"
)

func TestHandler_CreateAccountHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successfully create a new account", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		expectedParam := CreateAccountParams{
			Currency: "GBP",
			UserID:   userID,
		}
		expectedProfile := Profile{AccountID: uuid.New()}
		expectedResponse := api.SuccessResponse{
			Data:    expectedProfile,
			Message: "account created successfully",
		}

		body := `{"currency":"GBP", "user_id": "12345678-1234-1234-1234-123456789012"}`
		mockService.EXPECT().CreateAccount(gomock.Any(), expectedParam).
			Return(expectedProfile, nil)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("POST", "/v1/api/accounts", bytes.NewBufferString(body))
		resp := handler.CreateAccountHandler(c)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, resp.Data, expectedResponse)
	})

	t.Run("failed to create account - missing user_id", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))

		body := `{"currency":"GBP"}`
		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("POST", "/v1/api/accounts", bytes.NewBufferString(body))
		resp := handler.CreateAccountHandler(c)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("failed to create account - internal platform error", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		expectedParam := CreateAccountParams{
			Currency:       "GBP",
			InitialDeposit: 100,
			UserID:         userID,
		}

		body := `{"currency":"GBP", "user_id": "12345678-1234-1234-1234-123456789012", "initial_deposit": 100}`
		mockService.EXPECT().CreateAccount(gomock.Any(), expectedParam).
			Return(Profile{}, platformerrors.ErrInternal)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("POST", "/v1/api/accounts", bytes.NewBufferString(body))
		resp := handler.CreateAccountHandler(c)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("failed to create account - pre condition failure", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		expectedParam := CreateAccountParams{
			Currency: "GBP",
			UserID:   userID,
		}

		body := `{"currency":"GBP", "user_id": "12345678-1234-1234-1234-123456789012"}`
		mockService.EXPECT().CreateAccount(gomock.Any(), expectedParam).
			Return(Profile{}, platformerrors.MakeApiError(http.StatusPreconditionFailed, "user already exists"))

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("POST", "/v1/api/accounts", bytes.NewBufferString(body))
		resp := handler.CreateAccountHandler(c)
		assert.Equal(t, http.StatusPreconditionFailed, resp.Code)
		assert.Equal(t, resp.Error.Message, "user already exists")
	})
}

func TestHandler_AuthenticateAccountHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successfully authenticate an account", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		expectedParam := AuthenticateAccountParams{
			Email:    "jd@testmail.com",
			Password: "$PASSword001",
		}
		expectedToken := AccessToken{Token: uuid.NewString()}
		expectedResponse := api.SuccessResponse{
			Data:    expectedToken,
			Message: "account authenticated successfully",
		}

		body := `{"email": "jd@testmail.com", "password": "$PASSword001"}`
		mockService.EXPECT().AuthenticateAccount(gomock.Any(), expectedParam).
			Return(expectedToken, nil)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("POST", "/v1/api/accounts/authenticate", bytes.NewBufferString(body))
		resp := handler.AuthenticateAccountHandler(c)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedResponse, resp.Data)
	})

	t.Run("failed to authenticate account - missing or invalid email", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		body := `{"email": "jd@testmailcom", "password": "$PASSword001"}`

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("POST", "/v1/api/accounts/authenticate", bytes.NewBufferString(body))
		resp := handler.AuthenticateAccountHandler(c)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("failed to authenticate account - invalid credentials", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		expectedParam := AuthenticateAccountParams{
			Email:    "jd@testmail.com",
			Password: "$PASSword001",
		}
		body := `{"email": "jd@testmail.com", "password": "$PASSword001"}`

		mockService.EXPECT().AuthenticateAccount(gomock.Any(), expectedParam).
			Return(AccessToken{}, platformerrors.MakeApiError(http.StatusUnauthorized, "invalid login credentials"))

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("POST", "/v1/api/accounts/authenticate", bytes.NewBufferString(body))
		resp := handler.AuthenticateAccountHandler(c)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, resp.Error.Message, "invalid login credentials")
	})
}

func TestHandler_MeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successfully get current authenticated account profile", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		userID, accountID := uuid.New(), uuid.New()
		expectedProfile := Profile{AccountID: accountID}
		expectedResponse := api.SuccessResponse{
			Data:    expectedProfile,
			Message: "user profile retrieved successfully",
		}

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		mockService.EXPECT().GetProfile(gomock.Any(), userID).
			Return(expectedProfile, nil)

		c.Request = httptest.NewRequest("GET", "/v1/api/me", nil)
		injectClaim(c, userID, accountID)

		resp := handler.MeHandler(c)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedResponse, resp.Data)
	})

	t.Run("failure to get user - not authenticated", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/v1/api/me", nil)

		resp := handler.MeHandler(c)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}

func TestHandler_SuspendAccountHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("successfully suspend an account", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		accountID, userID := uuid.New(), uuid.New()
		expectedResponse := api.SuccessResponse{
			Message: "account suspended successfully",
		}

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		mockService.EXPECT().SuspendAccount(gomock.Any(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		}).Return(nil)
		c.Request = httptest.NewRequest("PATCH", "/v1/api/accounts/:id/suspend", nil)
		c.Params = gin.Params{{Key: "id", Value: accountID.String()}}
		injectClaim(c, userID, accountID)

		resp := handler.SuspendAccountHandler(c)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedResponse, resp.Data)
	})

	t.Run("failed to suspend account - not found", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		accountID, userID := uuid.New(), uuid.New()
		expectedResponse := api.ErrorResponse{
			Error: "account not found",
		}

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		mockService.EXPECT().SuspendAccount(gomock.Any(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		}).Return(platformerrors.MakeApiError(http.StatusNotFound, "account not found"))
		c.Request = httptest.NewRequest("PATCH", "/v1/api/accounts/:id/suspend", nil)
		c.Params = gin.Params{{Key: "id", Value: accountID.String()}}
		injectClaim(c, userID, accountID)

		resp := handler.SuspendAccountHandler(c)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, expectedResponse.Error, resp.Error.Message)
	})
}

func TestHandler_ActivateAccountHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successfully activate an account", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		accountID, userID := uuid.New(), uuid.New()
		expectedResponse := api.SuccessResponse{
			Message: "account activated successfully",
		}

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		mockService.EXPECT().ActivateAccount(gomock.Any(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		}).Return(nil)
		c.Request = httptest.NewRequest("PATCH", "/v1/api/accounts/:id/activate", nil)
		c.Params = gin.Params{{Key: "id", Value: accountID.String()}}
		injectClaim(c, userID, accountID)

		resp := handler.ActivateAccountHandler(c)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedResponse, resp.Data)
	})

	t.Run("failed to activate account - not found", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		accountID, userID := uuid.New(), uuid.New()
		expectedResponse := api.ErrorResponse{
			Error: "account not found",
		}

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		mockService.EXPECT().ActivateAccount(gomock.Any(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		}).Return(platformerrors.MakeApiError(http.StatusNotFound, "account not found"))
		c.Request = httptest.NewRequest("PATCH", "/v1/api/accounts/:id/activate", nil)
		c.Params = gin.Params{{Key: "id", Value: accountID.String()}}
		injectClaim(c, userID, accountID)

		resp := handler.ActivateAccountHandler(c)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, expectedResponse.Error, resp.Error.Message)
	})
}

func TestHandler_CloseAccountHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("successfully close an account", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		accountID, userID := uuid.New(), uuid.New()
		expectedResponse := api.SuccessResponse{
			Message: "account closed successfully",
		}

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		mockService.EXPECT().CloseAccount(gomock.Any(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		}).Return(nil)
		c.Request = httptest.NewRequest("PATCH", "/v1/api/accounts/:id/close", nil)
		c.Params = gin.Params{{Key: "id", Value: accountID.String()}}
		injectClaim(c, userID, accountID)

		resp := handler.CloseAccountHandler(c)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedResponse, resp.Data)
	})

	t.Run("failed to close account - not found", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		accountID, userID := uuid.New(), uuid.New()
		expectedResponse := api.ErrorResponse{
			Error: "account not found",
		}

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		mockService.EXPECT().CloseAccount(gomock.Any(), OperationParams{
			UserID:    userID,
			AccountID: accountID,
		}).Return(platformerrors.MakeApiError(http.StatusNotFound, "account not found"))
		c.Request = httptest.NewRequest("PATCH", "/v1/api/accounts/:id/close", nil)
		c.Params = gin.Params{{Key: "id", Value: accountID.String()}}
		injectClaim(c, userID, accountID)

		resp := handler.CloseAccountHandler(c)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, expectedResponse.Error, resp.Error.Message)
	})
}

func TestHandler_GetAccountStatusHistoryHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("successfully get account status history", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		accountID, userID := uuid.New(), uuid.New()
		expectedHistory := []ChangeHistory{
			{
				AccountID: accountID,
			},
		}
		expectedResponse := api.SuccessResponse{
			Data:    expectedHistory,
			Message: "account status history retrieved successfully",
		}

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		mockService.EXPECT().GetAccountStatusHistory(gomock.Any(), accountID).
			Return(expectedHistory, nil)
		c.Request = httptest.NewRequest("GET", "/v1/api/accounts/:id/status-history", nil)
		c.Params = gin.Params{{Key: "id", Value: accountID.String()}}
		injectClaim(c, userID, accountID)

		resp := handler.GetAccountStatusHistoryHandler(c)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedResponse, resp.Data)
	})
}

func injectClaim(ctx *gin.Context, userID, accountID uuid.UUID) {
	tokenData := generator.TokenData{
		AccountID: accountID,
		UserID:    userID,
	}
	customClaims := &generator.Claim{
		TokenData: tokenData,
	}
	claims := &validator.ValidatedClaims{
		CustomClaims: customClaims,
	}
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), jwtmiddleware.ContextKey{}, claims))
}
