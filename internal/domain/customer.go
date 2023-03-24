package domain

import (
	"errors"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// Default state not waiting for any make order response
	StateDefault                       = State{0}
	StateWaitingForOrderType           = State{1}
	StateWaitingForLocation            = State{2}
	StateWaitingForCalculatorOrderType = State{3}
	StateWaitingForCalculatorLocation  = State{4}
	StateWaitingForSize                = State{5}
	StateWaitingForButton              = State{6}
	StateWaitingForPrice               = State{7}
	StateWaitingForLink                = State{8}
	StateWaitingForCartPositionToEdit  = State{9}
	StateWaitingForCalculatorInput     = State{10}
	StateWaitingForFIO                 = State{11}
	StateWaitingForPhoneNumber         = State{12}
	StateWaitingForDeliveryAddress     = State{13}
)

var (
	ErrCustomerNotFound = errors.New("customer not found")
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

type Meta struct {
	NextOrderType *OrderType `json:"nextOrderType" bson:"nextOrderType"`
	Location      *Location  `json:"location" bson:"location"`
}

type Customer struct {
	CustomerID       primitive.ObjectID `json:"customerId" json:"customerID,omitempty" bson:"_id,omitempty"`
	TelegramID       int64              `json:"telegramID" bson:"telegramId"`
	Username         *string            `json:"username,omitempty" bson:"username,omitempty"`
	FullName         *string            `json:"fullName,omitempty" bson:"fullName,omitempty"`
	PhoneNumber      *string            `json:"phoneNumber,omitempty" bson:"phoneNumber,omitempty"`
	TgState          State              `json:"state" bson:"state"`
	Cart             Cart               `json:"cart" bson:"cart"`
	Meta             Meta               `json:"meta" bson:"meta"`
	CalculatorMeta   Meta               `json:"calculatorMeta" bson:"calculatorMeta"`
	CatalogOffset    uint               `json:"catalogOffset" bson:"catalogOffset"`
	LastEditPosition *Position          `json:"lastEditPosition,omitempty" bson:"lastEditPosition"`
}

func NewCustomer(telegramID int64, username string) Customer {
	return Customer{
		TelegramID: telegramID,
		Username:   &username,
		TgState:    StateDefault,
	}
}

func (c *Customer) UpdateState(newState State) {
	c.TgState = newState
}
func (c *Customer) GetTgState() uint8 {
	return c.TgState.V
}

func (c *Customer) SetLastEditPosition(p Position) {
	c.LastEditPosition = &p
}

func (c *Customer) UpdateLastEditPositionSize(s string) {
	c.LastEditPosition.Size = s
}

func (c *Customer) UpdateLastEditPositionPrice(priceRub uint64, priceYuan uint64) {
	c.LastEditPosition.PriceRUB = priceRub
	c.LastEditPosition.PriceYUAN = priceYuan

}

func (c *Customer) UpdateLastEditPositionLink(link string) {
	c.LastEditPosition.ShopLink = link
}

func (c *Customer) UpdateLastEditPositionButtonColor(button Button) {
	c.LastEditPosition.Button = button
}

func (c *Customer) UpdateMetaOrderType(typ OrderType) {
	c.Meta.NextOrderType = &typ
}

func (c *Customer) UpdateMetaLocation(loc Location) {
	c.Meta.Location = &loc
}

func (c *Customer) UpdateCalculatorMetaOrderType(typ OrderType) {
	c.CalculatorMeta.NextOrderType = &typ
}

func (c *Customer) UpdateCalculatorMetaLocation(loc Location) {
	c.CalculatorMeta.Location = &loc
}

func (c *Customer) IncrementCatalogOffset() {
	c.CatalogOffset++
}

func (c *Customer) NullifyCatalogOffset() {
	c.CatalogOffset = 0
}

func MakeUsername(firstName string, lastName string, username string) string {
	var out string
	if firstName == "" || lastName == "" {
		out = username
	} else if firstName != "" && lastName != "" {
		out = firstName + " " + lastName
	}
	return out
}

func IsValidFullName(fullName string) bool {
	spaceCount := strings.Count(fullName, " ")
	if spaceCount != 2 {
		return false
	}
	return true
}

var r = regexp.MustCompile(`(?m)^(\+7|7|8)?[\s\-]?\(?[489][0-9]{2}\)?[\s\-]?[0-9]{3}[\s\-]?[0-9]{2}[\s\-]?[0-9]{2}$`)

func IsValidPhoneNumber(phoneNumber string) bool {
	return r.MatchString(phoneNumber)
}

type State struct {
	V uint8 `json:"v" bson:"v"`
}

func (s State) Value() uint8 {
	return s.V
}
