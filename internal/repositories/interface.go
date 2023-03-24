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
	GetFreeShortID(ctx context.Context) (string, error)
	GetAll(ctx context.Context, customerID primitive.ObjectID) ([]domain.Order, error)
	UpdateToPaid(ctx context.Context, customerID primitive.ObjectID, shortID string) error
	Save(ctx context.Context, o domain.Order) error
}

type Business interface {
	GetRequisites(ctx context.Context) (domain.Requisites, error)
}

type Catalog interface {
	// When catalog is going to be inserted this function is going to be executed
	OnChange(func(newItems []domain.CatalogItem))
	GetCatalog(ctx context.Context) ([]domain.CatalogItem, error)
	AddItem(ctx context.Context, item domain.CatalogItem) error
	RemoveItem(ctx context.Context, itemID primitive.ObjectID) error
	UpdateRanks(ctx context.Context, dto dto.UpdateItemDTO) error
}
