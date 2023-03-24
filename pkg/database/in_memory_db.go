package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InMemoryRepo[V any] struct {
	mu   *sync.RWMutex
	data map[uint32]V
}

func NewInMemoryRepo[V any]() *InMemoryRepo[V] {
	return &InMemoryRepo[V]{
		mu:   new(sync.RWMutex),
		data: make(map[uint32]V),
	}
}

func (i *InMemoryRepo[V]) Save(ctx context.Context, v V) error {
	defer i.PrintDb()
	if c, ok := any(v).(domain.Customer); ok {
		exists, key := i.findByID(c.CustomerID)
		fmt.Println(exists, key)
		if exists {
			i.mu.Lock()
			defer i.mu.Unlock()
			c.LastEditPosition.PositionID = primitive.NewObjectID()
			i.data[key] = any(c).(V)
			return nil
		} else {
			i.mu.Lock()
			defer i.mu.Unlock()
			c.CustomerID = primitive.NewObjectID()
			i.data[i.nextId()] = any(c).(V)
			return nil
		}
	}
	return nil
}

func (i *InMemoryRepo[V]) findByID(id primitive.ObjectID) (bool, uint32) {
	fmt.Printf("printng db\n")
	i.mu.Lock()
	defer i.mu.Unlock()
	for k, v := range i.data {
		if c, ok := any(v).(domain.Customer); ok {
			if c.CustomerID == id {
				return true, k
			}
		}
	}
	return false, 0
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
func (i *InMemoryRepo[V]) nextId() uint32 {
	return uint32(len(i.data)+2) * 2
}

func (i *InMemoryRepo[V]) PrintDb() {
	i.mu.Lock()
	defer i.mu.Unlock()
	for k, v := range i.data {
		fmt.Printf("key:% d value: %v\n", k, v)
	}
}
