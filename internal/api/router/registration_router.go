package router

import (
	"cobra-template/internal/api/handlers"
	"cobra-template/internal/api/middleware"
	"cobra-template/internal/infrastructure/cache"
	"cobra-template/internal/infrastructure/queue"
	"cobra-template/internal/infrastructure/repository"
	"cobra-template/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRegistrationRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(middleware.Logger())
	r.Use(cors.Default())
	r.Use(gin.Recovery())

	// Initialize repositories
	studentRepo := repository.NewStudentRepository(db)
	sectionRepo := repository.NewSectionRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	waitlistRepo := repository.NewWaitlistRepository(db)

	// Initialize services
	cacheService := cache.NewRedisCache("localhost:6379", "", 0)
	queueService := queue.NewInMemoryQueue(1000, 10)
	registrationService := service.NewRegistrationService(
		studentRepo,
		sectionRepo,
		registrationRepo,
		waitlistRepo,
		cacheService,
		queueService,
	)

	// Initialize handlers
	registrationHandler := handlers.NewRegistrationHandler(registrationService)
	healthHandler := handlers.NewHealthHandler()

	// Health endpoints
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadinessCheck)
	r.GET("/live", healthHandler.LivenessCheck)

	// API v1 routes
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

	return r
}
