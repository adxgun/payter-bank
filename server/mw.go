package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"payter-bank/internal/api"
	"payter-bank/internal/auth"
	"payter-bank/internal/database/models"
)

func currentProfileMiddleWare(db models.Querier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := auth.GetTokenData(ctx)
		if err != nil {
			ctx.JSON(403, api.ErrorResponse{
				Error: err.Error(),
			})
			ctx.Abort()
			return
		}

		row, err := db.GetProfileByUserID(ctx.Request.Context(), token.UserID)
		if err != nil {
			ctx.JSON(403, api.ErrorResponse{
				Error: "Unauthorized",
			})
			ctx.Abort()
			return
		}

		profile := auth.Profile{
			AccountID:    row.AccountID,
			UserID:       row.UserID,
			Email:        row.Email,
			FirstName:    row.FirstName,
			LastName:     row.LastName,
			AccountType:  string(row.AccountType),
			UserType:     string(row.UserType),
			RegisteredAt: row.RegisteredAt.Time,
		}

		c := context.WithValue(ctx.Request.Context(), auth.ProfileKey, profile)
		ctx.Request = ctx.Request.WithContext(c)
		ctx.Next()
	}
}

func ensureAdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		profile, err := auth.GetCurrentProfile(ctx)
		if err != nil {
			ctx.JSON(403, api.ErrorResponse{
				Error: "Unauthorized",
			})
			ctx.Abort()
			return
		}

		if profile.UserType != string(models.UserTypeADMIN) {
			ctx.JSON(403, api.ErrorResponse{
				Error: "Forbidden",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
