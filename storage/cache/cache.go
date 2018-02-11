package cache

import (
	"fmt"
	"github.com/boreq/flightradar-backend/logging"
	"github.com/boreq/flightradar-backend/storage"
	"sort"
	"time"
)

var log = logging.GetLogger("storage/cache")

// icaoCache will only hold data for that many planes.
const icaoCacheLimit = 10

// timeRangeCache will only hold data from that many granularity ranges.
const timeRangeCacheLimit = 10

type cachedData struct {
	StoredData []storage.StoredData
	LastAccess time.Time
}

// Creates a new storage that wraps the provided storage providing in-memory
// caching to avoid unnecessary IO. For performance reasons this cache is not
// perfect and may loose some data.
func New(wrappedStorage storage.Storage) storage.Storage {
	rv := &cache{
		icaoCache:      make(map[string]*cachedData),
		timeRangeCache: make(map[int64]*cachedData),
		wrappedStorage: wrappedStorage,
	}
	return rv
}

type cache struct {
	icaoCache      map[string]*cachedData
	timeRangeCache map[int64]*cachedData
	wrappedStorage storage.Storage
}

func (c *cache) Store(data storage.StoredData) error {
	c.cleanup()

	if err := c.wrappedStorage.Store(data); err != nil {
		return err
	}

	// Add to icaoCache if it exists for this plane
	cData, ok := c.icaoCache[*data.Data.Icao]
	if ok {
		cData.StoredData = append(cData.StoredData, data)
		log.Debugf("Appending to cache for ICAO %s", *data.Data.Icao)
	}

	// Add to timeRangeCache if it exists for this time
	key := timeToCacheKey(data.Time)
	cData, ok = c.timeRangeCache[key]
	if ok {
		cData.StoredData = append(cData.StoredData, data)
		log.Debugf("Appending to cache for time key %d", key)
	}

	return nil
}

func (c *cache) Retrieve(icao string) ([]storage.StoredData, error) {
	c.cleanup()

	// Try to read from cache
	cData, ok := c.icaoCache[icao]
	if ok {
		cData.LastAccess = time.Now()
		return cData.StoredData, nil
	}

	// Build cache
	data, err := c.wrappedStorage.Retrieve(icao)
	if err != nil {
		return nil, err
	}
	log.Debugf("Building cache for %s (%d elements)", icao, len(data))
	c.icaoCache[icao] = &cachedData{LastAccess: time.Now(), StoredData: data}
	return data, nil
}

func (c *cache) RetrieveAll() ([]storage.StoredData, error) {
	return c.wrappedStorage.RetrieveAll()
}

func (c *cache) RetrieveTimerange(from time.Time, to time.Time) ([]storage.StoredData, error) {
	c.cleanup()

	var rv []storage.StoredData

	min := timeToCacheKey(from)
	max := timeToCacheKey(to)
	for key := min; key <= max; key += keyGranularity {
		cData, ok := c.timeRangeCache[key]
		// Build cache for this key
		if !ok {
			log.Debugf("Building cache for %d", key)
			f := time.Unix(key*keyGranularity, 0)
			t := time.Unix(key*keyGranularity+keyGranularity, 0)
			fmt.Printf("%s\n", f)
			fmt.Printf("%s\n", t)
			data, err := c.wrappedStorage.RetrieveTimerange(f, t)
			if err != nil {
				return nil, err
			}
			cData = &cachedData{LastAccess: time.Now(), StoredData: data}
			c.timeRangeCache[key] = cData
			log.Debugf("Building cache for %d (%d elements)", key, len(data))
		}

		cData.LastAccess = time.Now()
		for _, d := range cData.StoredData {
			if d.Time.After(from) && d.Time.Before(to) {
				rv = append(rv, d)
			}
		}
	}

	return rv, nil
}

type tmpIntStruct struct {
	Key  int64
	Time time.Time
}

type tmpStringStruct struct {
	Key  string
	Time time.Time
}

func (c *cache) cleanup() {
	c.cleanupTimeRangeCache()
	c.cleanupIcaoCache()
}

func (c *cache) cleanupTimeRangeCache() {
	var keys []tmpIntStruct

	for key, value := range c.timeRangeCache {
		keys = append(keys, tmpIntStruct{key, value.LastAccess})
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i].Time.Before(keys[j].Time) })

	for i := 10; i < len(keys); i++ {
		delete(c.timeRangeCache, keys[i].Key)
		log.Debugf("Deleted time range cache key %d", keys[i].Key)
	}
}

func (c *cache) cleanupIcaoCache() {
	var keys []tmpStringStruct

	for key, value := range c.icaoCache {
		keys = append(keys, tmpStringStruct{key, value.LastAccess})
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i].Time.Before(keys[j].Time) })

	for i := 10; i < len(keys); i++ {
		delete(c.icaoCache, keys[i].Key)
		log.Debugf("Deleted ICAO cache key %s", keys[i].Key)
	}
}

const keyGranularity int64 = 60 * 60 * 24

func timeToCacheKey(t time.Time) int64 {
	return t.Unix() / keyGranularity
}
