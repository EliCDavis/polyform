package generator

import "sync"

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
