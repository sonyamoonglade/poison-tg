package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Order struct {
	OrderID      primitive.ObjectID `json:"orderID,omitempty" bson:"_id,omitempty"`
	Customer     Customer           `json:"customer" bson:"customer"`
	Cart         Cart               `json:"cart" bson:"cart"`
	AmountRUB    uint64             `json:"amountRub" bson:"amountRub"`
	CDEK_Address string             `json:"cdekAddress" bson:"cdekAddress"`
	IsPaid       bool               `json:"isPaid" bson:"isPaid"`
	IsApproved   bool               `json:"isApproved" bson:"isApproved"`
}
