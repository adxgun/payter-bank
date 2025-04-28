package interestrate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"payter-bank/internal/api"
	"payter-bank/internal/auth"
	"payter-bank/internal/database/models"
	platformerrors "payter-bank/internal/errors"
	"testing"
)

func TestHandler_CreateInterestRateHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("successfully creates interest rate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		userID, accountID := uuid.New(), uuid.New()
		rateID := uuid.New()
		profile := auth.Profile{
			UserID:    userID,
			AccountID: accountID,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := CreateInterestRateParam{
			Rate:                 5.5,
			CalculationFrequency: "monthly",
		}
		jsonBody, _ := json.Marshal(body)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/interest-rate", bytes.NewBuffer(jsonBody))
		injectProfile(c, profile)

		expectedResponse := &Response{
			InterestRateID: rateID,
		}

		mockService.EXPECT().
			CreateInterestRate(gomock.Any(), CreateInterestRateParam{
				Rate:                 5.5,
				CalculationFrequency: "monthly",
				UserID:               userID,
			}).
			Return(expectedResponse, nil)

		response := handler.CreateInterestRateHandler(c)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, api.SuccessResponse{
			Data:    expectedResponse,
			Message: "interest rate created successfully",
		}, response.Data)
	})

	t.Run("returns error when request body is invalid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/interest-rate", bytes.NewBuffer([]byte(`{invalid json}`)))

		response := handler.CreateInterestRateHandler(c)

		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("returns error when auth profile is missing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := CreateInterestRateParam{
			Rate:                 5.5,
			CalculationFrequency: "monthly",
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/interest-rate", bytes.NewBuffer(jsonBody))

		response := handler.CreateInterestRateHandler(c)

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Contains(t, response.Error.Message, "unauthorized")
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		userID := uuid.New()
		accountID := uuid.New()
		profile := auth.Profile{
			UserID:    userID,
			AccountID: accountID,
		}

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := CreateInterestRateParam{
			Rate:                 5.5,
			CalculationFrequency: "monthly",
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/interest-rate", bytes.NewBuffer(jsonBody))
		injectProfile(c, profile)

		mockService.EXPECT().
			CreateInterestRate(gomock.Any(), CreateInterestRateParam{
				Rate:                 5.5,
				CalculationFrequency: "monthly",
				UserID:               userID,
			}).
			Return(nil, platformerrors.ErrInternal)

		response := handler.CreateInterestRateHandler(c)

		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Contains(t, response.Error.Message, platformerrors.ErrInternal.Error())
	})

	t.Run("validates request parameters", func(t *testing.T) {
		testCases := []struct {
			name  string
			param CreateInterestRateParam
		}{
			{
				name: "invalid rate",
				param: CreateInterestRateParam{
					Rate:                 -1.0,
					CalculationFrequency: "MONTHLY",
				},
			},
			{
				name: "invalid calculation frequency",
				param: CreateInterestRateParam{
					Rate:                 5.5,
					CalculationFrequency: "INVALID",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mockService := NewMockService(ctrl)
				handler := NewHandler(mockService)

				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				c.Set(auth.ProfileKey, &auth.Profile{
					UserID: uuid.New(),
				})

				jsonBody, _ := json.Marshal(tc.param)
				c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
				c.Request.Header.Set("Content-Type", "application/json")

				response := handler.CreateInterestRateHandler(c)

				assert.Equal(t, http.StatusBadRequest, response.Code)
			})
		}
	})
}

