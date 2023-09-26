package repository

import (
	"encoding/json"
	"fmt"

	"github.com/Dyleme/timecache"
)

type UniversalCache struct {
	cache *timecache.Cache[string, []byte]
}

func NewUniversalCache() *UniversalCache {
	return &UniversalCache{
		cache: timecache.New[string, []byte](),
	}
}

func (uc *UniversalCache) Get(key string, obj any) error {
	op := "UniversalCache.Get: %w"
	objBytes, err := uc.cache.Get(key)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	if err = json.Unmarshal(objBytes, obj); err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (uc *UniversalCache) Delete(key string) error {
	uc.cache.Delete(key)

	return nil
}

func (uc *UniversalCache) Add(key string, obj any) error {
	op := "UniversalCache.Add: %wAdd"
	objBytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	uc.cache.StoreDefDur(key, objBytes)

	return nil
}
