package repositories

import (
	"github.com/sonyamoonglade/poison-tg/pkg/database"
)

type Repositories struct {
	Customer *customerRepo
	Order    *orderRepo
	Catalog  *catalogRepo
}

const (
	customers = "customers"
	orders    = "orders"
	catalog   = "catalog"
)

func NewRepositories(db *database.Mongo, catalogOnChangeFunc OnChangeFunc) Repositories {
	return Repositories{
		Customer: NewCustomerRepo(db.Collection(customers)),
		Order:    NewOrderRepo(db.Collection(orders)),
		Catalog:  NewCatalogRepo(db.Collection(catalog), catalogOnChangeFunc),
	}
}
