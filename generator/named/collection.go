// Package named holds things a user saves, renames and deletes by name.
package named

import (
	"fmt"
	"maps"
	"slices"
	"sync"
)

type Collection[T any] struct {
	noun  string
	mu    sync.RWMutex
	items map[string]T
}

// noun is what errors call an item, such as "profile".
func New[T any](noun string) *Collection[T] {
	return &Collection[T]{noun: noun, items: make(map[string]T)}
}

// Set replaces any item already saved under name.
func (c *Collection[T]) Set(name string, item T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[name] = item
}

func (c *Collection[T]) Get(name string) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[name]
	if !ok {
		return item, fmt.Errorf("no %s exists with name %q", c.noun, name)
	}
	return item, nil
}

func (c *Collection[T]) Rename(name, newName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[name]; !ok {
		return fmt.Errorf("%s %q does not exist", c.noun, name)
	}
	if _, ok := c.items[newName]; ok {
		return fmt.Errorf("%s %q already exists", c.noun, newName)
	}

	c.items[newName] = c.items[name]
	delete(c.items, name)
	return nil
}

func (c *Collection[T]) Delete(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[name]; !ok {
		return fmt.Errorf("graph contains no %s named %q", c.noun, name)
	}
	delete(c.items, name)
	return nil
}

// Names are sorted.
func (c *Collection[T]) Names() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return slices.Sorted(maps.Keys(c.items))
}

func (c *Collection[T]) All() map[string]T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return maps.Clone(c.items)
}
