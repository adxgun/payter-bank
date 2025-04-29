package main

import (
	"context"
	"fmt"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"payter-bank/features/account"
	"payter-bank/features/auditlog"
	"payter-bank/features/interestrate"
	"payter-bank/features/transaction"
	"payter-bank/internal/config"
	"payter-bank/internal/database"
	"payter-bank/internal/database/models"
	"payter-bank/internal/logger"
	"payter-bank/internal/pkg/generator"
	"payter-bank/server"
	"syscall"
)

// @title           PayterBank API
// @version         1.0
// @description     REST API powering a payter banking platform.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    support@payterbank.app
// @contact.email  support@payterbank.app

// @host      localhost:2025
// @BasePath  /docs
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.Config{}
	err := envconfig.Process(ctx, &cfg)
	if err != nil {
		logger.Fatal(ctx, "Error loading configuration", zap.Error(err))
	}

	logger.Info(ctx, "Configuration loaded",
		zap.Any("config", cfg))

	db, err := database.Open(ctx, cfg.DB.DSN)
	if err != nil {
		logger.Fatal(ctx, "Error opening database", zap.Error(err))
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal(ctx, "Error closing database", zap.Error(err))
		}
	}()

	auditLogClient := auditlog.NewClient(cfg.Redis)

	querier := models.New(db)
	tokenGenerator := generator.NewTokenGenerator(cfg.JWT)
	auditLogService := auditlog.NewService(cfg, auditLogClient, querier)
	interestRateApplicationRunner := interestrate.NewRunner(querier, cfg.App)

	transactionService := transaction.NewService(querier, auditLogService)
	accountService := account.NewService(querier, auditLogService, transactionService, tokenGenerator)
	interestService := interestrate.NewService(querier, cfg.App, auditLogService, interestRateApplicationRunner)
	auditLogQueryService := auditlog.NewQueryService(querier)

	accountHandler := account.NewHandler(accountService)
	transactionHandler := transaction.NewHandler(transactionService)
	interestRateHandler := interestrate.NewHandler(interestService)
	auditLogHandler := auditlog.NewHandler(auditLogQueryService)

	srvHandler := server.New(cfg, querier, accountHandler, transactionHandler, interestRateHandler, auditLogHandler)
	routes, err := srvHandler.BuildRoutes()
	if err != nil {
		logger.Fatal(ctx, "Error building routes", zap.Error(err))
	}

	go func() {
		if err := auditLogService.Start(ctx); err != nil {
			logger.Fatal(ctx, "Error starting audit log service", zap.Error(err))
		}
	}()

	go func() {
		if err := interestRateApplicationRunner.Start(ctx); err != nil {
			logger.Warn(ctx, "Error starting interest-rate calculation job", zap.Error(err))
		}
	}()

	if err := accountService.InitialiseAdmin(ctx, cfg.App.AdminEmail, cfg.App.AdminPassword); err != nil {
		logger.Fatal(ctx, "Error initializing admin account", zap.Error(err))
	}

	addr := ":" + cfg.Server.Port
	srv := http.Server{
		Addr:    addr,
		Handler: routes,
	}
	go func() {
		logger.Info(ctx, fmt.Sprintf("Listening and serving HTTP on %s", addr))
		if err := srv.ListenAndServe(); err != nil {
			logger.Fatal(ctx, "Server closed", zap.Error(err))
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	logger.Info(ctx, "Shutting down...")
	ctx, cancel = context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(ctx, "Server forced to shutdown", zap.Error(err))
	}
}
