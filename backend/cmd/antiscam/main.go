package main

import (
	"log"

	"github.com/gin-gonic/gin"

	openapi "github.com/34Minnesota/avito-antiscam-monorepo/backend/generated/openapi"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/api"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/config"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/auth"
)

func main() {

	router := gin.Default()

	db := config.NewDatabase()
	defer db.Close()

	repository := auth.NewSessionStore()

	service := auth.NewService(repository)

	server := api.NewServer(service)

	authorized := router.Group("/")

	authorized.Use(server.SessionMiddleware())

	authorized.GET("/test", func(c *gin.Context) {
		session, _ := c.Get("session")
		c.JSON(200, session)
	})

	openapi.RegisterHandlers(router, server)

	log.Println("🚀 AntiScam API started on :8080")

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}