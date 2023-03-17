package repositories

import (
	"context"
	"fmt"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type customerRepo struct {
	customers *mongo.Collection
}

func NewCustomerRepo(db *mongo.Database) Customer {
	return &customerRepo{
		customers: db.Collection("customers"),
	}
}

func (c *customerRepo) Save(ctx context.Context, customer domain.Customer) error {
	if _, err := c.customers.InsertOne(ctx, customer); err != nil {
		return fmt.Errorf("customers.InsertOne: %w", err)
	}
	return nil
}

func (c *customerRepo) UpdateState(ctx context.Context, telegramID int64, newState domain.State) error {
	panic("not implemented") // TODO: Implement
}

func (c *customerRepo) GetByTelegramID(ctx context.Context, telegramID int64) (domain.Customer, error) {
	query := bson.M{"telegramId": telegramID}
	res := c.customers.FindOne(ctx, query)
	if err := res.Err(); err != nil {
		return domain.Customer{}, domain.ErrCustomerNotFound
	}
	var customer domain.Customer
	if err := res.Decode(&customer); err != nil {
		return domain.Customer{}, fmt.Errorf("cant decode customer: %w", err)
	}
	return customer, nil
}

func (c *customerRepo) PrintDb() {}
