package httpserver

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shuza/Autonoma/internal/platform/config"
	"github.com/shuza/Autonoma/internal/platform/health"
)

func New(cfg config.Config) *http.Server {
	return NewWithDependencies(cfg, nil)
}

func NewWithDependencies(ctf config.Config, leadCreator leadCreator) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.NoRoute(handleNotFound)
	router.NoMethod(handleMethodNotAllowed)

	router.GET("/healthz", handleHealth)
	router.POST("/v1/leads", handleCreateLead(leadCreator))

	return &http.Server{
		Addr:              ctf.Address(),
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func handleHealth(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, health.Check())
}

func handleNotFound(c *gin.Context) {
	c.Status(http.StatusNotFound)
}

func handleMethodNotAllowed(c *gin.Context) {
	c.String(http.StatusMethodNotAllowed, "Method not allowed")
}
