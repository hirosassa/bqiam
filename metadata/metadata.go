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

// Load reads cacheFile.
func (ms *Metas) Load(cacheFile string) error {
	if _, err := toml.DecodeFile(cacheFile, ms); err != nil {
		return fmt.Errorf("Failed to load medadata cache file. %v", err)
	}
	return nil
}

// Save stores the cache data to the file
func (ms *Metas) Save(cacheFile string) error {
	f, err := os.Create(cacheFile)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("Failed to save metadata to the file. err: %s", err)
	}
	return toml.NewEncoder(f).Encode(ms)
}
