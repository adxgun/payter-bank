package account

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payter-bank/internal/api"
	"payter-bank/internal/auth"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateAccountHandler godoc
// @Summary      Create account
// @Description  Create a new CUSTOMER or ADMIN account
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        account  body  CreateAccountParams  true  "Create account params"
// @Success      200  {object}  api.SuccessResponse{data=Profile}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/accounts [post]
func (h *Handler) CreateAccountHandler(ctx *gin.Context) api.Response {
	var params CreateAccountParams
	if err := ctx.ShouldBindJSON(&params); err != nil {
		return api.BadRequest(err.Error())
	}

	profile, err := h.service.CreateAccount(ctx, params)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("account created successfully", profile)
}

// AuthenticateAccountHandler godoc
// @Summary      Authenticate account
// @Description  Authenticate an account using email and password
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        account  body  AuthenticateAccountParams  true  "authenticate account params"
// @Success      200  {object}  api.SuccessResponse{data=AccessToken}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/accounts/authenticate [post]
func (h *Handler) AuthenticateAccountHandler(ctx *gin.Context) api.Response {
	var params AuthenticateAccountParams
	if err := ctx.ShouldBindJSON(&params); err != nil {
		return api.BadRequest(err.Error())
	}

	token, err := h.service.AuthenticateAccount(ctx, params)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("account authenticated successfully", token)
}

// MeHandler godoc
// @Summary      Get current user
// @Description  return the current authenticated user
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Success      200  {object}  api.SuccessResponse{data=Profile}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/me [get]
func (h *Handler) MeHandler(ctx *gin.Context) api.Response {
	token, err := auth.GetTokenData(ctx)
	if err != nil {
		return api.Unauthorized(err.Error())
	}

	userProfile, err := h.service.GetProfile(ctx, token.UserID)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("user profile retrieved successfully", userProfile)
}

// SuspendAccountHandler godoc
// @Summary      Suspend account
// @Description  Suspend an account - this will set the account status to SUSPENDED and this can only be done by an admin
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Success      200  {object}  api.SuccessResponse
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/accounts/:id/suspend [patch]
func (h *Handler) SuspendAccountHandler(ctx *gin.Context) api.Response {
	accountID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return api.BadRequest("account id is required")
	}

	profile, err := auth.GetTokenData(ctx)
	if err != nil {
		return api.Unauthorized(err.Error())
	}

	err = h.service.SuspendAccount(ctx, OperationParams{
		UserID:    profile.UserID,
		AccountID: accountID,
	})
	if err != nil {
		return api.Error(err)
	}

	return api.OK("account suspended successfully", nil)
}

// ActivateAccountHandler godoc
// @Summary      Activate account
// @Description  Activate an account - this will set the account status to ACTIVE and this can only be done by an admin
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Success      200  {object}  api.SuccessResponse
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/accounts/:id/activate [patch]
func (h *Handler) ActivateAccountHandler(ctx *gin.Context) api.Response {
	accountID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return api.BadRequest("account id is required")
	}

	profile, err := auth.GetTokenData(ctx)
	if err != nil {
		return api.Unauthorized(err.Error())
	}

	err = h.service.ActivateAccount(ctx, OperationParams{
		UserID:    profile.UserID,
		AccountID: accountID,
	})
	if err != nil {
		return api.Error(err)
	}

	return api.OK("account activated successfully", nil)
}

// CloseAccountHandler godoc
// @Summary      Close account
// @Description  Close an account - this will set the account status to CLOSED and this can only be done by an admin
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Success      200  {object}  api.SuccessResponse
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/accounts/:id/close [patch]
func (h *Handler) CloseAccountHandler(ctx *gin.Context) api.Response {
	accountID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return api.BadRequest("account id is required")
	}

	profile, err := auth.GetTokenData(ctx)
	if err != nil {
		return api.Unauthorized(err.Error())
	}

	err = h.service.CloseAccount(ctx, OperationParams{
		UserID:    profile.UserID,
		AccountID: accountID,
	})
	if err != nil {
		return api.Error(err)
	}

	return api.OK("account closed successfully", nil)
}

// GetAccountStatusHistoryHandler godoc
// @Summary      Get account status history
// @Description  return the audit history of the account status
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Success      200  {object}  api.SuccessResponse{data=[]ChangeHistory}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/accounts/:id/status-history [get]
func (h *Handler) GetAccountStatusHistoryHandler(ctx *gin.Context) api.Response {
	accountID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return api.BadRequest("account id is required")
	}

	history, err := h.service.GetAccountStatusHistory(ctx, accountID)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("account status history retrieved successfully", history)
}
