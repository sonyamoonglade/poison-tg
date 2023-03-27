package catalog

import (
	"sync"

	"github.com/sonyamoonglade/poison-tg/internal/domain"
)

type CatalogProvider struct {
	mu    *sync.RWMutex
	items []domain.CatalogItem
}

func NewCatalogProvider() *CatalogProvider {
	return &CatalogProvider{
		mu:    new(sync.RWMutex),
		items: nil,
	}
}

func (c *CatalogProvider) Load(items []domain.CatalogItem) {
	c.mu.Lock()
	c.items = make([]domain.CatalogItem, len(items), len(items))
	copy(c.items, items)
	c.mu.Unlock()
}

func (c *CatalogProvider) HasNext(offset uint) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return offset >= 0 && offset < uint(len(c.items)-1)
}

func (c *CatalogProvider) HasPrev(offset uint) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return offset > 0 && offset <= uint(len(c.items))
}

func (c *CatalogProvider) LoadNext(offset uint) domain.CatalogItem {
	if !c.HasNext(offset) {
		return domain.CatalogItem{}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.items[offset+1]
}

func (c *CatalogProvider) LoadPrev(offset uint) domain.CatalogItem {
	if !c.HasPrev(offset) {
		return domain.CatalogItem{}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.items[offset-1]
}

func (c *CatalogProvider) LoadFirst() domain.CatalogItem {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.items) > 0 {
		return c.items[0]
	}
	return domain.CatalogItem{}
}

func (c *CatalogProvider) LoadAt(offset uint) domain.CatalogItem {
	if len(c.items) == 0 {
		return domain.CatalogItem{}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	_ = c.items[offset]
	return c.items[offset]
}
