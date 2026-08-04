package main

import (
	"log"

	"github.com/gin-gonic/gin"

	openapi "github.com/34Minnesota/avito-antiscam-monorepo/backend/generated/openapi"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/config"
	transport "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/transport/http"

	authservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/auth"
	authstorage "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/auth"
)

func main() {

	router := gin.Default()

	// PostgreSQL
	db := config.NewDatabase()
	defer db.Close()

	// Repository
	authRepository := authstorage.NewRepository(db)

	// Service
	authService := authservice.NewService(authRepository)

	// HTTP
	server := transport.NewServer(authService)

	// Временная ручка для проверки middleware.
	authorized := router.Group("/")

	authorized.Use(server.SessionMiddleware())

	authorized.GET("/test", func(c *gin.Context) {
		session, _ := c.Get("session")
		c.JSON(200, session)
	})

	// OpenAPI
	openapi.RegisterHandlers(router, server)

	log.Println("🚀 AntiScam API started on :8080")

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
