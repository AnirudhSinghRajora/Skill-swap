package router

import (
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/admin"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/repository"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/availability"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/chat"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/cohort"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/email"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/moderation"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/qa"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/rating"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/session"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/skill"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/swap"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/video"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes configures all application routes by delegating to specific route files
func SetupRoutes(api *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	emailService := email.NewService(*cfg)
	authService := service.NewAuthServiceWithEmail(userRepo, *cfg, db, emailService)
	skillService := service.NewSkillService(db)
	notificationService := service.NewNotificationServiceWithEmail(db, emailService)
	swapService := service.NewSwapService(db, notificationService, *cfg)
	ratingService := service.NewRatingService(db)
	adminService := service.NewAdminService(db)
	availabilityService := service.NewAvailabilityService(db)
	searchService := service.NewSearchService(db)
	fileUploadService := service.NewFileUploadService(db)
	chatRepo := repository.NewChatRepository(db)
	chatService := service.NewChatService(chatRepo, db, notificationService)
	moderationService := service.NewModerationService(db)
	sessionService := service.NewSessionService(db)
	cohortService := service.NewCohortService(db)
	qaService := service.NewQAService(db)

	// Video call service (shared by video + cohort handlers)
	videoService := service.NewVideoService(*cfg)

	// Initialize handlers
	skillHandler := skill.NewHandler(skillService)
	swapHandler := swap.NewHandler(swapService)
	ratingHandler := rating.NewHandler(ratingService)
	adminHandler := admin.NewHandler(adminService)
	availabilityHandler := availability.NewHandler(availabilityService)
	chatHandler := chat.NewHandler(chatService)
	moderationHandler := moderation.NewHandler(moderationService)
	sessionHandler := session.NewHandler(sessionService)
	cohortHandler := cohort.NewHandler(cohortService, videoService)
	qaHandler := qa.NewHandler(qaService)
	videoHandler := video.NewHandler(videoService)

	// WebSocket hub — singleton for the lifetime of the application.
	hub := chat.NewHub()
	go hub.Run()
	wsHandler := chat.NewWSHandler(hub, chatService, chatRepo, cfg)

	// Setup route groups
	SetupAuthRoutes(api, authService, cfg)
	SetupUserRoutes(api, userService, cfg)
	SetupSkillRoutes(api, cfg, skillHandler)
	SetupSwapRoutes(api, cfg, swapHandler)
	SetupRatingRoutes(api, cfg, ratingHandler)
	SetupAvailabilityRoutes(api, cfg, availabilityHandler)
	SetupAdminRoutes(api, cfg, skillHandler, adminHandler)
	SetupNotificationRoutes(api, notificationService, cfg)
	SetupSearchRoutes(api, searchService, cfg)
	SetupFileRoutes(api, fileUploadService, cfg)
	SetupChatRoutes(api, cfg, chatHandler)
	SetupVideoRoutes(api, cfg, videoHandler)
	SetupModerationRoutes(api, cfg, moderationHandler)
	SetupSessionRoutes(api, cfg, sessionHandler)
	SetupCohortRoutes(api, cfg, cohortHandler)
	SetupQARoutes(api, cfg, qaHandler)

	// WebSocket endpoint — auth is handled inside the upgrade handler
	// (token passed via query param), so no JWT middleware here.
	api.GET("/ws", wsHandler.HandleWebSocket)
}
