package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
)

var t = new(templates)

type templates struct {
	Menu                string `json:"menu,omitempty"`
	Start               string `json:"start,omitempty"`
	Catalog             string `json:"catalog,omitempty"`
	CartPreviewStartFMT string `json:"cartPreviewStart,omitempty"`
	CartPreviewEndFMT   string `json:"cartPreviewEnd,omitempty"`
	CartPositionFMT     string `json:"cartPosition,omitempty"`
	CalculatorOutput    string `json:"calculatorOutput,omitempty"`
	OrderStart          string `json:"order,omitempty"`
	OrderEnd            string `json:"orderEnd,omitempty"`
	Requisites          string `json:"requisites,omitempty"`
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

	v := reflect.ValueOf(&templates).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Interface() == "" {
			return fmt.Errorf("missing %s template", v.Type().Field(i).Name)
		}
	}

	*t = templates
	return nil
}

func getCartPreviewStartTemplate(numPositions int, isExpress bool) string {
	var orderTypeText string
	if isExpress {
		orderTypeText = "Экспресс"
	} else {
		orderTypeText = "Обычный"
	}
	return fmt.Sprintf(t.CartPreviewStartFMT, numPositions, orderTypeText)
}

type cartPositionPreviewArgs struct {
	n         int
	link      string
	size      string
	priceRub  uint64
	priceYuan uint64
}

func getPositionTemplate(args cartPositionPreviewArgs) string {
	if args.size == "#" {
		args.size = "без размера"
	}
	return fmt.Sprintf(t.CartPositionFMT, args.n, args.link, args.size, args.priceRub, args.priceYuan)
}
func getCartPreviewEndTemplate(totalRub uint64, totalYuan uint64) string {
	return fmt.Sprintf(t.CartPreviewEndFMT, totalRub, totalYuan)
}

func getCalculatorOutput(priceForSPBRub, priceForOuterTown uint64) string {
	return fmt.Sprintf(t.CalculatorOutput, priceForSPBRub, priceForOuterTown)
}

type orderStartArgs struct {
	fullName        string
	shortOrderID    string
	phoneNumber     string
	isExpress       bool
	deliveryAddress string
	nCartItems      int
}

func getOrderStart(args orderStartArgs) string {
	var expressStr string
	if args.isExpress {
		expressStr = "Экспресс"
	} else {
		expressStr = "Обычный"
	}

	return fmt.Sprintf(t.OrderStart, args.fullName, args.shortOrderID, expressStr, args.fullName, args.phoneNumber, args.deliveryAddress, args.nCartItems)
}

func getOrderEnd(amountRub uint64) string {
	return fmt.Sprintf(t.OrderEnd, amountRub)
}

func getRequisites(reqs domain.Requisites, shortOrderID string) string {
	return fmt.Sprintf(t.Requisites, shortOrderID, reqs.SberID, reqs.TinkoffID, shortOrderID)
}

func extractShortOrderIDFromRequisites(text string) string {
	var (
		openIdx  int
		closeIdx int
	)
	for i, ch := range text {
		if ch == '[' {
			openIdx = i
			continue
		}
		if ch == ']' {
			closeIdx = i
			break
		}
	}
	return text[openIdx+1 : closeIdx]
}

func getCatalog(username string) string {
	return fmt.Sprintf(t.Catalog, username)
}
