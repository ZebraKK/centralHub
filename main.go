package main

import (
	"os"

	"github.com/gin-gonic/gin"

	hubserver "centralHub/hub_server"
	"centralHub/logger"
	"centralHub/middleware"
)

func main() {

	// config

	hubServer := hubserver.NewHubServer()

	router := setupRouter(hubServer)

	logger.RunLogger.Info().Str("addr", ":8080").Msg("centralhub server starting")
	if err := router.Run(":8080"); err != nil {
		logger.RunLogger.Fatal().Err(err).Msg("Failed to start centralhub server")
	}
}

func setupRouter(hubServer *hubserver.HubServer) *gin.Engine {

	// middleware can be added here if needed
	// log
	isProd := os.Getenv("GIN_MODE") == gin.ReleaseMode
	logger.InitLogger(isProd)
	// authorization
	// metrics
	// tracing
	// recovery

	// Define routes
	gin.SetMode(os.Getenv("GIN_MODE"))
	r := gin.Default()

	r.Use(middleware.AuditLog())

	r.POST("/create", hubServer.HandleCreate)
	r.GET("/query", hubServer.HandleQuery)

	return r
}
