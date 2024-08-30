package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewServer(loggingEnabled bool) http.Handler {
	var router *gin.Engine
	if loggingEnabled {
		router = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Recovery())
	}

	store := NewStore()
	riskHandler := NewRiskHandler(store)
	v1 := router.Group("/v1")
	{
		v1.GET("/risks", riskHandler.list())
		v1.POST("/risks", riskHandler.post())
		v1.GET("/risks/:id", riskHandler.get())
	}

	return router.Handler()
}
