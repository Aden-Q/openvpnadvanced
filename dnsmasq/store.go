package dnsmasq

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

const cacheFilePath = "assets/cache.json"

var storeLock sync.RWMutex

func LoadCacheFromFile() (map[string]DNSRecord, error) {
	storeLock.RLock()
	defer storeLock.RUnlock()

	data := make(map[string]DNSRecord)

	file, err := os.Open(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return data, nil
		}
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if len(bytes) == 0 {
		return data, nil
	}

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func SaveCacheToFile(cache *Cache) error {
	storeLock.Lock()
	defer storeLock.Unlock()

	data := cache.Raw()

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cacheFilePath, bytes, 0644)
	if err != nil {
		return err
	}

	fmt.Println("✅ Cache saved to cache.json")
	return nil
}
