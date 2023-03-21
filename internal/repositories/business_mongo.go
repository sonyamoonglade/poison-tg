package repositories

import (
	"context"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type businessRepo struct {
	business *mongo.Collection
}

func NewBusinessRepo(business *mongo.Collection) *businessRepo {
	return &businessRepo{
		business: business,
	}
}

func (b businessRepo) GetRequisites(ctx context.Context) (domain.Requisites, error) {
	cur, err := b.business.Find(ctx, bson.M{})
	if err != nil {
		return domain.Requisites{}, nil
	}
	var reqs domain.Requisites
	if err := cur.Decode(&reqs); err != nil {
		return domain.Requisites{}, nil
	}

	return reqs, nil
}
