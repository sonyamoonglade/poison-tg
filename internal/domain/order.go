package domain

import (
	"errors"

	"github.com/sonyamoonglade/poison-tg/pkg/functools"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
