package tests

import (
	"context"
	"encoding/json"
	"net/http"

	f "github.com/brianvoe/gofakeit/v6"
	"github.com/sonyamoonglade/poison-tg/internal/api/input"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
)

func (s *AppTestSuite) TestApiAddItem() {
	var (
		require = s.Require()
	)

	s.Run("should add item to catalog with rank 0 because catalog's empty", func() {
		inp := input.AddItemToCatalogInput{
			ImageURLs:       []string{f.URL(), f.URL()},
			AvailableSizes:  []string{f.StreetNumber(), f.StreetNumber()},
			AvailableInCity: []string{f.City(), f.City()},
			Title:           f.BeerName(),
			Quantity:        f.IntRange(1, 10),
			ShopLink:        f.URL(),
			PriceRUB:        234,
		}
		resp, err := s.app.Test(newJsonRequest(http.MethodGet, "/api/catalog/addItem", inp), -1)
		require.NoError(err)
		require.Equal(http.StatusCreated, resp.StatusCode)

		var respJson []domain.CatalogItem
		require.NoError(json.NewDecoder(resp.Body).Decode(&respJson))
		require.True(len(respJson) == 1)
		elem := respJson[0]
		require.Equal(uint(0), elem.Rank)

		// cleanup
		defer func() {
			s.repositories.Catalog.RemoveItem(context.Background(), elem.ItemID)
		}()
	})

	s.Run("should add item to catalog with rank 0 and then next one with 1 sequentially", func() {
		var inputs = []input.AddItemToCatalogInput{
			{
				ImageURLs:       []string{f.URL(), f.URL()},
				AvailableSizes:  []string{f.StreetNumber(), f.StreetNumber()},
				AvailableInCity: []string{f.City(), f.City()},
				Title:           f.Username(),
				Quantity:        f.IntRange(1, 10),
				ShopLink:        f.URL(),
				PriceRUB:        235,
			},
			{
				ImageURLs:       []string{f.URL(), f.URL()},
				AvailableSizes:  []string{f.StreetNumber(), f.StreetNumber()},
				AvailableInCity: []string{f.City(), f.City()},
				Title:           f.Username(),
				Quantity:        f.IntRange(1, 10),
				ShopLink:        f.URL(),
				PriceRUB:        236,
			},
			{
				ImageURLs:       []string{f.URL(), f.URL()},
				AvailableSizes:  []string{f.StreetNumber(), f.StreetNumber()},
				AvailableInCity: []string{f.City(), f.City()},
				Title:           f.Username(),
				Quantity:        f.IntRange(1, 10),
				ShopLink:        f.URL(),
				PriceRUB:        237,
			},
		}
		i1, i2, i3 := inputs[0], inputs[1], inputs[2]

		// Add first
		resp, err := s.app.Test(newJsonRequest(http.MethodGet, "/api/catalog/addItem", i1), -1)
		require.NoError(err)
		require.Equal(http.StatusCreated, resp.StatusCode)

		// Add second
		resp, err = s.app.Test(newJsonRequest(http.MethodGet, "/api/catalog/addItem", i2), -1)
		require.NoError(err)
		require.Equal(http.StatusCreated, resp.StatusCode)

		var respJson []domain.CatalogItem
		// Add third
		resp, err = s.app.Test(newJsonRequest(http.MethodGet, "/api/catalog/addItem", i3), -1)
		require.NoError(err)
		require.Equal(http.StatusCreated, resp.StatusCode)
		require.NoError(json.NewDecoder(resp.Body).Decode(&respJson))

		var elem1, elem2, elem3 domain.CatalogItem
		for _, item := range respJson {
			if item.Title == inputs[0].Title {
				elem1 = item
			}
			if item.Title == inputs[1].Title {
				elem2 = item
			}
			if item.Title == inputs[2].Title {
				elem3 = item
			}
		}
		// Added firstly
		require.Equal(uint(0), elem1.Rank)
		// Added secondly
		require.Equal(uint(1), elem2.Rank)
		require.Equal(uint(2), elem3.Rank)
		require.Equal(3, len(respJson))

		// cleanup
		defer func() {
			s.repositories.Catalog.RemoveItem(context.Background(), elem1.ItemID)
			s.repositories.Catalog.RemoveItem(context.Background(), elem2.ItemID)
			s.repositories.Catalog.RemoveItem(context.Background(), elem3.ItemID)
		}()
	})
}
