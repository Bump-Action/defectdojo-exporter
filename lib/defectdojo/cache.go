package defectdojo

import "sync"

var (
	productTypeCacheMu sync.RWMutex
	productTypeCache   = make(map[int]string)
)

func getCachedProductTypeName(productTypeID int) (string, bool) {
	productTypeCacheMu.RLock()
	name, ok := productTypeCache[productTypeID]
	productTypeCacheMu.RUnlock()
	return name, ok
}

func setCacheProductTypeName(productTypeID int, name string) {
	productTypeCacheMu.Lock()
	productTypeCache[productTypeID] = name
	productTypeCacheMu.Unlock()
}
