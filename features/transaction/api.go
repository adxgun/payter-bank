package transaction

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

// CreditAccountHandler godoc
// @Summary      Credit an account
// @Description  Credit an account with a specific amount - this endpoint can only be used by the admin. The originating account will be assumed to be an external account.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        account  body  AccountTransactionParams  true  "credit transaction params"
// @Success      200  {object}  api.SuccessResponse{data=Response}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/credit [post]
func (h *Handler) CreditAccountHandler(ctx *gin.Context) api.Response {
	var params AccountTransactionParams
	if err := ctx.ShouldBindJSON(&params); err != nil {
		return api.BadRequest(err.Error())
	}

	profile, err := auth.GetCurrentProfile(ctx)
	if err != nil {
		return api.Unauthorized("unauthorized")
	}

	params.UserID = profile.UserID
	resp, err := h.service.CreditAccount(ctx, params)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("account credited successfully", resp)
}

// DebitAccountHandler godoc
// @Summary      Debit an account
// @Description  Debit an account with a specific amount - this endpoint can only be used by the admin. The destination account will be assumed to be an external account.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        account  body  AccountTransactionParams  true  "account transaction params"
// @Success      200  {object}  api.SuccessResponse{data=Response}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/debit [post]
func (h *Handler) DebitAccountHandler(ctx *gin.Context) api.Response {
	var params AccountTransactionParams
	if err := ctx.ShouldBindJSON(&params); err != nil {
		return api.BadRequest(err.Error())
	}

	profile, err := auth.GetCurrentProfile(ctx)
	if err != nil {
		return api.Unauthorized("unauthorized")
	}

	params.UserID = profile.UserID
	resp, err := h.service.DebitAccount(ctx, params)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("account debited successfully", resp)
}

// TransferFundsHandler godoc
// @Summary      Transfer from one account to account.
// @Description  Transfer from one account to another account.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        account  body  AccountTransactionParams  true  "credit account params"
// @Success      200  {object}  api.SuccessResponse{data=Response}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/transfer [post]
func (h *Handler) TransferFundsHandler(ctx *gin.Context) api.Response {
	var params AccountTransactionParams
	if err := ctx.ShouldBindJSON(&params); err != nil {
		return api.BadRequest(err.Error())
	}

	profile, err := auth.GetCurrentProfile(ctx)
	if err != nil {
		return api.Unauthorized("unauthorized")
	}

	if profile.AccountID != params.FromAccountID {
		return api.PreConditionFailed("you do not have permission to transfer funds from this account")
	}

	params.UserID = profile.UserID
	resp, err := h.service.Transfer(ctx, params)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("transaction successful", resp)
}

// BalanceHandler godoc
// @Summary      Get account balance.
// @Description  Get account balance for the specified account.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Success      200  {object}  api.SuccessResponse{data=Balance}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/accounts/:id/balance [get]
func (h *Handler) BalanceHandler(ctx *gin.Context) api.Response {
	accountID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return api.BadRequest("account ID is required")
	}

	profile, err := auth.GetCurrentProfile(ctx)
	if err != nil {
		return api.Unauthorized("unauthorized")
	}

	if profile.UserType != "ADMIN" && profile.AccountID != accountID {
		return api.PreConditionFailed("you are not authorized to view this account balance")
	}

	balance, err := h.service.GetAccountBalance(ctx, accountID)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("account balance retrieved successfully", balance)
}

// GetTransactionHistoryHandler godoc
// @Summary      Get account transaction history.
// @Description  Get account transaction history.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Success      200  {object}  api.SuccessResponse{data=[]Transaction}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/accounts/:id/transactions [get]
func (h *Handler) GetTransactionHistoryHandler(ctx *gin.Context) api.Response {
	accountID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return api.BadRequest("account ID is required")
	}

	profile, err := auth.GetCurrentProfile(ctx)
	if err != nil {
		return api.Unauthorized("unauthorized")
	}

	if profile.UserType != "ADMIN" && profile.AccountID != accountID {
		return api.Unauthorized("you are not authorized to view this account's transactions")
	}

	data, err := h.service.GetTransactionHistory(ctx, accountID)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("transaction history retrieved successfully", data)
}
