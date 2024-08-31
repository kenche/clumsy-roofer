package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewHandler(loggingEnabled bool) http.Handler {
	var router *gin.Engine
	if loggingEnabled {
		router = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Recovery())
	}

	store := NewStore()
	riskCtrller := NewRiskController(store)
	v1 := router.Group("/v1")
	{
		v1.GET("/risks", riskCtrller.list())
		v1.GET("/risks/", riskCtrller.list())
		v1.POST("/risks", riskCtrller.post())
		v1.POST("/risks/", riskCtrller.post())
		v1.GET("/risks/:id", riskCtrller.get())
		v1.GET("/risks/:id/", riskCtrller.get())
	}
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
	})

	return router.Handler()
}
