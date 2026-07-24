package api

import (
	"facet-backend/internal/api/handler"
	"facet-backend/internal/api/middleware"
	"facet-backend/internal/config"
	"facet-backend/internal/domain"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewRouter creates and configures all application routes
func NewRouter(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	captchaHandler *handler.CaptchaHandler,
	userHandler *handler.UserHandler,
	environmentHandler *handler.EnvironmentHandler,
	ffHandler *handler.FeatureFlagHandler,
	tenantHandler *handler.TenantHandler,
	evalHandler *handler.EvaluationHandler,
	notificationHandler *handler.NotificationHandler,
	systemNotificationHandler *handler.SystemNotificationHandler,
	segmentHandler *handler.SegmentHandler,
	healthHandler *handler.HealthHandler,
) *gin.Engine {
	// Set Gin mode (debug/release/test)
	gin.SetMode(gin.ReleaseMode)
	if cfg.Logger.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// --- Health check endpoint (no auth) ---
	r.GET("/health", healthHandler.Check)

	// --- Swagger API docs ---
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- Register global error handling middleware ---
	r.Use(middleware.ErrorHandler())

	r.Use(middleware.RequestIDMiddleware())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: false,
	}))

	// API V1 route group
	apiV1 := r.Group("/api/v1")
	{
		// --- SDK evaluation route (no auth) ---
		apiV1.GET("/evaluate", evalHandler.Evaluate)

		// --- System-level management routes ---
		system := apiV1.Group("/system")
		system.Use(middleware.JWTMiddleware(cfg.JWT.SecretKey))
		system.Use(middleware.RequireRole(domain.RoleSystemAdmin))
		{
			// Tenant management routes
			tenants := system.Group("/tenants")
			{
				tenants.POST("", tenantHandler.Create)
				tenants.PUT("/:tenantId", tenantHandler.Update)
				tenants.PUT("/:tenantId/password", tenantHandler.ResetAdminPassword)
				tenants.GET("", tenantHandler.List)
				tenants.GET("/:tenantId", tenantHandler.GetByID)
			}

			// New: system notification management routes
			systemNotifications := system.Group("/notifications")
			{
				systemNotifications.POST("", systemNotificationHandler.Create)
				systemNotifications.GET("", systemNotificationHandler.List)
				systemNotifications.GET("/:notificationId", systemNotificationHandler.GetByID)
				systemNotifications.PUT("/:notificationId", systemNotificationHandler.Update)
				systemNotifications.DELETE("/:notificationId", systemNotificationHandler.Delete)
			}
		}

		// --- Tenant-level management routes ---
		manage := apiV1.Group("/manage")
		{
			// Routes without authentication
			auth := manage.Group("/auth")
			{
				auth.POST("/login", authHandler.Login)
				auth.GET("/captcha", captchaHandler.GetCaptcha) // <-- add captcha route
			}

			// Routes requiring authentication
			authorized := manage.Group("/")
			authorized.Use(middleware.JWTMiddleware(cfg.JWT.SecretKey))
			authorized.Use(middleware.RequireTenantWriteRole())
			{
				// User management routes
				users := authorized.Group("/users")
				{
					users.POST("", userHandler.Create)
					users.PUT("/:userId", userHandler.Update)
					users.DELETE("/:userId", userHandler.Delete)
					users.PUT("/:userId/password", userHandler.ResetPassword)
					users.GET("/:userId", userHandler.GetByID)
					users.GET("", userHandler.List)
				}

				// Environment management routes
				environments := authorized.Group("/environments")
				{
					environments.POST("", environmentHandler.Create)
					environments.PUT("/:envId", environmentHandler.Update)
					environments.DELETE("/:envId", environmentHandler.Delete)
					environments.GET("", environmentHandler.List)
					environments.GET("/active-list", environmentHandler.ListActive)

					// Feature flag routes (nested under environment routes)
					features := environments.Group("/:envId/features")
					{
						features.POST("", ffHandler.Create)
						features.GET("", ffHandler.List)
						features.GET("/:featureId", ffHandler.GetByID)
						features.PUT("/:featureId", ffHandler.Update)
						features.DELETE("/:featureId", ffHandler.Delete)
						features.GET("/:featureId/history", ffHandler.GetHistory)
						features.POST("/:featureId/archive", ffHandler.Archive)
						features.POST("/:featureId/unarchive", ffHandler.Unarchive)
					}
				}

				// Segment management routes
				segments := authorized.Group("/segments")
				{
					segments.POST("", segmentHandler.Create)
					segments.GET("", segmentHandler.List)
					segments.GET("/:segmentId", segmentHandler.GetByID)
					segments.PUT("/:segmentId", segmentHandler.Update)
					segments.DELETE("/:segmentId", segmentHandler.Delete)
				}

				// User notification routes
				notifications := authorized.Group("/notifications")
				{
					notifications.GET("", notificationHandler.ListForUser)
					notifications.GET("/status", notificationHandler.GetStatus)
					notifications.POST("/mark-as-read", notificationHandler.MarkAsRead)
				}
			}
		}
	}
	return r
}
