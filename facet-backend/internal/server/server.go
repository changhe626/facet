package server

import (
	"context"
	"facet-backend/internal/api"
	"facet-backend/internal/api/handler"
	"facet-backend/internal/cache"
	"facet-backend/internal/config"
	"facet-backend/internal/repository/postgres"
	"facet-backend/internal/service"
	"facet-backend/internal/worker"
	"facet-backend/pkg/logger"
	"fmt"
	"log/slog"
	"os"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Server holds all application dependencies
type Server struct {
	engine    *gin.Engine
	cfg       *config.Config
	logWorker *worker.LogWorker
}

// NewServer creates and configures a new Server instance
func NewServer() (*Server, error) {
	// 1. Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Initialize logger
	logger.InitLogger(cfg.Logger)

	// 3. Connect to database
	pgDSN := os.Getenv("POSTGRES_DSN")
	if pgDSN == "" {
		pgDSN = cfg.Postgres.DSN
	}
	dbConn, err := postgres.NewPostgresDB(pgDSN)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		return nil, err
	}

	// 3b. Run database migrations
	if err := postgres.RunMigrations(dbConn); err != nil {
		slog.Error("failed to run database migrations", "error", err)
		return nil, err
	}

	// 4. Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		slog.Error("failed to connect to redis", "error", err)
		return nil, err
	}

	// 5. Create snowflake node
	snowflakeNode, err := snowflake.NewNode(1)
	if err != nil {
		slog.Error("failed to create snowflake node", "error", err)
		return nil, err
	}

	// --- Dependency Injection ---

	// Repository layer
	userRepo := postgres.NewUserRepository(dbConn)
	envRepo := postgres.NewEnvironmentRepository(dbConn)
	ffRepo := postgres.NewFeatureFlagRepository(dbConn)
	ruleRepo := postgres.NewFeatureFlagRuleRepository(dbConn)
	tenantRepo := postgres.NewTenantRepository(dbConn)
	levelRepo := postgres.NewTenantLevelRepository(dbConn)
	historyRepo := postgres.NewFeatureFlagHistoryRepository(dbConn)
	notificationRepo := postgres.NewNotificationRepository(dbConn, snowflakeNode)
	segRepo := postgres.NewSegmentRepository(dbConn)
	condRepo := postgres.NewSegmentConditionRepository(dbConn)
	ffSegRepo := postgres.NewFeatureFlagSegmentRepository(dbConn)

	// Infrastructure / Cache Service
	cacheService := cache.NewRedisCacheService(redisClient)

	// Service layer
	eventService := service.NewEventService(redisClient) // <-- Added EventService
	captchaService := service.NewCaptchaService(redisClient)
	authService := service.NewAuthService(userRepo, cfg.JWT.SecretKey, cfg.JWT.ExpiresInHours)

	// 1. Create data source and evaluation services
	evalDataSource := service.NewEvaluationDataSource(cacheService, envRepo, tenantRepo, ffRepo, ruleRepo, segRepo, condRepo, ffSegRepo, slog.Default())
	evalService := service.NewEvaluationService(evalDataSource)

	// 2. Create tenant level service (arbitration center) and inject cache
	tenantLevelService := service.NewTenantLevelService(levelRepo, tenantRepo, envRepo, ffRepo, userRepo, cacheService)

	// 3. Create other services that depend on the arbitration center and cache
	tenantService := service.NewTenantService(dbConn, tenantRepo, userRepo, levelRepo, cacheService, snowflakeNode)
	userService := service.NewUserService(userRepo, snowflakeNode, tenantLevelService)
	envService := service.NewEnvironmentService(envRepo, snowflakeNode, cacheService, tenantLevelService)
	ffService := service.NewFeatureFlagService(ffRepo, ruleRepo, historyRepo, userRepo, envRepo, snowflakeNode, cacheService, tenantLevelService, ffSegRepo)
	notificationService := service.NewNotificationService(notificationRepo, snowflakeNode)
	segService := service.NewSegmentService(segRepo, condRepo, snowflakeNode)

	// Handler layer
	captchaHandler := handler.NewCaptchaHandler(captchaService)
	authHandler := handler.NewAuthHandler(authService, captchaService)
	userHandler := handler.NewUserHandler(userService)
	envHandler := handler.NewEnvironmentHandler(envService)
	ffHandler := handler.NewFeatureFlagHandler(ffService)
	tenantHandler := handler.NewTenantHandler(tenantService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	systemNotificationHandler := handler.NewSystemNotificationHandler(notificationService)
	segmentHandler := handler.NewSegmentHandler(segService)
	healthHandler := handler.NewHealthHandler(dbConn, redisClient)
	// EvaluationHandler now depends on EventService
	evalHandler := handler.NewEvaluationHandler(evalService, eventService, snowflakeNode)

	// Worker layer
	logWorker := worker.NewLogWorker(redisClient, dbConn)

	// --- End of Dependency Injection ---

	// Initialize Gin engine and configure routes
	engine := api.NewRouter(cfg, authHandler, captchaHandler, userHandler, envHandler, ffHandler, tenantHandler, evalHandler, notificationHandler, systemNotificationHandler, segmentHandler, healthHandler)

	server := &Server{
		engine:    engine,
		cfg:       cfg,
		logWorker: logWorker, // <-- Save Worker instance
	}

	return server, nil
}

// Start starts background workers
func (s *Server) Start() error {
	go s.logWorker.Start()
	slog.Info("server workers started", "port", s.cfg.Server.Port)
	return nil
}

// Stop gracefully shuts down background workers
func (s *Server) Stop() {
	slog.Info("stopping log worker...")
	s.logWorker.Stop()
}

// Engine returns the Gin engine
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// Cfg returns the server configuration
func (s *Server) Cfg() *config.Config {
	return s.cfg
}
