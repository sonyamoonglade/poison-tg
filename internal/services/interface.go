package services

import (
	"context"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
)

type UseFormulaArguments struct {
	Location  domain.Location
	IsExpress bool
}

type Yuan interface {
	// ApplyFormula Accepts price in yuan, returns rounded to up price in rub
	ApplyFormula(yuanAmount uint64, args UseFormulaArguments) (uint64, error)
	GetRate() (float64, error)
}

type Order interface {
	CreateNewOrder(ctx context.Context, cart domain.Cart, customer domain.Customer)
}
