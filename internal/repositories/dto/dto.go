package dto

import "github.com/sonyamoonglade/poison-tg/internal/domain"

type UpdateCustomerDTO struct {
	LastPosition *domain.Position
	Username     *string
	PhoneNumber  *string
	Cart         *domain.Cart
	State        *domain.State
}
