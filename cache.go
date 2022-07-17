package main

type LookupCache[K comparable, V any] struct {
	cache map[K]*V
}

func NewLookupCache[K comparable, V any](values []V, selector func(r *V) K) *LookupCache[K, V] {
	cache := map[K]*V{}
	for i := range values {
		value := &values[i]
		cache[selector(value)] = value
	}
	return &LookupCache[K, V]{
		cache: cache,
	}
}

func (l *LookupCache[K, V]) Get(key K) (*V, bool) {
	result, ok := l.cache[key]
	return result, ok
}

type LookupGroupCache[K comparable, V any] struct {
	cache map[K][]*V
}

func NewLookupGroupCache[K comparable, V any](values []V, selector func(r *V) K) *LookupGroupCache[K, V] {
	cache := map[K][]*V{}
	for i := range values {
		value := &values[i]
		key := selector(value)
		_, ok := cache[key]
		if !ok {
			cache[key] = []*V{}
		}
		cache[key] = append(cache[key], value)
	}
	return &LookupGroupCache[K, V]{
		cache: cache,
	}
}

func (l *LookupGroupCache[K, V]) Get(key K) []*V {
	result, ok := l.cache[key]
	if ok {
		return result
	}
	return []*V{}
}
