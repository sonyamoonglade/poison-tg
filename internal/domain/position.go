package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Position struct {
	PositionID primitive.ObjectID `json:"positionId,omitempty" bson:"positionId"`
	ShopLink   string             `json:"shopLink" bson:"shopLink"`
	PriceRUB   uint64             `json:"priceRub" bson:"priceRub"`
	PriceYUAN  uint64             `json:"priceYuan" bson:"priceYuan"`
	Size       string             `json:"size" bson:"size"`
}

func NewEmptyPosition() Position {
	return Position{}
}
