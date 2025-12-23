package hubserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (hs *HubServer) HandleQuery(c *gin.Context) {
	c.String(http.StatusOK, "Hello, World!")
}
