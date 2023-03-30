package domain

import (
	"errors"
	"math"

	"github.com/sonyamoonglade/poison-tg/pkg/functools"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderType int

const (
	OrderTypeExpress OrderType = iota + 1
	OrderTypeNormal
)

type Location int

const (
	LocationSPB Location = iota + 1
	LocationIZH
	LocationOther
)

type Status int

const (
	StatusNotApproved Status = iota + 1
	StatusApproved
	StatusBuyout
	StatusTransferToPoison
	StatusSentFromPoison
	StatusGotToRussia
	StatusCheckTrack
	StatusGotToOrdererCity
)

var StatusTexts = map[Status]string{
	StatusNotApproved:      "Не подтвержден",
	StatusApproved:         "Подтвержден",
	StatusBuyout:           "Выкуплен",
	StatusTransferToPoison: "Передан на склад POIZON",
	StatusSentFromPoison:   "Отправлен со склада POIZON в Россию",
	StatusGotToRussia:      "Пришл на склад распределения",
	StatusCheckTrack:       "Трэк номер",
	StatusGotToOrdererCity: "Пришел в город назначения",
}

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrNoOrders      = errors.New("no orders")
)

type Order struct {
	OrderID         primitive.ObjectID `json:"orderId,omitempty" bson:"_id,omitempty"`
	ShortID         string             `json:"shortId" bson:"shortId"`
	Customer        Customer           `json:"customer" bson:"customer"`
	Cart            Cart               `json:"cart" bson:"cart"`
	AmountRUB       uint64             `json:"amountRub" bson:"amountRub"`
	AmountYUAN      uint64             `json:"amountYuan" bson:"amountYuan"`
	DeliveryAddress string             `json:"deliveryAddress" bson:"deliveryAddress"`
	Comment         *string            `json:"comment" bson:"comment"`
	IsPaid          bool               `json:"isPaid" bson:"isPaid"`
	IsApproved      bool               `json:"isApproved" bson:"isApproved"`
	IsExpress       bool               `json:"isExpress" bson:"isExpress"`
	Status          Status             `json:"status" bson:"status"`
}

func NewOrder(customer Customer, deliveryAddress string, isExpress bool, shortID string) Order {
	type total struct {
		rub, yuan uint64
	}

	totals := functools.Reduce(func(t total, cartItem Position) total {
		t.yuan += cartItem.PriceYUAN
		t.rub += cartItem.PriceRUB
		return t
	}, customer.Cart, total{})

	return Order{
		Customer:        customer,
		ShortID:         shortID,
		Cart:            customer.Cart,
		AmountRUB:       totals.rub,
		AmountYUAN:      totals.yuan,
		DeliveryAddress: deliveryAddress,
		IsPaid:          false,
		IsExpress:       isExpress,
		IsApproved:      false,
		Status:          StatusNotApproved,
	}
}

func IsValidOrderStatus(s Status) bool {
	_, ok := StatusTexts[s]
	return ok
}

type formula func(x uint64, rate float64) (rub uint64)

type FormulaMap = map[OrderType]map[Location]map[Category]formula

const (
	othMul   = 0.5
	lightMul = 1.6
	heavyMul = 2.6
)

var formulas = FormulaMap{
	OrderTypeExpress: {
		LocationOther: {
			CategoryOther: expressfn(othMul, 764),
			CategoryLight: expressfn(lightMul, 764),
			CategoryHeavy: expressfn(heavyMul, 764),
		},
		LocationIZH: {
			CategoryOther: expressfn(othMul, 764),
			CategoryLight: expressfn(lightMul, 764),
			CategoryHeavy: expressfn(heavyMul, 764),
		},
		LocationSPB: {
			CategoryOther: expressfn(othMul, 764),
			CategoryLight: expressfn(lightMul, 764),
			CategoryHeavy: expressfn(heavyMul, 764),
		},
	},
	OrderTypeNormal: {
		LocationOther: {
			CategoryOther: normalfn(othMul, 764),
			CategoryLight: normalfn(lightMul, 764),
			CategoryHeavy: normalfn(heavyMul, 764),
		},
		LocationIZH: {
			CategoryOther: normalfn(othMul, 1075),
			CategoryLight: normalfn(lightMul, 1075),
			CategoryHeavy: normalfn(heavyMul, 1075),
		},
		LocationSPB: {
			CategoryOther: normalfn(othMul, 1075),
			CategoryLight: normalfn(lightMul, 1075),
			CategoryHeavy: normalfn(heavyMul, 1075),
		},
	},
}

func expressfn(kg_mul float64, fee float64) formula {
	return func(x uint64, rate float64) (rub uint64) {
		v := (float64(x)*rate)*1.09 + (170.0 * kg_mul * rate) + fee
		return uint64(math.Ceil(v))
	}
}

func normalfn(kg_mul float64, fee float64) formula {
	return func(x uint64, rate float64) (rub uint64) {
		v := (float64(x)*rate)*1.09 + (50.0 * kg_mul * rate) + fee
		return uint64(math.Ceil(v))
	}
}

type ConvertYuanArgs struct {
	X         uint64
	Rate      float64
	OrderType OrderType
	Location  Location
	Category  Category
}

func ConvertYuan(args ConvertYuanArgs) (rub uint64) {
	return formulas[args.OrderType][args.Location][args.Category](args.X, args.Rate)
}
