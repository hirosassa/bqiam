package metadata

import (
	"fmt"
	"os"

	bq "cloud.google.com/go/bigquery"
	"github.com/BurntSushi/toml"
)

var ms Metas

type Metas struct {
	Metas []Meta `toml:"Metas"`
}

type Meta struct {
	Project string        `toml:"Project"`
	Dataset string        `toml:"Dataset"`
	Role    bq.AccessRole `toml:"Role"`
	Entity  string        `toml:"Entity"`
}

// Load reads cacbeFile and loads cache data.
func (ms *Metas) Load(cacheFile string) error {
	cache := cacheFile
	if _, err := toml.DecodeFile(cache, ms); err != nil {
		return fmt.Errorf("Failed to load meda data cache file. %v", err)
	}
	return nil
}

// Save stores the cache data to cacheFile
func (ms *Metas) Save(cacheFile string) error {
	cache := cacheFile
	f, err := os.Create(cache)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("Failed to save meta data cache file. err: %s", err)
	}
	return toml.NewEncoder(f).Encode(ms)
}
