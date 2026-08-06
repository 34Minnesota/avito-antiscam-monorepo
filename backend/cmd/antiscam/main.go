package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	openapi "github.com/34Minnesota/avito-antiscam-monorepo/backend/generated/openapi"
	transport "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/transport/http"

	applogger "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/logger"
	authservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/auth"
	usersservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/users"
	usersutils "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/users/utils"
	authstorage "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/auth"
	postgrespool "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/pool"
	userstorage "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/user"
	usershttp "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/transport/http/users"
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

	router := gin.New()
	router.Use(gin.Recovery())

	// PostgreSQL
	dbConfig, err := postgrespool.NewConfig()
	if err != nil {
		appLogger.Error("load PostgreSQL config failed", zap.Error(err))
		return
	}

	db, err := postgrespool.NewPool(context.Background(), dbConfig)
	if err != nil {
		appLogger.Error("connect to PostgreSQL failed", zap.Error(err))
		return
	}
	defer db.Close()

	// Repository
	authRepository := authstorage.NewRepository(db)
	userRepository := userstorage.NewRepository(db)

	// Users service
	userService := usersservice.NewUsersService(
		userRepository,
		usersutils.BcryptPasswordHasher{},
		usersutils.UUIDGenerator{},
		usersutils.RealClock{},
	)

	// Auth service
	authService := authservice.NewService(
		authRepository,
		userRepository,
		authservice.BcryptPasswordVerifier{},
		appLogger,
	)
	userHandler := usershttp.NewUsersHandler(userService)

	// HTTP
	server := transport.NewServer(authService, appLogger)
	router.Use(server.LoggerMiddleware())
	usershttp.RegisterUsersRoutes(router, userHandler)

	// Временная ручка для проверки middleware.
	authorized := router.Group("/")

	authorized.Use(server.SessionMiddleware())

	authorized.GET("/test", func(c *gin.Context) {
		session, _ := c.Get("session")
		c.JSON(200, session)
	})

	// OpenAPI
	openapi.RegisterHandlers(router, server)

	appLogger.Info("AntiScam API started", zap.String("address", ":8080"))

	if err := router.Run(":8080"); err != nil {
		appLogger.Error("HTTP server stopped with error", zap.Error(err))
	}
}
