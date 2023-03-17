package services

import (
	"context"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories"
)

type customerService struct {
	customerRepo repositories.Customer
}

func NewCustomerService(customerRepo repositories.Customer) Customer {
	return &customerService{
		customerRepo: customerRepo,
	}
}

func (c *customerService) Save(ctx context.Context, customer domain.Customer) error {
	return c.customerRepo.Save(ctx, customer)
}

func (c *customerService) GetByTelegramID(ctx context.Context, telegramID int64) (domain.Customer, error) {
	return c.customerRepo.GetByTelegramID(ctx, telegramID)
}

func (c *customerService) UpdateState(ctx context.Context, telegramID int64, newState domain.State) error {
	return c.customerRepo.UpdateState(ctx, telegramID, newState)
}

func (c *customerService) PrintDb() {
	c.customerRepo.PrintDb()
}
