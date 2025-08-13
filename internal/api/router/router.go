package router

import (
	"cobra-template/internal/api/handlers"
	"cobra-template/internal/api/middleware"
	"cobra-template/internal/infrastructure/repository"
	"cobra-template/internal/service"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	userRepo := repository.NewMockUserRepository()
	userService := service.NewUserService(userRepo)

	userHandler := handlers.NewUserHandler(userService)
	healthHandler := handlers.NewHealthHandler()

	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadinessCheck)
	r.GET("/live", healthHandler.LivenessCheck)

	v1 := r.Group("/api/v1")
	{

		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.GET("/email/:email", userHandler.GetUserByEmail)
			users.GET("/username/:username", userHandler.GetUserByUsername)
		}
	}
	return r
}
