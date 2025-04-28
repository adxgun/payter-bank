package interestrate

import (
	"github.com/gin-gonic/gin"
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

// CreateInterestRateHandler godoc
// @Summary      Create interest rate
// @Description  Create a new interest rate
// @Tags         interest-rate
// @Accept       json
// @Produce      json
// @Param        interest_rate  body  CreateInterestRateParam  true  "Create interest rate params"
// @Success      200  {object}  api.SuccessResponse{data=Response}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/interest-rate [post]
func (h *Handler) CreateInterestRateHandler(ctx *gin.Context) api.Response {
	var param CreateInterestRateParam
	if err := ctx.ShouldBindJSON(&param); err != nil {
		return api.BadRequest(err.Error())
	}

	profile, err := auth.GetCurrentProfile(ctx)
	if err != nil {
		return api.Unauthorized("unauthorized")
	}

	param.UserID = profile.UserID
	response, err := h.service.CreateInterestRate(ctx, param)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("interest rate created successfully", response)
}

// UpdateRateHandler godoc
// @Summary      Update interest rate
// @Description  Update an existing interest rate
// @Tags         interest-rate
// @Accept       json
// @Produce      json
// @Param        interest_rate  body  UpdateRateParam  true  "Update interest rate params"
// @Success      200  {object}  api.SuccessResponse{data=Response}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/interest-rate [put]
func (h *Handler) UpdateRateHandler(ctx *gin.Context) api.Response {
	var param UpdateRateParam
	if err := ctx.ShouldBindJSON(&param); err != nil {
		return api.BadRequest(err.Error())
	}

	profile, err := auth.GetCurrentProfile(ctx)
	if err != nil {
		return api.Unauthorized("unauthorized")
	}

	param.UserID = profile.UserID
	response, err := h.service.UpdateRate(ctx, param)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("interest rate updated successfully", response)
}

// UpdateCalculationFrequencyHandler godoc
// @Summary      Update calculation frequency
// @Description  Update the calculation frequency of an existing interest rate
// @Tags         interest-rate
// @Accept       json
// @Produce      json
// @Param        calculation_frequency  body  UpdateCalculationFrequencyParam  true  "Update calculation frequency params"
// @Success      200  {object}  api.SuccessResponse{data=Response}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/interest-rate/calculation-frequency [put]
func (h *Handler) UpdateCalculationFrequencyHandler(ctx *gin.Context) api.Response {
	var param UpdateCalculationFrequencyParam
	if err := ctx.ShouldBindJSON(&param); err != nil {
		return api.BadRequest(err.Error())
	}

	profile, err := auth.GetCurrentProfile(ctx)
	if err != nil {
		return api.Unauthorized("unauthorized")
	}

	param.UserID = profile.UserID
	response, err := h.service.UpdateCalculationFrequency(ctx, param)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("calculation frequency updated successfully", response)
}

// GetCurrentRateHandler godoc
// @Summary      Get current interest rate
// @Description  Get the current interest rate
// @Tags         interest-rate
// @Accept       json
// @Produce      json
// @Success      200  {object}  api.SuccessResponse{data=Response}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/interest-rate/current [get]
func (h *Handler) GetCurrentRateHandler(ctx *gin.Context) api.Response {
	response, err := h.service.GetCurrentRate(ctx)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("current interest rate retrieved successfully", response)
}
