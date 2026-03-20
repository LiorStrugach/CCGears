package preset

import (
	"os"
	"sort"

	"github.com/mryan/ccgears/internal/config"
)

// List returns all presets sorted by name.
func List(cfg *config.Config) ([]*Meta, error) {
	entries, err := os.ReadDir(cfg.StorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var presets []*Meta
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		meta, _ := readMeta(cfg.StorePath + "/" + e.Name())
		if meta != nil {
			presets = append(presets, meta)
		}
	}

	sort.Slice(presets, func(i, j int) bool {
		return presets[i].Name < presets[j].Name
	})
	return presets, nil
}
