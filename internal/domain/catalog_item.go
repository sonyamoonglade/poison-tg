package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type CatalogItem struct {
	ItemID         primitive.ObjectID
	ImageURLs      []string
	Title          string
	AvailableSizes []string
}
