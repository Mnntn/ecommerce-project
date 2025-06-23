package main

import (
	"log"
	"net/http"

	"github.com/mnntn/ecommerce-project/api-gateway/internal/config"
	"github.com/mnntn/ecommerce-project/api-gateway/internal/router"
)

func main() {
	// Load configuration
	cfg := config.NewConfig()

	// Create router
	r := router.NewRouter(cfg)

	// Start server
	log.Printf("Starting API Gateway on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
