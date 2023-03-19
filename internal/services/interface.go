package services

import (
	"context"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
)

type Yuan interface {
	GetYuanRate() (float64, error)
	// ApplyFormula math.Ceil the value up
	// Accepts price in yuan, returns price in rub
	ApplyFormula(x_yuan uint64) (x_rub uint64, err error)
}

type Order interface {
	CreateNewOrder(ctx context.Context, cart domain.Cart, customer domain.Customer)
}