func TestHandler_UpdateRateHandler(t *testing.T) {
	t.Run("successfully updates interest rate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		userID := uuid.New()
		rateID := uuid.New()

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		profile := auth.Profile{
			UserID: userID,
		}

		reqBody := UpdateRateParam{
			Rate: 6.5,
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/v1/api/interest-rate", bytes.NewBuffer(jsonBody))
		injectProfile(c, profile)

		expectedResponse := &Response{
			InterestRateID: rateID,
		}

		mockService.EXPECT().
			UpdateRate(gomock.Any(), UpdateRateParam{
				Rate:   6.5,
				UserID: userID,
			}).
			Return(expectedResponse, nil)

		// Execute
		response := handler.UpdateRateHandler(c)

		// Verify
		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, api.SuccessResponse{
			Data:    expectedResponse,
			Message: "interest rate updated successfully",
		}, response.Data)
	})

	t.Run("returns error when request body is invalid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(`{invalid json}`)))
		c.Request.Header.Set("Content-Type", "application/json")

		response := handler.UpdateRateHandler(c)

		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("returns error when auth profile is missing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := UpdateRateParam{
			Rate: 6.5,
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(jsonBody))

		response := handler.UpdateRateHandler(c)

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Contains(t, response.Error.Message, "unauthorized")
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		userID := uuid.New()

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := UpdateRateParam{
			Rate: 6.5,
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(jsonBody))
		injectProfile(c, auth.Profile{UserID: userID})

		mockService.EXPECT().
			UpdateRate(gomock.Any(), UpdateRateParam{
				Rate:   6.5,
				UserID: userID,
			}).
			Return(nil, platformerrors.ErrInternal)

		response := handler.UpdateRateHandler(c)

		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Contains(t, response.Error.Message, platformerrors.ErrInternal.Error())
	})

	t.Run("validates rate parameter", func(t *testing.T) {
		testCases := []struct {
			name string
			rate float64
		}{
			{
				name: "zero rate",
				rate: 0.0,
			},
			{
				name: "negative rate",
				rate: -1.0,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mockService := NewMockService(ctrl)
				handler := NewHandler(mockService)

				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				reqBody := UpdateRateParam{
					Rate: tc.rate,
				}
				jsonBody, _ := json.Marshal(reqBody)
				c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(jsonBody))
				injectProfile(c, auth.Profile{UserID: uuid.New()})

				response := handler.UpdateRateHandler(c)

				assert.Equal(t, http.StatusBadRequest, response.Code)
			})
		}
	})
}

func TestHandler_UpdateCalculationFrequencyHandler(t *testing.T) {
	t.Run("successfully updates calculation frequency", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		userID := uuid.New()
		rateID := uuid.New()

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := UpdateCalculationFrequencyParam{
			CalculationFrequency: "weekly",
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/v1/api/interest-rate/calculation-frequency", bytes.NewBuffer(jsonBody))
		profile := auth.Profile{
			UserID: userID,
		}
		injectProfile(c, profile)

		expectedResponse := &Response{
			InterestRateID: rateID,
		}

		mockService.EXPECT().
			UpdateCalculationFrequency(gomock.Any(), UpdateCalculationFrequencyParam{
				CalculationFrequency: "weekly",
				UserID:               userID,
			}).
			Return(expectedResponse, nil)

		response := handler.UpdateCalculationFrequencyHandler(c)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, api.SuccessResponse{
			Data:    expectedResponse,
			Message: "calculation frequency updated successfully",
		}, response.Data)
	})

	t.Run("returns error when request body is invalid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPut, "/v1/api/interest-rate/calculation-frequency", bytes.NewBuffer([]byte(`{invalid json}`)))
		c.Request.Header.Set("Content-Type", "application/json")

		response := handler.UpdateCalculationFrequencyHandler(c)

		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("returns error when auth profile is missing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := UpdateCalculationFrequencyParam{
			CalculationFrequency: "weekly",
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/v1/api/interest-rate/calculation-frequency", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		response := handler.UpdateCalculationFrequencyHandler(c)

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Contains(t, response.Error.Message, "unauthorized")
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		userID := uuid.New()

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := UpdateCalculationFrequencyParam{
			CalculationFrequency: "weekly",
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/v1/api/interest-rate/calculation-frequency", bytes.NewBuffer(jsonBody))
		injectProfile(c, auth.Profile{UserID: userID})

		mockService.EXPECT().
			UpdateCalculationFrequency(gomock.Any(), UpdateCalculationFrequencyParam{
				CalculationFrequency: "weekly",
				UserID:               userID,
			}).
			Return(nil, platformerrors.ErrInternal)

		response := handler.UpdateCalculationFrequencyHandler(c)

		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Contains(t, response.Error.Message, platformerrors.ErrInternal.Error())
	})

	t.Run("validates calculation frequency parameter", func(t *testing.T) {
		testCases := []struct {
			name                 string
			calculationFrequency string
			expectedError        string
		}{
			{
				name:                 "empty frequency",
				calculationFrequency: "",
			},
			{
				name:                 "invalid frequency",
				calculationFrequency: "invalid",
			},
			{
				name:                 "wrong case",
				calculationFrequency: "WEEKLY",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mockService := NewMockService(ctrl)
				handler := NewHandler(mockService)

				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				reqBody := UpdateCalculationFrequencyParam{
					CalculationFrequency: tc.calculationFrequency,
				}
				jsonBody, _ := json.Marshal(reqBody)
				c.Request = httptest.NewRequest(http.MethodPut, "/v1/api/interest-rate/calculation-frequency", bytes.NewBuffer(jsonBody))
				injectProfile(c, auth.Profile{UserID: uuid.New()})

				response := handler.UpdateCalculationFrequencyHandler(c)

				assert.Equal(t, http.StatusBadRequest, response.Code)
			})
		}
	})

	t.Run("validates supported frequencies", func(t *testing.T) {
		validFrequencies := []string{"hourly", "daily", "weekly", "monthly", "yearly"}

		for _, freq := range validFrequencies {
			t.Run(fmt.Sprintf("accepts %s frequency", freq), func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mockService := NewMockService(ctrl)
				handler := NewHandler(mockService)

				userID := uuid.New()
				rateID := uuid.New()

				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				reqBody := UpdateCalculationFrequencyParam{
					CalculationFrequency: freq,
				}
				jsonBody, _ := json.Marshal(reqBody)
				c.Request = httptest.NewRequest(http.MethodPut, "/v1/api/interest-rate/calculation-frequency", bytes.NewBuffer(jsonBody))
				injectProfile(c, auth.Profile{UserID: userID})

				expectedResponse := &Response{
					InterestRateID: rateID,
				}

				mockService.EXPECT().
					UpdateCalculationFrequency(gomock.Any(), UpdateCalculationFrequencyParam{
						CalculationFrequency: freq,
						UserID:               userID,
					}).
					Return(expectedResponse, nil)

				response := handler.UpdateCalculationFrequencyHandler(c)

				assert.Equal(t, http.StatusOK, response.Code)
			})
		}
	})
}

