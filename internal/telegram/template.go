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

const (
	yes string = "✅"
	no         = "❌"
)

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
	AfterPaid           string `json:"afterPaid,omitempty"`
	Requisites          string `json:"requisites,omitempty"`
	GuideStep1          string `json:"guide_step1,omitempty"`
	GuideStep2          string `json:"guide_step2,omitempty"`
	GuideStep3          string `json:"guide_step3,omitempty"`
	GuideStep4          string `json:"guide_step4,omitempty"`
	GuideStep5          string `json:"guide_step5,omitempty"`
	GuideStep6          string `json:"guide_step6,omitempty"`
	MyOrdersStart       string `json:"myOrdersStart,omitempty"`
	MyOrdersEnd         string `json:"myOrdersEnd,omitempty"`
	SingleOrderPreview  string `json:"singleOrderPreview,omitempty"`
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

func getCalculatorOutput(price uint64) string {
	return fmt.Sprintf(t.CalculatorOutput, price)
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

func getCatalog(username string) string {
	return fmt.Sprintf(t.Catalog, username)
}

func getAfterPaid(fullname, shortOrderID string) string {
	return fmt.Sprintf(t.AfterPaid, fullname, shortOrderID)
}

func getMyOrdersStart(fullname string) string {
	return fmt.Sprintf(t.MyOrdersStart, fullname)
}

type singleOrderArgs struct {
	shortID                       string
	isExpress, isPaid, isApproved bool
	cartLen                       int
	deliveryAddress               string
	totalYuan                     uint64
	totalRub                      uint64
}

func getSingleOrderPreview(args singleOrderArgs) string {
	var (
		expressStr  string
		paidStr     string
		approvedStr string
	)
	if args.isExpress {
		expressStr = "Экспресс"
	} else {
		expressStr = "Обычный"
	}

	if args.isPaid {
		paidStr = yes
	} else {
		paidStr = no
	}

	if args.isApproved {
		approvedStr = yes
	} else {
		approvedStr = no
	}

	return fmt.Sprintf(t.SingleOrderPreview, args.shortID, paidStr, approvedStr, expressStr, args.deliveryAddress, args.cartLen, args.totalYuan, args.totalRub)
}
