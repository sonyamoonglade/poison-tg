package domain

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRemove(t *testing.T) {

	t.Run("remove at 0 idx", func(t *testing.T) {
		cart := make(Cart, 0)
		cart.Add(Product{
			ProductID: primitive.NewObjectID(),
		})
		cart.Add(Product{
			ProductID: primitive.NewObjectID(),
		})
		cart.Add(Product{
			ProductID: primitive.NewObjectID(),
		})

		removeID := cart[0].ProductID

		cart.Remove(removeID.Hex())

		for _, item := range cart {
			if item.ProductID == removeID {
				t.Fatalf("product has not been removed: %s", item.ProductID.Hex())
			}
		}
	})

	t.Run("remove at 1st idx", func(t *testing.T) {
		cart := make(Cart, 0)
		cart.Add(Product{
			ProductID: primitive.NewObjectID(),
		})
		cart.Add(Product{
			ProductID: primitive.NewObjectID(),
		})
		cart.Add(Product{
			ProductID: primitive.NewObjectID(),
		})

		removeID := cart[1].ProductID

		cart.Remove(removeID.Hex())

		for _, item := range cart {
			if item.ProductID == removeID {
				t.Fatalf("product has not been removed: %s", item.ProductID.Hex())
			}
		}

	})

	t.Run("remove at 2nd idx", func(t *testing.T) {
		cart := make(Cart, 0)
		cart.Add(Product{
			ProductID: primitive.NewObjectID(),
		})
		cart.Add(Product{
			ProductID: primitive.NewObjectID(),
		})
		cart.Add(Product{
			ProductID: primitive.NewObjectID(),
		})

		removeID := cart[2].ProductID

		cart.Remove(removeID.Hex())

		for _, item := range cart {
			if item.ProductID == removeID {
				t.Fatalf("product has not been removed: %s", item.ProductID.Hex())
			}
		}
	})

}