func TestHandler_GetCurrentRateHandler(t *testing.T) {
	t.Run("successfully gets current interest rate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/v1/api/interest-rate/current", nil)

		expectedRate := &models.InterestRate{
			ID:                   uuid.New(),
			Rate:                 550, // 5.5%
			CalculationFrequency: "monthly",
		}

		mockService.EXPECT().
			GetCurrentRate(gomock.Any()).
			Return(expectedRate, nil)

		response := handler.GetCurrentRateHandler(c)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, api.SuccessResponse{
			Data:    expectedRate,
			Message: "current interest rate retrieved successfully",
		}, response.Data)
	})

	t.Run("returns error when rate is not initialized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/v1/api/interest-rate/current", nil)

		expectedError := platformerrors.MakeApiError(
			http.StatusPreconditionFailed,
			"interest rate has not been initialized",
		)

		mockService.EXPECT().
			GetCurrentRate(gomock.Any()).
			Return(nil, expectedError)

		response := handler.GetCurrentRateHandler(c)

		assert.Equal(t, http.StatusPreconditionFailed, response.Code)
		assert.Contains(t, response.Error.Message, "interest rate has not been initialized")
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		handler := NewHandler(mockService)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/v1/api/interest-rate/current", nil)

		mockService.EXPECT().
			GetCurrentRate(gomock.Any()).
			Return(nil, platformerrors.ErrInternal)

		response := handler.GetCurrentRateHandler(c)

		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Contains(t, response.Error.Message, platformerrors.ErrInternal.Error())
	})

	t.Run("handles different rate configurations", func(t *testing.T) {
		testCases := []struct {
			name                 string
			rate                 int64
			calculationFrequency string
		}{
			{
				name:                 "high rate monthly",
				rate:                 2000,
				calculationFrequency: "monthly",
			},
			{
				name:                 "low rate daily",
				rate:                 100,
				calculationFrequency: "daily",
			},
			{
				name:                 "zero rate yearly",
				rate:                 0,
				calculationFrequency: "yearly",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mockService := NewMockService(ctrl)
				handler := NewHandler(mockService)

				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				c.Request = httptest.NewRequest(http.MethodGet, "/v1/api/interest-rate/current", nil)

				expectedRate := &models.InterestRate{
					ID:                   uuid.New(),
					Rate:                 tc.rate,
					CalculationFrequency: tc.calculationFrequency,
				}

				mockService.EXPECT().
					GetCurrentRate(gomock.Any()).
					Return(expectedRate, nil)

				response := handler.GetCurrentRateHandler(c)

				assert.Equal(t, http.StatusOK, response.Code)
			})
		}
	})
}

func injectProfile(ctx *gin.Context, profile auth.Profile) {
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), auth.ProfileKey, profile))
}
