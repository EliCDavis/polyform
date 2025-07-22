package run

import "time"

// TODO: JUST IMPORT A LIBRARY DUDE

type lruCacheEntry[T any] struct {
	data T
	time time.Time
}

type lruCache[T any] struct {
	data map[string]*lruCacheEntry[T]
	max  int
}

func (lru lruCache[T]) Get(id string) (T, bool) {
	v, ok := lru.data[id]
	if !ok {
		var t T
		return t, ok
	}
	v.time = time.Now()
	return v.data, ok
}

func (lru lruCache[T]) Add(item T) string {

	if len(lru.data) > lru.max {
		oldestId := ""
		oldestTime := time.Now()
		for id, v := range lru.data {
			if v.time.Before(oldestTime) {
				oldestId = id
				oldestTime = v.time
			}
		}
		delete(lru.data, oldestId)
	}

	id := generateRandomString(10)
	lru.data[id] = &lruCacheEntry[T]{
		data: item,
		time: time.Now(),
	}
	return id
}
