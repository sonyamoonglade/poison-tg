package dto

import (
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateCustomerDTO struct {
	LastPosition   *domain.Position
	Username       *string
	FullName       *string
	Meta           *domain.Meta
	CalculatorMeta *domain.CalculatorMeta
	PhoneNumber    *string
	Cart           *domain.Cart
	State          *domain.State
	CatalogOffset  *uint
}

type UpdateItemDTO struct {
	RankUPItemID   primitive.ObjectID
	RankDownItemID primitive.ObjectID
}

type AddCommentDTO struct {
	OrderID primitive.ObjectID
	Comment string
}

type ChangeOrderStatusDTO struct {
	OrderID   primitive.ObjectID
	NewStatus domain.Status
}
