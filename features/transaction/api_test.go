package transaction

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"payter-bank/internal/api"
	"payter-bank/internal/auth"
	platformerrors "payter-bank/internal/errors"
	"testing"
)

func TestHandler_CreditAccountHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("credit account successfully", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		fromAccountID, toAccountID, userID := uuid.New(), uuid.New(), uuid.New()
		expectedParam := AccountTransactionParams{
			FromAccountID: fromAccountID,
			ToAccountID:   toAccountID,
			Amount:        100,
			Narration:     "Spending money for dinner",
			UserID:        userID,
		}

		authProfile := auth.Profile{AccountID: uuid.New(), UserID: userID}
		response := Response{
			TransactionID: uuid.New(),
		}

		body, _ := json.Marshal(expectedParam)
		mockService.EXPECT().CreditAccount(gomock.Any(), expectedParam).
			Return(&response, nil)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/api/credit", bytes.NewBuffer(body))
		injectProfile(c, authProfile)

		resp := handler.CreditAccountHandler(c)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, resp.Data, api.SuccessResponse{
			Data:    &response,
			Message: "account credited successfully",
		})
	})

	t.Run("failed to credit account - missing amount", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		fromAccountID := uuid.MustParse("1938dc36-aef5-4ef9-b0ae-1bb08b2ccbab")
		toAccountID := uuid.MustParse("824312b8-ec3c-467a-8c84-8d14a2f2fc76")
		userID := uuid.MustParse("1938dc36-aef5-4ef9-b0ae-1bb08b2ccbab")

		body := `{"from_account_id": "` + fromAccountID.String() + `", "to_account_id": "` + toAccountID.String() + `", "narration": "Spending money for dinner", "user_id": "` + userID.String() + `"}`
		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/api/credit", bytes.NewBufferString(body))

		resp := handler.CreditAccountHandler(c)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("failed to credit account - missing from_account_id", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		toAccountID := uuid.MustParse("824312b8-ec3c-467a-8c84-8d14a2f2fc76")
		userID := uuid.MustParse("1938dc36-aef5-4ef9-b0ae-1bb08b2ccbab")
		profile := auth.Profile{AccountID: uuid.New(), UserID: userID}
		expectedParam := AccountTransactionParams{
			ToAccountID: toAccountID,
			Amount:      100,
			Narration:   "Spending money for dinner",
			UserID:      userID,
		}

		mockService.EXPECT().CreditAccount(gomock.Any(), expectedParam).
			Return(nil, platformerrors.MakeApiError(http.StatusNotFound, "account not found"))

		body := `{"to_account_id": "` + toAccountID.String() + `", "amount":100, "narration": "Spending money for dinner", "user_id": "` + userID.String() + `"}`
		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/api/credit", bytes.NewBufferString(body))
		injectProfile(c, profile)

		resp := handler.CreditAccountHandler(c)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, resp.Error.Message, "account not found")
	})
}

