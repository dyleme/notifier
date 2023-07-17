package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func Load() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	for {
		err = godotenv.Load(dir + ".env")
		if err == nil {
			return nil
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return fmt.Errorf("file not found")
		}
	}
}
