package repositories

import (
	"context"
	"errors"
	"io"

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

func (c *catalogRepo) GetIDByRank(ctx context.Context, rank uint) (primitive.ObjectID, error) {
	res := c.catalog.FindOne(ctx, bson.M{"rank": rank})
	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return primitive.ObjectID{}, domain.ErrItemNotFound
		}
		return primitive.ObjectID{}, res.Err()
	}
	var item domain.CatalogItem
	if err := res.Decode(&item); err != nil {
		return primitive.ObjectID{}, err
	}
	return item.ItemID, nil

}

func (c *catalogRepo) GetRankByID(ctx context.Context, itemID primitive.ObjectID) (uint, error) {
	res := c.catalog.FindOne(ctx, bson.M{"_id": itemID})
	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return 0, domain.ErrItemNotFound
		}
		return 0, res.Err()
	}
	var item domain.CatalogItem
	if err := res.Decode(&item); err != nil {
		return 0, err
	}
	return item.Rank, nil
}

func (c *catalogRepo) GetLastRank(ctx context.Context) (uint, error) {
	filter := bson.D{}
	opts := options.Find()
	opts.SetSort(bson.M{"rank": -1})
	opts.SetLimit(1)
	res, err := c.catalog.Find(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, domain.ErrItemNotFound
		}
		if errors.Is(err, io.EOF) {
			return 0, nil
		}
		return 0, err
	}
	var item domain.CatalogItem
	if err := res.Decode(&item); err != nil {
		if errors.Is(err, io.EOF) {
			return 0, nil
		}
		return 0, err
	}
	return item.Rank, nil
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
	if err != nil {
		return err
	}
	// TODO: move to service
	catalog, err := c.GetCatalog(ctx)
	if err != nil {
		return err
	}

	// Update ranks
	// TODO: move to service(bl)
	newCatalog := domain.UpdateRanks(catalog)
	for _, newItem := range newCatalog {
		if _, err := c.catalog.UpdateOne(ctx, bson.M{"_id": newItem.ItemID}, bson.M{"$set": bson.M{"rank": newItem.Rank}}); err != nil {
			return err
		}
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

func (c *catalogRepo) UpdateRanks(ctx context.Context, dto dto.UpdateItemDTO) error {
	client := c.catalog.Database().Client()

	if err := client.UseSession(ctx, func(tx mongo.SessionContext) error {
		if err := tx.StartTransaction(); err != nil {
			return err
		}

		// Increment rank by id
		incQuery := bson.M{"$inc": bson.M{"rank": 1}}
		if _, err := c.catalog.UpdateOne(tx, bson.M{"_id": dto.RankUPItemID}, incQuery); err != nil {
			tx.AbortTransaction(ctx)
			return err
		}

		// Decrement rank by id
		decrQuery := bson.M{"$inc": bson.M{"rank": -1}}
		if _, err := c.catalog.UpdateOne(tx, bson.M{"_id": dto.RankDownItemID}, decrQuery); err != nil {
			tx.AbortTransaction(ctx)
			return err
		}

		return tx.CommitTransaction(ctx)
	}); err != nil {
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
