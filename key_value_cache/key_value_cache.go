// Package key_value_cache provides an extremely simple string key to value cache.
// This is used to reduce the number of lookups from MySQL filename
// to the equivalent table or logical name given the conversion
// routines to do this use many regexps and this is quite expensive.
package key_value_cache

import (
	"errors"

	"github.com/sjmudd/ps-top/logger"
)

// KeyValueCache provides a mapping from filename to table.schema etc
type KeyValueCache struct {
	cache           map[string]string
	readRequests    int
	servedFromCache int
	writeRequests   int
}

// NewKeyValueCache creates a new KeyValueCache entry.
func NewKeyValueCache() KeyValueCache {
	logger.Println("NewKeyValueCache()")

	return KeyValueCache{}
}

// Get will return the value in the cache if found
func (kvc *KeyValueCache) Get(key string) (result string, err error) {
	//	logger.Println("KeyValueCache.Get(", key, ")")
	if kvc.cache == nil {
		//		logger.Println("KeyValueCache.Get() kvc.cache is empty so enabling it")
		kvc.cache = make(map[string]string)
		kvc.readRequests = 0
		kvc.servedFromCache = 0
		kvc.writeRequests = 0
	}

	kvc.readRequests++

	if result, ok := kvc.cache[key]; ok {
		kvc.servedFromCache++
		//		logger.Println("Found: readRequests/servced_from_cache:", kvc.readRequests, kvc.servedFromCache)
		return result, nil
	}
	//	logger.Println("Not found: readRequests/servedFromCache:", kvc.readRequests, kvc.servedFromCache)

	return "", errors.New("Not found")
}

// Put writes to cache and return the value saved.
func (kvc *KeyValueCache) Put(key, value string) string {
	//	logger.Println("KeyValueCache.Put(", key, ",", value, ")")
	kvc.writeRequests++
	kvc.cache[key] = value

	return value
}

// Statistics returns some staticts on read and write requests and the number
// of requests served from cache.
func (kvc *KeyValueCache) Statistics() (int, int, int) {
	return kvc.readRequests, kvc.servedFromCache, kvc.writeRequests
}
