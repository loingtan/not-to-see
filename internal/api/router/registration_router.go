package router

import (
	"context"
	"fmt"
	"time"

	"cobra-template/internal/api/handlers"
	"cobra-template/internal/api/middleware"
	"cobra-template/internal/config"
	"cobra-template/internal/infrastructure/cache"
	"cobra-template/internal/infrastructure/queue"
	"cobra-template/internal/infrastructure/repository"
	interfaces "cobra-template/internal/interfaces/infrastructure"
	"cobra-template/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RouterComponents struct {
	Router       *gin.Engine
	QueueService interfaces.QueueService
}

func NewRegistrationRouter(db *gorm.DB) *gin.Engine {
	components := NewRegistrationRouterWithQueue(db)
	return components.Router
}

func NewRegistrationRouterWithQueue(db *gorm.DB) *RouterComponents {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(middleware.Logger())
	r.Use(cors.Default())
	r.Use(gin.Recovery())
	studentRepo := repository.NewStudentRepository(db)
	sectionRepo := repository.NewSectionRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	waitlistRepo := repository.NewWaitlistRepository(db)

	cfg := config.Get()
	redisAddr := fmt.Sprintf("%s:%d", cfg.Cache.Host, cfg.Cache.Port)
	cacheService := cache.NewRedisCache(redisAddr, cfg.Cache.Password, cfg.Cache.DB)

	queueService := queue.NewInMemoryQueue(1000, 10)
	registrationService := service.NewRegistrationService(
		studentRepo,
		sectionRepo,
		registrationRepo,
		waitlistRepo,
		cacheService,
		queueService,
	)

	if err := initializeSectionCache(cacheService, sectionRepo); err != nil {

		fmt.Printf("Warning: Failed to initialize section cache: %v\n", err)
	}

	queueService.SetRegistrationService(registrationService)
	queueService.StartWorkers()

	registrationHandler := handlers.NewRegistrationHandler(registrationService)
	healthHandler := handlers.NewHealthHandler()
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadinessCheck)
	r.GET("/live", healthHandler.LivenessCheck)
	v1 := r.Group("/api/v1")
	{
		registration := v1.Group("/register")
		{
			registration.POST("", registrationHandler.Register)
			registration.POST("/drop", registrationHandler.DropCourse)
		}

		students := v1.Group("/students")
		{
			students.GET("/:student_id/registrations", registrationHandler.GetStudentRegistrations)
			students.GET("/:student_id/waitlist", registrationHandler.GetWaitlistStatus)
		}

		sections := v1.Group("/sections")
		{
			sections.GET("/available", registrationHandler.GetAvailableSections)
		}
	}

	return &RouterComponents{
		Router:       r,
		QueueService: queueService,
	}
}

func initializeSectionCache(cacheService interfaces.CacheService, sectionRepo interfaces.SectionRepository) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sections, err := sectionRepo.GetAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active sections: %w", err)
	}

	for _, section := range sections {
		if err := cacheService.SetAvailableSeats(ctx, section.SectionID, section.AvailableSeats, 24*time.Hour); err != nil {

			fmt.Printf("Warning: Failed to cache seats for section %s: %v\n", section.SectionID, err)
		}
	}

	fmt.Printf("Successfully initialized cache with %d sections\n", len(sections))
	return nil
}
