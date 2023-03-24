package services

import (
	"testing"
)

type mockRateProvider struct{}

func (m *mockRateProvider) GetYuanRate() (float64, error) {
	return 6.5, nil
}

func TestApplyFormula(t *testing.T) {
	testCases := []struct {
		description string
		x           uint64
		args        UseFormulaArguments
		result      uint64
	}{}
	_ = testCases
}
