package repositories

import (
	"context"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Customer interface {
	Save(ctx context.Context, c domain.Customer) error
	GetByTelegramID(ctx context.Context, telegramID int64) (domain.Customer, error)
	UpdateState(ctx context.Context, telegramID int64, newState domain.State) error
	Update(ctx context.Context, customerID primitive.ObjectID, dto dto.UpdateCustomerDTO) error
	PrintDb()
}

type Order interface {
	GetByShortID(ctx context.Context, shortID string) (domain.Order, error)
	Save(ctx context.Context, o domain.Order) error
}
