package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mnntn/ecommerce-project/order-service/internal/kafka"
	"github.com/mnntn/ecommerce-project/order-service/internal/migration"
	"github.com/mnntn/ecommerce-project/order-service/internal/outbox"
	"github.com/mnntn/ecommerce-project/order-service/internal/repository/postgres"
	"github.com/mnntn/ecommerce-project/order-service/internal/service"
	httptransport "github.com/mnntn/ecommerce-project/order-service/internal/transport/http"
)

func main() {
	dbConnString := os.Getenv("DB_CONNECTION_STRING")
	if dbConnString == "" {
		log.Fatal("DB_CONNECTION_STRING is not set")
	}

	db, err := sqlx.Connect("postgres", dbConnString)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	migrationRunner := migration.NewMigrationRunner(db.DB)
	if err := migrationRunner.RunMigrations("/app/migrations"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS is not set")
	}

	producer := kafka.NewProducer(strings.Split(kafkaBrokers, ","))
	defer producer.Close()

	// Initialize Kafka consumer for payment status updates
	consumer := kafka.NewConsumer(strings.Split(kafkaBrokers, ","), "order-service")
	defer consumer.Close()

	// Initialize repositories
	orderRepo := postgres.NewOrderRepository(db)
	productRepo := postgres.NewProductRepository(db)
	outboxRepo := postgres.NewOutboxRepository(db.DB)

	// Initialize service
	appService := service.New(orderRepo, productRepo, producer, outboxRepo)

	// Initialize status processor
	statusProcessor := service.NewStatusProcessor(orderRepo)

	// OutboxProcessor для transactional outbox
	outboxProcessor := outbox.NewOutboxProcessor(outboxRepo, producer)
	outboxProcessor.Start()

	// Start Kafka consumer for status updates
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			event, err := consumer.ReadOrderStatusUpdatedEvent(ctx)
			if err != nil {
				log.Printf("Error reading order status updated event: %v", err)
				continue
			}

			if err := statusProcessor.ProcessOrderStatusUpdated(ctx, event); err != nil {
				log.Printf("Error processing order status update: %v", err)
			}
		}
	}()

	// Initialize HTTP handler and router
	handler := httptransport.NewHandler(appService)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start server
	go func() {
		log.Println("Starting server on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not listen on %s: %v\n", server.Addr, err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	log.Println("Server gracefully stopped")
}
