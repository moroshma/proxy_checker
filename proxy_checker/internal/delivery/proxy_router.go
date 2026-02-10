package delivery

import "github.com/gin-gonic/gin"

func RegisterServiceRoutes(server *gin.Engine, proxyHandler *ProxyHandler) {
	proxyRoute := server.Group("api/v1/proxy")
	proxyRoute.POST("", proxyHandler.Create)
	proxyRoute.GET("/history", proxyHandler.GetHistory)
	proxyRoute.GET("/:id", proxyHandler.GetStatus)
}