func TestHandler_DebitAccountHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("debit account successfully", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		fromAccountID, toAccountID, userID := uuid.New(), uuid.New(), uuid.New()
		expectedParam := AccountTransactionParams{
			FromAccountID: fromAccountID,
			ToAccountID:   toAccountID,
			Amount:        100,
			Narration:     "Spending money for dinner",
			UserID:        userID,
		}
		authProfile := auth.Profile{AccountID: uuid.New(), UserID: userID}
		response := Response{
			TransactionID: uuid.New(),
		}
		body, _ := json.Marshal(expectedParam)
		mockService.EXPECT().DebitAccount(gomock.Any(), expectedParam).
			Return(&response, nil)
		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/api/debit", bytes.NewBuffer(body))
		injectProfile(c, authProfile)

		resp := handler.DebitAccountHandler(c)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, resp.Data, api.SuccessResponse{
			Data:    &response,
			Message: "account debited successfully",
		})
	})

	t.Run("failed to debit account - missing amount", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		fromAccountID := uuid.MustParse("1938dc36-aef5-4ef9-b0ae-1bb08b2ccbab")
		toAccountID := uuid.MustParse("824312b8-ec3c-467a-8c84-8d14a2f2fc76")
		userID := uuid.MustParse("1938dc36-aef5-4ef9-b0ae-1bb08b2ccbab")
		body := `{"from_account_id": "` + fromAccountID.String() + `", "to_account_id": "` + toAccountID.String() + `", "narration": "Spending money for dinner", "user_id": "` + userID.String() + `"}`
		handler := NewHandler(mockService)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/api/debit", bytes.NewBufferString(body))
		resp := handler.DebitAccountHandler(c)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("failed to debit account - missing from_account_id", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		toAccountID := uuid.MustParse("824312b8-ec3c-467a-8c84-8d14a2f2fc76")
		userID := uuid.MustParse("1938dc36-aef5-4ef9-b0ae-1bb08b2ccbab")
		profile := auth.Profile{AccountID: uuid.New(), UserID: userID}
		expectedParam := AccountTransactionParams{
			ToAccountID: toAccountID,
			Amount:      100,
			Narration:   "Spending money for dinner",
			UserID:      userID,
		}
		mockService.EXPECT().DebitAccount(gomock.Any(), expectedParam).
			Return(nil, platformerrors.MakeApiError(http.StatusNotFound, "account not found"))
		body := `{"to_account_id": "` + toAccountID.String() + `", "amount":100, "narration": "Spending money for dinner", "user_id": "` + userID.String() + `"}`
		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/api/debit", bytes.NewBufferString(body))
		injectProfile(c, profile)
		resp := handler.DebitAccountHandler(c)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, resp.Error.Message, "account not found")
	})

	t.Run("service internal error", func(t *testing.T) {
		mockService := NewMockService(gomock.NewController(t))
		fromAccountID, toAccountID, userID := uuid.New(), uuid.New(), uuid.New()
		expectedParam := AccountTransactionParams{
			FromAccountID: fromAccountID,
			ToAccountID:   toAccountID,
			Amount:        100,
			Narration:     "Spending money for dinner",
			UserID:        userID,
		}
		authProfile := auth.Profile{AccountID: uuid.New(), UserID: userID}
		body, _ := json.Marshal(expectedParam)
		mockService.EXPECT().DebitAccount(gomock.Any(), expectedParam).
			Return(nil, platformerrors.ErrInternal)
		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/api/debit", bytes.NewBuffer(body))
		injectProfile(c, authProfile)

		resp := handler.DebitAccountHandler(c)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

func TestTransferFundsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("successfully transfers funds", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)
		userID, accountID := uuid.New(), uuid.New()
		profile := auth.Profile{
			UserID:    userID,
			AccountID: accountID,
		}

		req := AccountTransactionParams{
			FromAccountID: accountID,
			ToAccountID:   uuid.New(),
			Amount:        100.50,
			Narration:     "Test transfer",
			UserID:        userID,
		}

		expectedResponse := &Response{
			TransactionID: uuid.New(),
		}

		mockService.EXPECT().
			Transfer(gomock.Any(), req).
			Return(expectedResponse, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/api/transfer", bytes.NewBuffer(body))
		injectProfile(c, profile)

		resp := handler.TransferFundsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, api.SuccessResponse{
			Data:    expectedResponse,
			Message: "transaction successful",
		}, resp.Data)
	})

	t.Run("fails when user is not authenticated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		req := AccountTransactionParams{
			FromAccountID: uuid.New(),
			ToAccountID:   uuid.New(),
			Amount:        100.50,
			Narration:     "Test transfer",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/api/transfer", bytes.NewBuffer(body))

		resp := handler.TransferFundsHandler(c)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("fails with invalid request body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		invalidBody := `{"amount": "invalid"}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/api/transfer", bytes.NewBufferString(invalidBody))

		c.Set(auth.ProfileKey, auth.Profile{
			UserID: uuid.New(),
		})

		resp := handler.TransferFundsHandler(c)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("fails with missing required fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		req := AccountTransactionParams{
			FromAccountID: uuid.New(),
			ToAccountID:   uuid.New(),
			Narration:     "Test transfer",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/api/transfer", bytes.NewBuffer(body))

		c.Set(auth.ProfileKey, auth.Profile{
			UserID: uuid.New(),
		})

		resp := handler.TransferFundsHandler(c)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("fails when service returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)
		userID, accountID := uuid.New(), uuid.New()
		profile := auth.Profile{
			UserID:    userID,
			AccountID: accountID,
		}

		req := AccountTransactionParams{
			FromAccountID: accountID,
			ToAccountID:   uuid.New(),
			Amount:        100.50,
			Narration:     "Test transfer",
			UserID:        userID,
		}

		mockService.EXPECT().
			Transfer(gomock.Any(), req).
			Return(nil, platformerrors.MakeApiError(http.StatusPreconditionFailed, "insufficient funds"))

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/api/transfer", bytes.NewBuffer(body))
		injectProfile(c, profile)

		resp := handler.TransferFundsHandler(c)
		assert.Equal(t, http.StatusPreconditionFailed, resp.Code)
		assert.Contains(t, resp.Error.Message, "insufficient funds")
	})
}

