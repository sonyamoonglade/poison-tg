package repositories

import (
	"context"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
)

type Customer interface {
	Save(ctx context.Context, c domain.Customer) error
	GetByTelegramID(ctx context.Context, telegramID int64) (domain.Customer, error)
	UpdateState(ctx context.Context, telegramID int64, newState domain.State) error
	PrintDb()
}
