//go:build !solution

package lrucache

import "container/list"

type LRUCache struct {
	data        map[int]*list.Element
	accessOrder *list.List
	cap         int
}

type entity struct {
	Key   int
	Value int
}

var _ Cache = (*LRUCache)(nil)

func New(cap int) Cache {
	return &LRUCache{
		data:        make(map[int]*list.Element, cap),
		accessOrder: list.New(),
		cap:         cap,
	}
}

func (c *LRUCache) Get(key int) (int, bool) {
	e, exists := c.data[key]
	if !exists {
		return 0, false
	}
	c.accessOrder.MoveToBack(e)
	return e.Value.(*entity).Value, true
}

func (c *LRUCache) Set(key, value int) {
	if c.cap <= 0 {
		return
	}

	e, exists := c.data[key]
	if exists {
		c.accessOrder.MoveToBack(e)
		e.Value.(*entity).Value = value
		return
	}

	if c.accessOrder.Len() < c.cap {
		c.data[key] = c.accessOrder.PushBack(&entity{Key: key, Value: value})
		return
	}

	removed := c.accessOrder.Remove(c.accessOrder.Front())
	delete(c.data, removed.(*entity).Key)
	c.data[key] = c.accessOrder.PushBack(&entity{Key: key, Value: value})
}

func (c *LRUCache) Range(f func(key, value int) bool) {
	for e := c.accessOrder.Front(); e != nil; e = e.Next() {
		en := e.Value.(*entity)
		if !f(en.Key, en.Value) {
			return
		}
	}
}

func (c *LRUCache) Clear() {
	c.data = make(map[int]*list.Element, c.cap)
	c.accessOrder = list.New()
}
