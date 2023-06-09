package domain

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/sonyamoonglade/poison-tg/pkg/functools"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrNoCatalog    = errors.New("catalog not found")
	ErrItemNotFound = errors.New("item not found")
)

type CatalogItem struct {
	ItemID          primitive.ObjectID `json:"itemId,omitempty" bson:"_id,omitempty"`
	ImageURLs       []string           `json:"imageUrls" bson:"imageUrls"`
	AvailableSizes  []string           `json:"availableSizes" bson:"availableSizes"`
	AvailableInCity []string           `json:"availableInCity" bson:"availableInCity"`
	Quantity        int                `json:"quantity" bson:"quantity"`
	Title           string             `json:"title" bson:"title"`
	ShopLink        string             `json:"shopLink" bson:"shopLink"`
	Rank            uint               `json:"rank" bson:"rank"`
	PriceRUB        uint64             `json:"priceRub" bson:"priceRub"`
}

func (c *CatalogItem) GetCaption() string {
	template := "Товар: <a href=\"%s\">%s</a>\n" +
		"Размер(ы): %s\n" +
		"Есть в городе: %s\n" +
		"Количество товара: %d\n\n" +
		"Стоимость в рублях: %d ₽"
	return fmt.Sprintf(template, c.ShopLink, c.Title, c.getSizes(), c.getCities(), c.Quantity, c.PriceRUB)
}

func (c *CatalogItem) getSizes() string {
	var out string
	for i, size := range c.AvailableSizes {
		// last
		if i == len(c.AvailableSizes)-1 {
			out += fmt.Sprintf("(%s)", size)
			continue
		}
		out += fmt.Sprintf("(%s); ", size)
	}
	return out
}

func (c *CatalogItem) getCities() string {
	return strings.Join(c.AvailableInCity, "; ")
}

// catalog must be sorted by rank ascending
func UpdateRanks(catalog []CatalogItem) []CatalogItem {
	if catalog == nil {
		return nil
	}
	// If first item's rank is 0 then down all subsequent
	if catalog[0].Rank != uint(0) && catalog[0].Rank > uint(0) {
		return functools.Map(func(item CatalogItem, i int) CatalogItem {
			item.Rank--
			return item
		}, catalog)
	}

	// Found gap somewhere in between (only one at a time)
	var idxGap int
	for i := 0; i < len(catalog)-1; i++ {
		curr, next := catalog[i], catalog[i+1]
		if math.Abs(float64(curr.Rank)-float64(next.Rank)) > 1 {
			idxGap = i + 1
			break
		}
	}
	// All fine
	if idxGap == 0 {
		return catalog
	}
	return functools.Map(func(item CatalogItem, i int) CatalogItem {
		if i >= idxGap {
			item.Rank--
			return item
		}
		return item
	}, catalog)
}
