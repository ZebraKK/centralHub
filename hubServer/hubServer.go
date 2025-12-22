package hubserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HubServer struct {
}

func NewHubServer() *HubServer {
	return &HubServer{}
}

func (hs *HubServer) HandleCreate(c *gin.Context) {

	//parse form data

	// write http response
	// error
}

func (hs *HubServer) HandleQuery(c *gin.Context) {
	c.String(http.StatusOK, "Hello, World!")
}
