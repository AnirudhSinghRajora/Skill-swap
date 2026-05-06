package router

import (
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/middleware"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/notification"
	"github.com/gin-gonic/gin"
)

func SetupNotificationRoutes(api *gin.RouterGroup, notificationService *service.NotificationService, cfg *config.Config) {
	notificationHandler := notification.NewHandler(notificationService)

	// Protected notification routes
	notifications := api.Group("/notifications")
	notifications.Use(middleware.JWTAuth(*cfg))
	{
		notifications.GET("", notificationHandler.GetUserNotifications)              // GET /api/notifications
		notifications.GET("/stats", notificationHandler.GetNotificationStats)        // GET /api/notifications/stats
		notifications.GET("/:id", notificationHandler.GetNotificationByID)           // GET /api/notifications/:id
		notifications.PUT("/mark-read", notificationHandler.MarkNotificationsAsRead) // PUT /api/notifications/mark-read
		notifications.PUT("/mark-all-read", notificationHandler.MarkAllAsRead)       // PUT /api/notifications/mark-all-read
		notifications.DELETE("/:id", notificationHandler.DeleteNotification)         // DELETE /api/notifications/:id

		// Admin only routes
		adminNotifications := notifications.Group("")
		adminNotifications.Use(middleware.AdminAuth())
		{
			adminNotifications.POST("", notificationHandler.CreateNotification) // POST /api/notifications (admin only)
		}
	}
}
