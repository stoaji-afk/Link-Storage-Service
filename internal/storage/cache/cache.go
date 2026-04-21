package cache

import (
	"service/internal/models"
	"sync"
	"time"
)

type Cache struct {
    items map[string]*cacheItem
    mu    sync.RWMutex
    size  int
}

type cacheItem struct {
    link *models.Link
    ttl  time.Time
}

func New(size int) *Cache {
    return &Cache{
        items: make(map[string]*cacheItem),
        size:  size,
    }
}

func (c *Cache) Get(shortCode string) (*models.Link, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    item, exists := c.items[shortCode]
    if !exists || time.Now().After(item.ttl) {
        delete(c.items, shortCode)
        return nil, false
    }
    return item.link, true
}

func (c *Cache) Set(shortCode string, link *models.Link, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Ограничение размера кеша
    if len(c.items) >= c.size {
        // Простая очистка — удаляем первый элемент
        for key := range c.items {
            delete(c.items, key)
            break
        }
    }

    c.items[shortCode] = &cacheItem{
        link: link,
        ttl:  time.Now().Add(ttl),
    }
}

func (c *Cache) Delete(shortCode string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.items, shortCode)
}
