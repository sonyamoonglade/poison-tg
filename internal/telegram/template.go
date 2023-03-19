package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

var t = new(templates)

type templates struct {
	Menu                string `json:"menu,omitempty"`
	Start               string `json:"start,omitempty"`
	CartPreviewStartFMT string `json:"cartPreviewStart,omitempty"`
	CartPreviewEndFMT   string `json:"cartPreviewEnd,omitempty"`
	CartPositionFMT     string `json:"cartPosition,omitempty"`
	CalculatorOutput    string `json:"calculatorOutput,omitempty"`
}

func getTemplate() *templates {
	return t
}

func LoadTemplates(path string) error {
	var templates templates

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("can't read file %s: %w", path, err)
	}
	if len(content) < 10 {
		return fmt.Errorf("can't decode file content. File is empty")
	}
	if err := json.NewDecoder(bytes.NewReader(content)).Decode(&templates); err != nil {
		return fmt.Errorf("can't decode file content: %w", err)
	}

	if templates.Menu == "" {
		return fmt.Errorf("missing MENU template")
	}
	if templates.Start == "" {
		return fmt.Errorf("missing START template")
	}
	if templates.CartPreviewStartFMT == "" {
		return fmt.Errorf("missing CART_PREVIEW_START_FMT template")
	}
	if templates.CartPreviewEndFMT == "" {
		return fmt.Errorf("missing CART_PREVIEW_END_FMT template")
	}
	if templates.CartPositionFMT == "" {
		return fmt.Errorf("missing CART_POSITION_FMT template")
	}
	if templates.CalculatorOutput == "" {
		return fmt.Errorf("missing CALCULATOR_OUTPUT template")
	}
	*t = templates
	return nil
}
func getCartPreviewStartTemplate(numPositions int) string {
	return fmt.Sprintf(t.CartPreviewStartFMT, numPositions)
}

type cartPositionPreviewArgs struct {
	n         int
	link      string
	size      string
	priceRub  uint64
	priceYuan uint64
}

func getPositionTemplate(args cartPositionPreviewArgs) string {
	return fmt.Sprintf(t.CartPositionFMT, args.n, args.link, args.size, args.priceRub, args.priceYuan)
}
func getCartPreviewEndTemplate(totalRub uint64, totalYuan uint64) string {
	return fmt.Sprintf(t.CartPreviewEndFMT, totalRub, totalYuan)
}

func getCalculatorOutput(priceForSPBRub, priceForOuterTown uint64) string {
	return fmt.Sprintf(t.CalculatorOutput, priceForSPBRub, priceForOuterTown)
}
