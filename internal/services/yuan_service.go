package services

import "math"

type yuanService struct {
}

func NewYuanService() YuanService {
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
	shopFee = 875
)

func (y yuanService) ApplyFormula(x_yuan uint64) (x_rub uint64, err error) {
	// Товар (юань) x 11,6 (курс юаня) + 9% + 75 юаней х 11,6 (курс юаня) + 875 руб
	rate, err := y.getYuanRate()
	if err != nil {
		return 0, err
	}
	v := ((float64(x_yuan) * rate) * 1.09) + (chinaFee * rate) + shopFee
	rubInt64 := uint64(math.Ceil(v))
	return rubInt64, nil
}

func (y yuanService) getYuanRate() (float64, error) {
	return 11.6, nil
}