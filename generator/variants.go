package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/variable"
)

func RunVariants(instance *graph.Instance, profiles []variable.Profile, root string) error {
	for i, profile := range profiles {
		if err := instance.ApplyProfile(profile); err != nil {
			return fmt.Errorf("applying variant %d: %w", i, err)
		}

		folder := filepath.Join(root, fmt.Sprintf("variant-%04d", i))
		if err := os.MkdirAll(folder, os.ModePerm); err != nil {
			return err
		}

		if err := graph.WriteToFolder(instance, folder); err != nil {
			return fmt.Errorf("writing variant %d: %w", i, err)
		}

		if err := writeVariantProfile(folder, profile); err != nil {
			return fmt.Errorf("writing variant %d profile: %w", i, err)
		}
	}
	return nil
}

func writeVariantProfile(folder string, profile variable.Profile) error {
	data, err := json.MarshalIndent(profile, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(folder, "profile.json"), data, 0666)
}
