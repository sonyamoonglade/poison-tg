package repositories

import (
	"context"
	"errors"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
	"github.com/sonyamoonglade/poison-tg/pkg/nanoid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type orderRepo struct {
	orders *mongo.Collection
}

func NewOrderRepo(orders *mongo.Collection) *orderRepo {
	return &orderRepo{
		orders: orders,
	}
}

func (o *orderRepo) AddComment(ctx context.Context, dto dto.AddCommentDTO) (domain.Order, error) {
	filter := bson.M{"_id": dto.OrderID}
	update := bson.M{"$set": bson.M{"comment": dto.Comment}}
	return o.findOneAndUpdate(ctx, filter, update)
}

func (o *orderRepo) Approve(ctx context.Context, orderID primitive.ObjectID) (domain.Order, error) {
	filter := bson.M{"_id": orderID}
	update := bson.M{"$set": bson.M{"isApproved": true}}
	return o.findOneAndUpdate(ctx, filter, update)
}

func (o *orderRepo) Delete(ctx context.Context, orderID primitive.ObjectID) error {
	filter := bson.M{"_id": orderID}
	_, err := o.orders.DeleteOne(ctx, filter)
	return err
}

func (o *orderRepo) ChangeStatus(ctx context.Context, dto dto.ChangeOrderStatusDTO) (domain.Order, error) {
	filter := bson.M{"_id": dto.OrderID}
	update := bson.M{"$set": bson.M{"status": dto.NewStatus}}
	return o.findOneAndUpdate(ctx, filter, update)
}

func (o *orderRepo) GetAll(ctx context.Context) ([]domain.Order, error) {
	findOpts := options.Find()
	findOpts.SetSort(bson.M{"isApproved": -1})
	res, err := o.orders.Find(ctx, bson.D{}, findOpts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNoOrders
		}
		return nil, err
	}
	var orders []domain.Order
	if err := res.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (o *orderRepo) UpdateToPaid(ctx context.Context, customerID primitive.ObjectID, shortID string) error {
	filter := bson.M{"customer._id": customerID, "shortId": shortID}
	query := bson.M{
		"$set": bson.M{
			"isPaid": true,
		},
	}

	_, err := o.orders.UpdateOne(ctx, filter, query)
	if err != nil {
		return err
	}

	return nil
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

func (o *orderRepo) GetAllForCustomer(ctx context.Context, customerID primitive.ObjectID) ([]domain.Order, error) {
	filter := bson.M{"customer._id": customerID}
	res, err := o.orders.Find(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNoOrders
		}
		return nil, err
	}
	var orders []domain.Order
	if err := res.All(ctx, &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (o *orderRepo) findOneAndUpdate(ctx context.Context, filter, update any) (domain.Order, error) {
	opts := options.FindOneAndUpdate()
	opts.SetReturnDocument(options.After)
	res := o.orders.FindOneAndUpdate(ctx, filter, update, opts)
	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return domain.Order{}, domain.ErrOrderNotFound
		}
		return domain.Order{}, res.Err()
	}
	var ord domain.Order
	if err := res.Decode(&ord); err != nil {
		return domain.Order{}, err
	}
	return ord, nil
}
