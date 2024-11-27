package config

import (
	"github.com/joho/godotenv"
)

// Load some comment
func Load(path string) error {
	err := godotenv.Load(path)
	if err != nil {
		return err
	}

	return nil
}

// GRPCConfig some comment
type GRPCConfig interface {
	Address() string
}

// PGConfig some comment
type PGConfig interface {
	DSN() string
}
