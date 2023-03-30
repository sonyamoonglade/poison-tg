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
	StateWaitingForCategory            = State{3}
	StateWaitingForCalculatorCategory  = State{4}
	StateWaitingForCalculatorOrderType = State{5}
	StateWaitingForCalculatorLocation  = State{6}
	StateWaitingForSize                = State{7}
	StateWaitingForButton              = State{8}
	StateWaitingForPrice               = State{9}
	StateWaitingForLink                = State{10}
	StateWaitingForCartPositionToEdit  = State{11}
	StateWaitingForCalculatorInput     = State{12}
	StateWaitingForFIO                 = State{13}
	StateWaitingForPhoneNumber         = State{14}
	StateWaitingForDeliveryAddress     = State{15}
)

var (
	ErrCustomerNotFound = errors.New("customer not found")
)

type Meta struct {
	NextOrderType *OrderType `json:"nextOrderType" bson:"nextOrderType"`
	Location      *Location  `json:"location" bson:"location"`
}

type CalculatorMeta struct {
	NextOrderType *OrderType `json:"nextOrderType" bson:"nextOrderType"`
	Location      *Location  `json:"location" bson:"location"`
	Category      *Category  `bson:"category"`
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
	CalculatorMeta   CalculatorMeta     `json:"calculatorMeta" bson:"calculatorMeta"`
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

func (c *Customer) UpdateLastEditPositionCategory(cat Category) {
	if c.LastEditPosition == nil {
		c.LastEditPosition = &Position{}
	}
	c.LastEditPosition.Category = cat
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

func (c *Customer) UpdateCalculatorMetaCategory(cat Category) {
	c.CalculatorMeta.Category = &cat
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

const DefaultUsername = "User"

func MakeUsername(username string) string {
	if username == "" {
		return DefaultUsername
	}
	return username
}

func IsValidFullName(fullName string) bool {
	spaceCount := strings.Count(fullName, " ")
	if spaceCount != 2 {
		return false
	}
	return true
}

var r = regexp.MustCompile(`^(8|7)((\d{10})|(\s\(\d{3}\)\s\d{3}\s\d{2}\s\d{2}))`)

func IsValidPhoneNumber(phoneNumber string) bool {
	if len(phoneNumber) != 11 {
		return false
	}
	return r.MatchString(phoneNumber)
}

type State struct {
	V uint8 `json:"v" bson:"v"`
}

func (s State) Value() uint8 {
	return s.V
}
