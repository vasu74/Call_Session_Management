package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/vasu74/Call_Session_Management/internal/handler"
	"github.com/vasu74/Call_Session_Management/internal/middleware"
)

func Routes(server *gin.Engine) {
	// Public routes
	auth := server.Group("/auth")
	{
		auth.POST("/register", handler.RegisterHandler)
		auth.POST("/login", handler.LoginHandler)
	}

	// Protected routes
	api := server.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// User profile
		api.GET("/profile", handler.GetProfileHandler)

		// Session routes
		sessions := api.Group("/sessions")
		{
			sessions.GET("", handler.ListSessionsHandler)
			sessions.POST("/start", handler.StartSessionHandler)
			sessions.POST("/:sessionId/events", handler.LogSessionEventHandler)
			sessions.POST("/:sessionId/end", handler.EndSessionHandler)
			sessions.GET("/:sessionId", handler.GetSessionDetailsHandler)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			// Add admin-specific routes here
		}
	}
}
