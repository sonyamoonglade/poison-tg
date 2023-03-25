package services

import (
	"math"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
)

type RateProvider interface {
	GetYuanRate() (float64, error)
}

type yuanService struct {
	rateProvider RateProvider
}

func NewYuanService(rateProvider RateProvider) *yuanService {
	return &yuanService{
		rateProvider: rateProvider,
	}
}

func (y yuanService) ApplyFormula(x uint64, args UseFormulaArguments) (uint64, error) {
	rate, err := y.rateProvider.GetYuanRate()
	if err != nil {
		return 0, err
	}

	if args.IsExpress {
		return y.expressFormula(rate, x), nil
	}

	return y.locFormula(rate, args.Location, x), nil
}

func (y yuanService) GetRate() (float64, error) {
	return y.rateProvider.GetYuanRate()
}

const (
	// yuan
	expressOrderFee = 255
	// rub
	expressShopFee = 764
)

func (y yuanService) expressFormula(rate float64, x uint64) uint64 {
	v := ((float64(x) * rate) * 1.09) + (expressOrderFee * rate) + expressShopFee
	return uint64(math.Ceil(v))
}

const (
	// yuan
	normalOrderFee = 75
	// rub
	normalShopFee = 764
	// rub
	spbOrIzhFee = 875 + 200
)

func (y yuanService) locFormula(rate float64, loc domain.Location, x uint64) uint64 {
	var v float64
	if loc == domain.LocationOther {
		v = ((float64(x) * rate) * 1.09) + (normalOrderFee * rate) + normalShopFee
	} else {
		v = ((float64(x) * rate) * 1.09) + (normalOrderFee * rate) + spbOrIzhFee
	}

	return uint64(math.Ceil(v))
}