func TestHandler_BalanceHandler(t *testing.T) {
	t.Run("successfully gets balance for own account", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		accountID, userID := uuid.New(), uuid.New()
		expectedBalance := Balance{
			AccountID:     accountID,
			Balance:       1000.50,
			AccountNumber: "1234567890",
			AccountType:   "CURRENT",
			Currency:      "GBP",
		}
		profile := auth.Profile{
			AccountID: accountID,
			UserID:    userID,
		}

		mockService.EXPECT().
			GetAccountBalance(gomock.Any(), accountID).
			Return(expectedBalance, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: make(http.Header)}
		c.Params = []gin.Param{{Key: "id", Value: accountID.String()}}
		injectProfile(c, profile)

		response := handler.BalanceHandler(c)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, api.SuccessResponse{
			Data:    expectedBalance,
			Message: "account balance retrieved successfully",
		}, response.Data)
	})

	t.Run("successfully gets balance for any account as admin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		adminAccountID := uuid.New()
		targetAccountID := uuid.New()
		expectedBalance := Balance{
			AccountID:     targetAccountID,
			Balance:       1000.50,
			AccountNumber: "1234567890",
			AccountType:   "CURRENT",
			Currency:      "GBP",
		}
		profile := auth.Profile{
			UserID:    uuid.New(),
			AccountID: adminAccountID,
			UserType:  "ADMIN",
		}

		mockService.EXPECT().
			GetAccountBalance(gomock.Any(), targetAccountID).
			Return(expectedBalance, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: targetAccountID.String()}}
		c.Request = &http.Request{Header: make(http.Header)}
		injectProfile(c, profile)

		response := handler.BalanceHandler(c)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, api.SuccessResponse{
			Data:    expectedBalance,
			Message: "account balance retrieved successfully",
		}, response.Data)
	})

	t.Run("fails with invalid account ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "invalid-uuid"}}

		response := handler.BalanceHandler(c)

		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("fails when unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		accountID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: accountID.String()}}
		c.Request = &http.Request{Header: make(http.Header)}
		// Don't set profile in context

		response := handler.BalanceHandler(c)

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Contains(t, response.Error.Message, "unauthorized")
	})

	t.Run("fails when service returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		accountID := uuid.New()

		mockService.EXPECT().
			GetAccountBalance(gomock.Any(), accountID).
			Return(Balance{}, platformerrors.ErrInternal)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: accountID.String()}}
		c.Request = &http.Request{Header: make(http.Header)}
		injectProfile(c, auth.Profile{
			AccountID: accountID,
			UserID:    uuid.New(),
			UserType:  "ADMIN",
		})

		response := handler.BalanceHandler(c)

		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})
}

