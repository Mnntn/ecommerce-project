package service

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/mnntn/ecommerce-project/order-service/internal/domain"
	"github.com/mnntn/ecommerce-project/order-service/internal/repository"
)

type StatusProcessor struct {
	orderRepo repository.OrderRepository
}

func NewStatusProcessor(orderRepo repository.OrderRepository) *StatusProcessor {
	return &StatusProcessor{
		orderRepo: orderRepo,
	}
}

func (p *StatusProcessor) ProcessOrderStatusUpdated(ctx context.Context, event *domain.OrderStatusUpdatedEvent) error {
	log.Printf("Processing order status update: OrderID=%s, Status=%s, Reason=%s",
		event.OrderID, event.Status, event.Reason)

	orderID, err := uuid.Parse(event.OrderID)
	if err != nil {
		log.Printf("Invalid order ID format: %s", event.OrderID)
		return err
	}

	var status domain.OrderStatus
	switch event.Status {
	case "FINISHED":
		status = domain.StatusFinished
	case "CANCELLED":
		status = domain.StatusCancelled
	default:
		log.Printf("Unknown status: %s", event.Status)
		return nil
	}

	err = p.orderRepo.UpdateStatus(ctx, orderID, status)
	if err != nil {
		log.Printf("Failed to update order status: %v", err)
		return err
	}

	log.Printf("Successfully updated order %s status to %s", orderID, status)
	return nil
}
