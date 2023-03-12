package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Product struct {
	ProductID primitive.ObjectID `json:"productId,omitempty" bson:"productId"`
	ShopLink  string             `json:"shopLink" bson:"shopLink"`
	ImageURL  string             `json:"imageUrl" bson:"imageUrl"`
	PriceRUB  uint64             `json:"priceRub" bson:"priceRub"`
	PriceYUAN uint64             `json:"priceYuan" bson:"priceYuan"`
	Size      uint8              `json:"size" bson:"size"`
}
