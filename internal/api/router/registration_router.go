package router

import (
	"context"
	"fmt"
	"time"

	"cobra-template/internal/api/handlers"
	"cobra-template/internal/api/middleware"
	"cobra-template/internal/config"
	domain "cobra-template/internal/domain/registration"
	"cobra-template/internal/infrastructure/cache"
	"cobra-template/internal/infrastructure/queue"
	"cobra-template/internal/infrastructure/repository"
	interfaces "cobra-template/internal/interfaces/infrastructure"
	"cobra-template/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	semesterRepo := repository.NewSemesterRepository(db)

	registrationRepo := repository.NewRegistrationRepository(db)

	cfg := config.Get()
	cacheService := cache.NewRedisCacheWithConfig(&cfg.Cache)

	var waitlistRepo interfaces.WaitlistRepository
	if cfg.Registration.WaitlistRepository == "redis" {
		waitlistRepo = repository.NewRedisWaitlistRepository(cacheService.GetClient())
		fmt.Println("Using Redis waitlist repository")
	} else {
		waitlistRepo = repository.NewWaitlistRepository(db)
		fmt.Println("Using database waitlist repository")
	}
	idempotencyRepo := repository.NewRedisIdempotencyRepository(cacheService.GetClient())
	var queueService interfaces.QueueService
	if cfg.Queue.Type == "redis" {
		queueService = queue.NewRedisQueue(&cfg.Cache, 3)
		fmt.Println("Using Redis queue service")
	} else {
		queueService = queue.NewInMemoryQueue(cfg.Queue.BufferSize, 3)
		fmt.Println("Using in-memory queue service")
	}

	registrationService := service.NewRegistrationService(
		studentRepo,
		sectionRepo,
		registrationRepo,
		waitlistRepo,
		cacheService,
		queueService,
		idempotencyRepo,
		cfg.Registration.WaitlistFallbackEnabled,
	)

	if err := initializeMinimalCache(cacheService, sectionRepo, semesterRepo); err != nil {
		fmt.Printf("Warning: Failed to initialize minimal cache: %v\n", err)
	}

	queueService.SetRegistrationService(registrationService)
	queueService.StartWorkers()
	registrationHandler := handlers.NewRegistrationHandler(registrationService)
	healthHandler := handlers.NewHealthHandler()
	r.Use(middleware.IdempotencyMiddleware())
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

// initializeMinimalCache implements minimal pre-caching for seat availability and semester sections availability only
func initializeMinimalCache(
	cacheService interfaces.CacheService,
	sectionRepo interfaces.SectionRepository,
	semesterRepo interfaces.SemesterRepository,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Starting minimal cache initialization...")
	startTime := time.Now()

	if err := cacheActiveSectionsMinimal(ctx, cacheService, sectionRepo); err != nil {
		return fmt.Errorf("failed to cache active sections: %w", err)
	}

	if err := cacheSpecificSemesterSections(ctx, cacheService, sectionRepo); err != nil {
		return fmt.Errorf("failed to cache semester sections availability: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Minimal cache initialization completed in %v\n", duration)
	return nil
}

func cacheActiveSectionsMinimal(ctx context.Context, cacheService interfaces.CacheService, sectionRepo interfaces.SectionRepository) error {
	sections, err := sectionRepo.GetAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active sections: %w", err)
	}

	cached := 0
	for _, section := range sections {
		if err := cacheService.SetAvailableSeats(ctx, section.SectionID, section.AvailableSeats, 24*time.Hour); err != nil {
			fmt.Printf("Warning: Failed to cache seats for section %s: %v\n", section.SectionID, err)
			continue
		}
		cached++
	}

	fmt.Printf("📊 Cached seat availability for %d active sections\n", cached)
	return nil
}

func cacheSpecificSemesterSections(ctx context.Context, cacheService interfaces.CacheService, sectionRepo interfaces.SectionRepository) error {
	semesterID := uuid.MustParse("e093bb58-78e2-4985-bb7f-7a9b36c9102d")

	sections, err := sectionRepo.GetBySemester(ctx, semesterID)
	if err != nil {
		return fmt.Errorf("failed to get sections for semester %s: %w", semesterID, err)
	}

	availableSections := make([]*domain.Section, 0)
	for _, section := range sections {

		if cachedSeats, cacheErr := cacheService.GetAvailableSeats(ctx, section.SectionID); cacheErr == nil {
			section.AvailableSeats = cachedSeats
		}

		if section.AvailableSeats > 0 {
			availableSections = append(availableSections, section)
		}
	}

	if err := cacheService.SetAvailableSections(ctx, semesterID, availableSections, 8*time.Hour); err != nil {
		return fmt.Errorf("failed to cache available sections for semester %s: %w", semesterID, err)
	}

	fmt.Printf("📊 Cached %d available sections for semester %s\n", len(availableSections), semesterID)
	return nil
}
