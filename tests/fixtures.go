package tests

import (
	f "github.com/brianvoe/gofakeit/v6"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
)

func catalogItemFixture() domain.CatalogItem {
	return domain.CatalogItem{
		ImageURLs:       []string{f.BeerName()},
		AvailableSizes:  []string{f.City(), f.City()},
		AvailableInCity: []string{f.City(), f.City()},
		Quantity:        f.IntRange(1, 10),
		Title:           f.Word(),
		ShopLink:        f.URL(),
		PriceRUB:        uint64(f.IntRange(1, 15)),
	}
}
