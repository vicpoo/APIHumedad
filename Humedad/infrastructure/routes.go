// routes.go
package infrastructure

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine, hub *Hub) {
	r.GET("/ws", hub.HandleWebSocket)
}
