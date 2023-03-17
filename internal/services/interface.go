package services

type YuanService interface {
	GetYuanRate() (float64, error)
	// ApplyFormula math.Ceil the value up
	// Accepts price in yuan, returns price in rub
	ApplyFormula(x_yuan uint64) (x_rub uint64, err error)
}
