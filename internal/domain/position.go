package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Button string

const (
	ButtonTorqoise Button = "Бирюзовый"
	ButtonGrey            = "Серый"
	Button95              = "95 БУ"
)

type Position struct {
	PositionID primitive.ObjectID `json:"positionId,omitempty" bson:"_id,omitempty"`
	ShopLink   string             `json:"shopLink" bson:"shopLink"`
	PriceRUB   uint64             `json:"priceRub" bson:"priceRub"`
	PriceYUAN  uint64             `json:"priceYuan" bson:"priceYuan"`
	Button     Button             `json:"button" bson:"button"`
	Size       string             `json:"size" bson:"size"`
}

func NewEmptyPosition() Position {
	return Position{}
}
