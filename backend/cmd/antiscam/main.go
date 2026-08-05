package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	applogger "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/logger"
	authservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/auth"
	trainingservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
	usersservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/users"
	usersutils "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/users/utils"
	authstorage "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/auth"
	postgrespool "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/pool"
	trainingstorage "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/training"
	userstorage "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/user"
	transport "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/transport/http"
	usershttp "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/transport/http/users"
)

const (
	serverAddress   = ":8080"
	shutdownTimeout = 10 * time.Second

	// Каталог со сценариями, читается на старте. Путь относительно рабочей
	// директории процесса. Сид идемпотентен: ключ — slug из документа.
	scenariosDir = "./docs/scenarios"
)

func main() {
	loggerConfig, err := applogger.NewConfig()
	if err != nil {
		log.Fatalf("load logger config: %v", err)
	}

	appLogger, err := applogger.NewLogger(loggerConfig)
	if err != nil {
		log.Fatalf("create logger: %v", err)
	}
	defer func() {
		if err := appLogger.Sync(); err != nil {
			log.Printf("sync logger: %v", err)
		}
		appLogger.Close()
	}()

	// Контекст живёт до первого SIGINT/SIGTERM и служит сигналом к остановке.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// PostgreSQL
	dbConfig, err := postgrespool.NewConfig()
	if err != nil {
		appLogger.Error("load PostgreSQL config failed", zap.Error(err))
		return
	}

	db, err := postgrespool.NewPool(ctx, dbConfig)
	if err != nil {
		appLogger.Error("connect to PostgreSQL failed", zap.Error(err))
		return
	}
	defer db.Close()

	// Repository
	authRepository := authstorage.NewRepository(db)
	trainingRepository := trainingstorage.New(db)

	// Service
	authService := authservice.NewService(authRepository, appLogger)
	trainingService := trainingservice.New(trainingRepository)

	// Сценарии заливаются на старте. Отсутствие каталога не повод не подниматься:
	// API останется рабочим, просто каталог сценариев будет пустым.
	if loaded, err := trainingservice.Seed(ctx, trainingRepository, os.DirFS(scenariosDir), "."); err != nil {
		appLogger.Warn("seed scenarios failed", zap.String("dir", scenariosDir), zap.Error(err))
	} else {
		appLogger.Info("scenarios seeded", zap.Int("count", loaded))
	}

	// users feature
	userRepository := userstorage.NewRepository(db)
	userService := usersservice.NewUsersService(
		userRepository,
		usersutils.BcryptPasswordHasher{},
		usersutils.UUIDGenerator{},
		usersutils.RealClock{},
	)
	userHandler := usershttp.NewUsersHandler(userService)

	// HTTP
	server := transport.NewServer(authService, trainingService, appLogger)

	router := gin.New()
	router.Use(gin.Recovery(), server.LoggerMiddleware())

	router.GET("/healthz", server.HealthCheck)
	router.POST("/v1/sessions", server.CreateSession)

	usershttp.RegisterUsersRoutes(router, userHandler)

	// Всё остальное требует X-Session-ID.
	v1 := router.Group("/v1", server.SessionMiddleware())
	v1.GET("/scenarios", server.ListScenarios)
	v1.POST("/attempts", server.StartAttempt)
	v1.POST("/attempts/:attemptID/choice", server.SubmitChoice)
	v1.GET("/attempts/:attemptID/summary", server.GetSummary)

	httpServer := &http.Server{
		Addr:              serverAddress,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("HTTP server stopped with error", zap.Error(err))
			stop()
		}
	}()

	appLogger.Info("AntiScam API started", zap.String("address", serverAddress))

	<-ctx.Done()
	appLogger.Info("shutting down")

	// Отдельный контекст: ctx уже отменён сигналом, по нему Shutdown вернётся сразу.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("graceful shutdown failed", zap.Error(err))
	}
}
