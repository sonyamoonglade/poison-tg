package domain

import (
	"errors"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// Default state not waiting for any make order response
	StateDefault                      = State{0}
	StateWaitingForSize               = State{1}
	StateWaitingForButton             = State{2}
	StateWaitingForPrice              = State{3}
	StateWaitingForLink               = State{4}
	StateWaitingForCartPositionToEdit = State{5}
	StateWaitingForCalculatorInput    = State{6}
	StateWaitingForFIO                = State{7}
	StateWaitingForPhoneNumber        = State{8}
	StateWaitingForDeliveryAddress    = State{9}
)

var (
	ErrCustomerNotFound = errors.New("customer not found")
)

type Customer struct {
	CustomerID       primitive.ObjectID `json:"customerId" json:"customerID,omitempty" bson:"_id,omitempty"`
	TelegramID       int64              `json:"telegramID" bson:"telegramId"`
	Username         string             `json:"username" bson:"username"`
	FullName         *string            `json:"fullName,omitempty" bson:"fullName,omitempty"`
	PhoneNumber      *string            `json:"phoneNumber,omitempty" bson:"phoneNumber,omitempty"`
	TgState          State              `json:"state" bson:"state"`
	Cart             Cart               `json:"cart" bson:"cart"`
	LastEditPosition *Position          `json:"lastEditPosition,omitempty" bson:"lastEditPosition"`
}

func NewCustomer(telegramID int64, username string) Customer {
	return Customer{
		TelegramID: telegramID,
		Username:   username,
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
