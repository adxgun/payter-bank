package server

import (
	"context"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	_ "payter-bank/docs"
	"payter-bank/features/account"
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
	cfg                 config.Config
	db                  models.Querier
}

func New(cfg config.Config, db models.Querier,
	accountHandler *account.Handler, txHandler *transaction.Handler, interestRateHandler *interestrate.Handler) *Server {
	return &Server{accountHandler: accountHandler, db: db, cfg: cfg, transactionHandler: txHandler, interestRateHandler: interestRateHandler}
}

func (s *Server) BuildRoutes() (*gin.Engine, error) {
	authMW, err := s.jwtMiddleware()
	if err != nil {
		return nil, err
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	if s.cfg.Server.EnableSwagger {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	v1 := r.Group("/api/v1")
	v1.POST("/accounts", api.Wrap(s.accountHandler.CreateAccountHandler))
	v1.POST("/accounts/authenticate", api.Wrap(s.accountHandler.AuthenticateAccountHandler))

	authenticated := r.Group("/api/v1")
	authenticated.Use(authMW, currentProfileMiddleWare(s.db))
	authenticated.GET("/me", api.Wrap(s.accountHandler.MeHandler))
	authenticated.PATCH(
		"/accounts/:id/suspend", ensureAdminMiddleware(), api.Wrap(s.accountHandler.SuspendAccountHandler))
	authenticated.PATCH(
		"/accounts/:id/activate", ensureAdminMiddleware(), api.Wrap(s.accountHandler.ActivateAccountHandler))
	authenticated.PATCH(
		"/accounts/:id/close", ensureAdminMiddleware(), api.Wrap(s.accountHandler.SuspendAccountHandler))
	authenticated.GET(
		"/accounts/:id/status-history", ensureAdminMiddleware(), api.Wrap(s.accountHandler.GetAccountStatusHistoryHandler))
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
	adminOnly := authenticated.Group("/api/v1")
	adminOnly.Use(ensureAdminMiddleware())
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
