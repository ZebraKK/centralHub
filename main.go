package main

import (
	hubserver "centralHub/hubServer"

	"github.com/gin-gonic/gin"
)

func main() {

	// config

	hubServer := hubserver.NewHubServer()

	router := setupRouter(hubServer)
	router.Run()
}

func setupRouter(hubServer *hubserver.HubServer) *gin.Engine {
	r := gin.Default()
	// middleware can be added here if needed
	// log
	// authorization
	// metrics
	// tracing
	// recovery

	// Define routes

	r.POST("/create", hubServer.HandleCreate)
	r.GET("/query", hubServer.HandleQuery)

	return r
}
