package repositories

import (
	"context"
	"errors"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/pkg/nanoid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type orderRepo struct {
	orders *mongo.Collection
}

func NewOrderRepo(orders *mongo.Collection) *orderRepo {
	return &orderRepo{
		orders: orders,
	}
}

func (o *orderRepo) Save(ctx context.Context, order domain.Order) error {
	_, err := o.orders.InsertOne(ctx, order)
	if err != nil {
		return err
	}
	return nil
}

func (o *orderRepo) GetFreeShortID(ctx context.Context) (string, error) {
	for {
		shortID := nanoid.GenerateNanoID()
		res := o.orders.FindOne(ctx, bson.M{"shortId": shortID})
		err := res.Err()
		if err != nil {
			// Found free shortID
			if errors.Is(err, mongo.ErrNoDocuments) {
				return shortID, nil
			}
			return "", err
		}
		// if reached, means something has been found - skip and go again
		continue
	}
}

func (o *orderRepo) GetByShortID(ctx context.Context, shortID string) (domain.Order, error) {
	res := o.orders.FindOne(ctx, bson.M{"shortId": shortID})
	err := res.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Order{}, domain.ErrOrderNotFound
		}
		return domain.Order{}, err
	}
	var ord domain.Order
	if err := res.Decode(&ord); err != nil {
		return domain.Order{}, err
	}
	return ord, nil
}