func TestHandler_GetTransactionHistoryHandler(t *testing.T) {
	t.Run("successfully gets transaction history for own account", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		accountID, transactionID, accountID1 := uuid.New(), uuid.New(), uuid.New()
		profile := auth.Profile{
			AccountID: accountID,
			UserID:    uuid.New(),
		}

		expectedTransactions := []Transaction{
			{
				TransactionID:   transactionID,
				FromAccountID:   accountID,
				ToAccountID:     accountID1,
				Amount:          Amount{Amount: 100.50, Currency: "GBP"},
				ReferenceNumber: "TRX123456",
				Description:     "Test transaction 1",
				Status:          "COMPLETED",
				Currency:        "GBP",
			},
		}

		mockService.EXPECT().
			GetTransactionHistory(gomock.Any(), accountID).
			Return(expectedTransactions, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: accountID.String()}}

		c.Request = &http.Request{Header: make(http.Header)}
		injectProfile(c, profile)

		response := handler.GetTransactionHistoryHandler(c)

		assert.Equal(t, http.StatusOK, response.Code)

		assert.Equal(t, api.SuccessResponse{
			Data:    expectedTransactions,
			Message: "transaction history retrieved successfully",
		}, response.Data)
	})

	t.Run("successfully gets transaction history as admin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)
		targetAccountID, transactionID, accountID1 := uuid.New(), uuid.New(), uuid.New()
		profile := auth.Profile{
			AccountID: uuid.New(),
			UserID:    uuid.New(),
			UserType:  "ADMIN",
		}

		expectedTransactions := []Transaction{
			{
				TransactionID:   transactionID,
				FromAccountID:   targetAccountID,
				ToAccountID:     accountID1,
				Amount:          Amount{Amount: 100.50, Currency: "GBP"},
				ReferenceNumber: "TRX123456",
				Description:     "Test transaction 1",
				Status:          "COMPLETED",
				Currency:        "GBP",
			},
		}

		mockService.EXPECT().
			GetTransactionHistory(gomock.Any(), targetAccountID).
			Return(expectedTransactions, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: targetAccountID.String()}}

		c.Request = &http.Request{Header: make(http.Header)}
		injectProfile(c, profile)

		response := handler.GetTransactionHistoryHandler(c)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, api.SuccessResponse{
			Data:    expectedTransactions,
			Message: "transaction history retrieved successfully",
		}, response.Data)
	})

	t.Run("fails with invalid account ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "invalid-uuid"}}
		c.Request = &http.Request{Header: make(http.Header)}

		response := handler.GetTransactionHistoryHandler(c)
		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("fails when unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		accountID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: accountID.String()}}
		c.Request = &http.Request{Header: make(http.Header)}
		// Don't set profile in context

		response := handler.GetTransactionHistoryHandler(c)

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Contains(t, response.Error.Message, "unauthorized")
	})

	t.Run("fails when customer tries to access different account", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		userAccountID := uuid.New()
		differentAccountID := uuid.New()
		profile := auth.Profile{
			AccountID: userAccountID,
			UserID:    uuid.New(),
			UserType:  "CUSTOMER",
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: differentAccountID.String()}}
		c.Request = &http.Request{Header: make(http.Header)}
		injectProfile(c, profile)

		response := handler.GetTransactionHistoryHandler(c)

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Contains(t, response.Error.Message, "you are not authorized to view this account's transactions")
	})

	t.Run("successfully returns empty transaction history", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		accountID := uuid.New()
		profile := auth.Profile{
			AccountID: accountID,
			UserID:    uuid.New(),
			UserType:  "CUSTOMER",
		}

		mockService.EXPECT().
			GetTransactionHistory(gomock.Any(), accountID).
			Return([]Transaction{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: accountID.String()}}

		c.Request = &http.Request{Header: make(http.Header)}
		injectProfile(c, profile)

		response := handler.GetTransactionHistoryHandler(c)
		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, api.SuccessResponse{
			Data:    []Transaction{},
			Message: "transaction history retrieved successfully",
		}, response.Data)
	})
}

func injectProfile(ctx *gin.Context, profile auth.Profile) {
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), auth.ProfileKey, profile))
}
