package infra

import (
	"os"
	"path/filepath"
)

func AddEnvironmentFile() error {
	path := filepath.Join(".upify", ".env")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
		return nil
	}

	return nil
}
