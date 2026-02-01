package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterPaymentRoutes 注册支付相关路由
func RegisterPaymentRoutes(v1 *gin.RouterGroup, h *handler.Handlers, jwtAuth servermiddleware.JWTAuthMiddleware) {
	// user authenticated
	authed := v1.Group("")
	authed.Use(gin.HandlerFunc(jwtAuth))
	{
		payments := authed.Group("/payments")
		{
			payments.GET("/products", h.Payment.ListProducts)
			payments.GET("/orders", h.Payment.ListOrders)
			payments.POST("/orders", h.Payment.CreateOrder)
			payments.GET("/orders/:order_no", h.Payment.GetOrder)
		}
	}

	// public notify
	public := v1.Group("/payments")
	{
		// Some providers send GET callbacks; accept both GET/POST for compatibility.
		public.GET("/notify/:provider", h.Payment.Notify)
		public.POST("/notify/:provider", h.Payment.Notify)
	}
}
