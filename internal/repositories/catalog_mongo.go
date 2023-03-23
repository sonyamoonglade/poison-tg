package repositories

import (
	"context"
	"errors"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
	"github.com/sonyamoonglade/poison-tg/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type OnChangeFunc func(items []domain.CatalogItem)

type catalogRepo struct {
	catalog  *mongo.Collection
	onChange OnChangeFunc
}

func NewCatalogRepo(catalog *mongo.Collection, onChangeFunc OnChangeFunc) *catalogRepo {
	return &catalogRepo{
		catalog:  catalog,
		onChange: onChangeFunc,
	}
}

func (c *catalogRepo) GetCatalog(ctx context.Context) ([]domain.CatalogItem, error) {
	opts := options.Find()
	opts.SetSort(bson.M{"rank": 1})
	res, err := c.catalog.Find(ctx, bson.D{}, opts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNoCatalog
		}
		return nil, err
	}

	var catalog []domain.CatalogItem
	defer res.Close(ctx)
	if err := res.All(ctx, &catalog); err != nil {
		return nil, err
	}

	return catalog, nil
}

func (c *catalogRepo) AddItem(ctx context.Context, item domain.CatalogItem) error {
	_, err := c.catalog.InsertOne(ctx, item)
	if err != nil {
		return err
	}
	defer func() {
		// Notify on change
		newCatalog, err := c.GetCatalog(ctx)
		if err != nil {
			logger.Get().Error("deferred catalog notify", zap.Error(err))
			return
		}
		c.onChange(newCatalog)
	}()
	return nil
}

func (c *catalogRepo) RemoveItem(ctx context.Context, itemID primitive.ObjectID) error {
	_, err := c.catalog.DeleteOne(ctx, bson.M{"_id": itemID})
	return err
}

func (c *catalogRepo) UpdateRanks(ctx context.Context, dto dto.UpdateItemDTO) error {
	client := c.catalog.Database().Client()

	if err := client.UseSession(ctx, func(tx mongo.SessionContext) error {
		if err := tx.StartTransaction(); err != nil {
			return err
		}

		// Increment rank by id
		incQuery := bson.E{
			Key: "$inc",
			Value: bson.E{
				Key:   "rank",
				Value: 1,
			},
		}
		if _, err := c.catalog.UpdateOne(tx, bson.M{"_id": dto.RankUPItemID}, incQuery); err != nil {
			tx.AbortTransaction(ctx)
			return err
		}

		// Decrement rank by id
		decrQuery := bson.E{
			Key: "$inc",
			Value: bson.E{
				Key:   "rank",
				Value: -1,
			},
		}
		if _, err := c.catalog.UpdateOne(tx, bson.M{"_id": dto.RankDownItemID}, decrQuery); err != nil {
			tx.AbortTransaction(ctx)
			return err
		}

		return tx.CommitTransaction(ctx)
	}); err != nil {
		return err
	}

	return nil
}
