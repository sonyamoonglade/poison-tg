package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Customer struct {
	CustomerID  primitive.ObjectID `json:"customerId" json:"customerID,omitempty" bson:"customerId,omitempty"`
	TelegramID  string             `json:"telegramID" bson:"telegramID"`
	Username    string             `json:"username" bson:"username"`
	PhoneNumber *string            `json:"phoneNumber,omitempty" bson:"phoneNumber,omitempty"`
}
