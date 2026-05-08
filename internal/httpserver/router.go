package httpserver

import (
	"net/http"

	"github.com/Danil-Ivonin/TestSubs/internal/subscription"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type RouterOptions struct {
	SubscriptionHandler *subscription.Handler
}

func NewRouter(logger *logrus.Logger, opts RouterOptions) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger(logger))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if opts.SubscriptionHandler != nil {
		api := router.Group("/api/v1")
		opts.SubscriptionHandler.RegisterRoutes(api)
	}

	return router
}
