package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type cacheEntry struct {
	key   Key
	value interface{}
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem),
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	if item, ok := c.items[key]; ok {
		entry := item.Value.(*cacheEntry)
		entry.value = value
		c.queue.MoveToFront(item)
		return true
	}

	entry := &cacheEntry{key: key, value: value}
	item := c.queue.PushFront(entry)
	c.items[key] = item

	if c.queue.Len() > c.capacity {
		oldest := c.queue.Back()
		oldEntry := oldest.Value.(*cacheEntry)
		delete(c.items, oldEntry.key)
		c.queue.Remove(oldest)
	}

	return false
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	c.queue.MoveToFront(item)
	return item.Value.(*cacheEntry).value, true
}

func (c *lruCache) Clear() {
	c.queue = NewList()
	c.items = make(map[Key]*ListItem)
}
