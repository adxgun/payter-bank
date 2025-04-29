package auditlog

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payter-bank/internal/api"
)

type Handler struct {
	service Query
}

func NewHandler(service Query) *Handler {
	return &Handler{
		service: service,
	}
}

// GetAccountAuditLogsHandler godoc
// @Summary      Get account details.
// @Description  Get account details.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Success      200  {object}  api.SuccessResponse{data=[]AuditLog}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /v1/api/accounts/:id/logs [get]
func (h *Handler) GetAccountAuditLogsHandler(ctx *gin.Context) api.Response {
	accountID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return api.BadRequest("account ID is required")
	}

	data, err := h.service.GetAuditLogs(ctx, accountID)
	if err != nil {
		return api.Error(err)
	}

	return api.OK("account audit logs retrieved successfully", data)
}
