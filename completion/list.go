package completion

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type List struct {
	Users    []string `toml:"Users"`
	Datasets []string `toml:"Datasets"`
	Projects []string `toml:"Projects"`

	DisplaySizeLimit int `toml:"DisplaySizeLimit"`
}

func (l *List) Load(file string) error {
	if _, err := toml.DecodeFile(file, l); err != nil {
		return fmt.Errorf("failed to load completion list file. %v", err)
	}
	return nil
}

func (l *List) Save(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to save completion list to the file. err: %s", err)
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(l)
}
