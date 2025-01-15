package generator

import (
	"fmt"
	"strings"
	"sync"
)

func NewSyncMap[Key comparable, Value any]() *SyncMap[Key, Value] {
	return &SyncMap[Key, Value]{
		data: make(map[Key]Value),
	}
}

type SyncMap[Key comparable, Value any] struct {
	data  map[Key]Value
	mutex sync.RWMutex
}

func (sm *SyncMap[Key, Value]) Get(key Key) Value {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.data[key]
}

func (sm *SyncMap[Key, Value]) Set(key Key, value Value) {
	sm.mutex.Lock()
	sm.data[key] = value
	sm.mutex.Unlock()
}

// ============================================================================

type NestedSyncMap struct {
	data  map[string]any
	mutex sync.RWMutex
}

func NewNestedSyncMap() *NestedSyncMap {
	return &NestedSyncMap{
		data: make(map[string]any),
	}
}

func (sm *NestedSyncMap) OverwriteData(data map[string]any) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if data == nil {
		sm.data = make(map[string]any)
	} else {
		sm.data = data
	}

}

func (sm *NestedSyncMap) lookup(key string) (map[string]any, string) {
	elements := strings.Split(key, ".")
	current := sm.data
	for i := 0; i < len(elements)-1; i++ {
		v, ok := current[elements[i]]
		if !ok {
			panic(fmt.Errorf("key %s doens't exist on key %s", elements[i], key))
		}

		casted, ok := v.(map[string]any)
		if !ok {
			panic(fmt.Errorf("%s isn't a map", key))
		}
		current = casted
	}
	return current, elements[len(elements)-1]
}

func (sm *NestedSyncMap) Get(key string) any {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	data, key := sm.lookup(key)
	return data[key]
}

func (sm *NestedSyncMap) Delete(key string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	data, key := sm.lookup(key)
	delete(data, key)
}

func (sm *NestedSyncMap) Set(key string, value any) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	elements := strings.Split(key, ".")
	current := sm.data
	for i := 0; i < len(elements)-1; i++ {
		v, ok := current[elements[i]]
		if ok {
			casted, ok := v.(map[string]any)
			if !ok {
				panic(fmt.Errorf("%s isn't a map", key))
			}
			current = casted
		} else {
			newMap := make(map[string]any)
			current[elements[i]] = newMap
			current = newMap
		}

	}
	current[elements[len(elements)-1]] = value
}
