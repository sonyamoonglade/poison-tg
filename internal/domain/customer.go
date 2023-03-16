package domain

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// Default state not waiting for any make order response
	StateDefault               = State{V: 0}
	StateWaitingForSize        = State{V: 1}
	StateWaitingForPrice       = State{V: 2}
	StateWaitingForButtonColor = State{V: 3}
	StateWaitingForLink        = State{V: 4}
)

var (
	ErrCustomerNotFound = errors.New("customer not found")
)

type Customer struct {
	CustomerID  primitive.ObjectID `json:"customerId" json:"customerID,omitempty" bson:"customerId,omitempty"`
	TelegramID  int64              `json:"telegramID" bson:"telegramID"`
	Username    string             `json:"username" bson:"username"`
	PhoneNumber *string            `json:"phoneNumber,omitempty" bson:"phoneNumber,omitempty"`
	TgState     State              `json:"state" bson:"state"`
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
func (c Customer) GetTgState() uint8 {
	return c.TgState.V
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

type State struct {
	V uint8 `json:"v" bson:"v"`
}

func (s State) Value() uint8 {
	return s.V
}
