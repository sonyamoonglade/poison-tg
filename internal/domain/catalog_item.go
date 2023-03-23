package domain

import (
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrNoCatalog = errors.New("catalog not found")
)

type CatalogItem struct {
	ItemID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ImageURLs      []string           `json:"imageUrls" bson:"imageUrls"`
	Title          string             `json:"title" bson:"title"`
	Rank           uint               `json:"rank" bson:"rank"`
	AvailableSizes []string           `json:"availableSizes" bson:"availableSizes"`
}

func (c *CatalogItem) GetCaption() string {
	return fmt.Sprintf("%s\n\n%s", c.getTitleText(), c.getSizesText())
}

func (c *CatalogItem) getSizesText() string {
	return "Доступные размеры: [" + strings.Join(c.AvailableSizes, ", ") + "]"
}

func (c *CatalogItem) getTitleText() string {
	return "Товар: " + c.Title
}
