package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/mnntn/ecommerce-project/payment-service/internal/kafka"
	"github.com/mnntn/ecommerce-project/payment-service/internal/migration"
	"github.com/mnntn/ecommerce-project/payment-service/internal/repository/postgres"
	"github.com/mnntn/ecommerce-project/payment-service/internal/service"
	phttp "github.com/mnntn/ecommerce-project/payment-service/internal/transport/http"
)

func main() {
	// Initialize Kafka producer
	producer := kafka.NewProducer([]string{os.Getenv("KAFKA_BROKERS")})
	defer producer.Close()

	// Initialize Kafka consumer
	consumer := kafka.NewConsumer([]string{os.Getenv("KAFKA_BROKERS")}, "payment-service")
	defer consumer.Close()

	// Подключение к БД
	db, err := sql.Open("postgres", os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// Run migrations
	migrationRunner := migration.NewMigrationRunner(db)
	if err := migrationRunner.RunMigrations("/app/migrations"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Репозитории
	userRepo := postgres.NewUserRepository(db)
	accountRepo := postgres.NewAccountRepository(db)
	inboxRepo := postgres.NewInboxRepository(db)
	outboxRepo := postgres.NewOutboxRepository(db)

	// Сервис аккаунтов
	accountService := service.NewAccountService(accountRepo, userRepo)

	// Обработчик заказов с transactional inbox/outbox
	orderProcessor := service.NewOrderProcessor(accountRepo, producer, inboxRepo, outboxRepo, db)

	// OutboxProcessor для transactional outbox
	outboxProcessor := postgres.NewOutboxProcessor(outboxRepo, producer)
	outboxProcessor.Start()

	// Start message processing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Kafka consumer stopped due to context cancellation")
				return
			default:
				event, err := consumer.ReadOrderCreatedEvent(ctx)
				if err != nil {
					if ctx.Err() != nil {
						log.Println("Kafka consumer stopped due to context cancellation")
						return
					}
					log.Printf("Error reading order created event: %v", err)
					continue
				}

				if err := orderProcessor.ProcessOrderCreated(ctx, event); err != nil {
					log.Printf("Error processing order created: %v", err)
				}
			}
		}
	}()

	// HTTP handler
	handler := phttp.NewHandler(accountService)
	r := mux.NewRouter()
	handler.RegisterRoutes(r)

	go func() {
		log.Println("Payment Service HTTP server started on :8080")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}
