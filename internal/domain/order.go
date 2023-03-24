package domain

import (
	"errors"

	"github.com/sonyamoonglade/poison-tg/pkg/functools"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Status int

const (
	StatusDefault Status = iota + 1
	StatusConfirmed
	StatusGot
	StatusTracking
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrNoOrders      = errors.New("no orders")
)

type Order struct {
	OrderID         primitive.ObjectID `json:"orderID,omitempty" bson:"_id,omitempty"`
	ShortID         string             `json:"shortId" bson:"shortId"`
	Customer        Customer           `json:"customer" bson:"customer"`
	Cart            Cart               `json:"cart" bson:"cart"`
	AmountRUB       uint64             `json:"amountRub" bson:"amountRub"`
	AmountYUAN      uint64             `json:"amountYuan" bson:"amountYuan"`
	DeliveryAddress string             `json:"deliveryAddress" bson:"deliveryAddress"`
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
		Status:          StatusDefault,
	}
}
