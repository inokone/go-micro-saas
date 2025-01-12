package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @BasePath /api/v1

// Healthcheck is the REST handler for polling whether the application is available.
// @Summary Health check endpoint of the Schedlue.me app
// @Schemes
// @Description Returns the status and version of the application
// @Accept json
// @Produce json
// @Success 200 {object} common.Health
// @Router /healthcheck [get]
func Healthcheck(g *gin.Context) {
	g.JSON(http.StatusOK, Health{
		Status:  "Ok",
		Version: "0.1",
	})
}
