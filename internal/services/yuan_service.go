package services

import "math"

type yuanService struct {
}

func NewYuanService() Yuan {
	return &yuanService{}
}

func (y yuanService) GetYuanRate() (float64, error) {
	//TODO implement me
	panic("implement me")
}

const (
	// yuan
	chinaFee = 75
	// rub
	shopFee = 850 - 111
	// rub
	deliveryFee = 35
)

func (y yuanService) ApplyFormula(x_yuan uint64) (x_rub uint64, err error) {
	rate, err := y.getYuanRate()
	if err != nil {
		return 0, err
	}
	v := ((float64(x_yuan) * rate) * 1.09) + (chinaFee * rate) + shopFee + deliveryFee
	rubInt64 := uint64(math.Ceil(v))
	return rubInt64, nil
}

func (y yuanService) getYuanRate() (float64, error) {
	return 11.6, nil
}
