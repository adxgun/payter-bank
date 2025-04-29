package server

import (
	"context"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	_ "payter-bank/docs"
	"payter-bank/features/account"
	"payter-bank/features/auditlog"
	"payter-bank/features/interestrate"
	"payter-bank/features/transaction"
	"payter-bank/internal/api"
	"payter-bank/internal/config"
	"payter-bank/internal/database/models"
	"payter-bank/internal/pkg/generator"
)

type Server struct {
	accountHandler      *account.Handler
	transactionHandler  *transaction.Handler
	interestRateHandler *interestrate.Handler
	auditLogHandler     *auditlog.Handler
	cfg                 config.Config
	db                  models.Querier
}

func New(cfg config.Config, db models.Querier,
	accountHandler *account.Handler, txHandler *transaction.Handler, interestRateHandler *interestrate.Handler, auditLogHandler *auditlog.Handler) *Server {
	return &Server{accountHandler: accountHandler, db: db, cfg: cfg, transactionHandler: txHandler, interestRateHandler: interestRateHandler, auditLogHandler: auditLogHandler}
}

func (s *Server) BuildRoutes() (*gin.Engine, error) {
	authMW, err := s.jwtMiddleware()
	if err != nil {
		return nil, err
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(s.corsConfig()))
	if s.cfg.Server.EnableSwagger {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	v1 := r.Group("/api/v1")
	v1.POST("/users", api.Wrap(s.accountHandler.CreateUserHandler))
	v1.POST("/users/authenticate", api.Wrap(s.accountHandler.AuthenticateAccountHandler))

	authenticated := r.Group("/api/v1")
	authenticated.Use(authMW, currentProfileMiddleWare(s.db))
	authenticated.POST("/accounts", api.Wrap(s.accountHandler.CreateAccountHandler))
	authenticated.GET("/me", api.Wrap(s.accountHandler.MeHandler))
	authenticated.PATCH(
		"/accounts/:id/suspend", ensureAdminMiddleware(), api.Wrap(s.accountHandler.SuspendAccountHandler))
	authenticated.PATCH(
		"/accounts/:id/activate", ensureAdminMiddleware(), api.Wrap(s.accountHandler.ActivateAccountHandler))
	authenticated.PATCH(
		"/accounts/:id/close", ensureAdminMiddleware(), api.Wrap(s.accountHandler.CloseAccountHandler))
	authenticated.GET(
		"/accounts/:id/status-history", ensureAdminMiddleware(), api.Wrap(s.accountHandler.GetAccountStatusHistoryHandler))
	authenticated.GET(
		"/accounts/:id",
		api.Wrap(s.accountHandler.GetAccountDetailsHandler))
	authenticated.POST(
		"/credit",
		ensureAdminMiddleware(),
		api.Wrap(s.transactionHandler.CreditAccountHandler))
	authenticated.POST(
		"/debit",
		ensureAdminMiddleware(),
		api.Wrap(s.transactionHandler.DebitAccountHandler))
	authenticated.GET(
		"/accounts/:id/transactions",
		api.Wrap(s.transactionHandler.GetTransactionHistoryHandler))
	authenticated.GET(
		"/accounts/:id/balance",
		api.Wrap(s.transactionHandler.BalanceHandler))
	authenticated.POST(
		"/transfer",
		api.Wrap(s.transactionHandler.TransferFundsHandler))

	adminOnly := r.Group("/api/v1")
	adminOnly.Use(authMW, currentProfileMiddleWare(s.db), ensureAdminMiddleware())
	adminOnly.POST(
		"/interest-rate",
		api.Wrap(s.interestRateHandler.CreateInterestRateHandler))
	adminOnly.PUT(
		"/interest-rate",
		api.Wrap(s.interestRateHandler.UpdateRateHandler))
	adminOnly.PUT(
		"/interest-rate/calculation-frequency",
		api.Wrap(s.interestRateHandler.UpdateCalculationFrequencyHandler))
	adminOnly.GET(
		"/interest-rate/current",
		api.Wrap(s.interestRateHandler.GetCurrentRateHandler))
	adminOnly.POST("/admin/users", api.Wrap(s.accountHandler.CreateAdminUserHandler))
	adminOnly.GET("/accounts", api.Wrap(s.accountHandler.GetAllCurrentAccountsHandler))
	adminOnly.GET("/accounts/stats", api.Wrap(s.accountHandler.GetAccountsStatsHandler))
	adminOnly.GET("/accounts/:id/logs", api.Wrap(s.auditLogHandler.GetAccountAuditLogsHandler))

	return r, nil
}

func (s *Server) jwtMiddleware() (gin.HandlerFunc, error) {
	keyFunc := func(ctx context.Context) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	}

	customClaims := func() validator.CustomClaims {
		return &generator.Claim{}
	}

	jwtValidator, err := validator.New(
		keyFunc,
		validator.HS256,
		s.cfg.JWT.Issuer,
		[]string{s.cfg.JWT.Audience},
		validator.WithCustomClaims(customClaims),
	)
	if err != nil {
		return nil, err
	}

	mw := jwtmiddleware.New(jwtValidator.ValidateToken)
	return func(ctx *gin.Context) {
		encounteredError := true
		var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			encounteredError = false
			ctx.Request = r
			ctx.Next()
		}

		mw.CheckJWT(handler).ServeHTTP(ctx.Writer, ctx.Request)

		if encounteredError {
			ctx.Abort()
		}
	}, nil
}

func (s *Server) corsConfig() cors.Config {
	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = []string{s.cfg.Server.CorsOrigin}
	cfg.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	cfg.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	cfg.AllowCredentials = true
	return cfg
}
