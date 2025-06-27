package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"internal-transfers/internal/config"
	"internal-transfers/internal/handlers"
	"internal-transfers/internal/repository"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)
	
	// Connect to database
	db, err := config.ConnectDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// Run database migrations
	if err := config.MigrateDatabase(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	
	// Initialize repository
	repo := repository.NewRepository(db)
	
	// Initialize handlers
	handler := handlers.NewHandler(repo)
	
	// Setup router
	router := setupRouter(handler)
	
	// Create server
	server := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRouter configures the Gin router with all routes
func setupRouter(h *handlers.Handler) *gin.Engine {
	router := gin.New()
	
	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	
	// API routes with proper method/handler naming
	api := router.Group("/")
	{
		// Account routes
		api.POST("/accounts", h.CreateAccountHandler)                 // POST /accounts -> CreateAccountHandler
		api.GET("/accounts/:account_id", h.GetAccountBalanceHandler)  // GET /accounts/{id} -> GetAccountBalanceHandler
		
		// Transaction routes  
		api.POST("/transactions", h.SubmitTransactionHandler)               // POST /submit -> SubmitTransactionHandler
		
		// Health check route
		api.GET("/health", h.HealthCheckHandler)                      // GET /health -> HealthCheckHandler
	}
	
	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}