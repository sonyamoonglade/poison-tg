package database

import (
	"context"
	"sync"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InMemoryRepo[V any] struct {
	mu   *sync.RWMutex
	data map[primitive.ObjectID]V
}

func NewInMemoryRepo[V any]() *InMemoryRepo[V] {
	return &InMemoryRepo[V]{
		mu:   new(sync.RWMutex),
		data: make(map[primitive.ObjectID]V),
	}
}

func (i *InMemoryRepo[V]) Save(ctx context.Context, c V) error {
	i.mu.Lock()
	i.data[primitive.NewObjectID()] = c
	i.mu.Unlock()
	return nil
}

func (i *InMemoryRepo[V]) GetByTelegramID(ctx context.Context, telegramID int64) (domain.Customer, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	for _, v := range i.data {
		if v, ok := (any(v).(domain.Customer)); ok {
			if v.TelegramID == telegramID {
				return v, nil
			}
		}
	}
	return domain.Customer{}, domain.ErrCustomerNotFound
}

func (i *InMemoryRepo[V]) UpdateState(ctx context.Context, telegramID int64, newState domain.State) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	for k, v := range i.data {
		if v, ok := (any(v).(domain.Customer)); ok {
			if v.TelegramID == telegramID {
				v.UpdateState(newState)
				i.data[k] = any(v).(V)
				return nil
			}
		}
	}
	return domain.ErrCustomerNotFound
}
