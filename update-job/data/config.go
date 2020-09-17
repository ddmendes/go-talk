package data

import (
	"encoding/json"
	"io"
)

// Configuration defines system configuration
type Configuration struct {
	PostgresDSN string
	GitHubToken string
	PageSize    int
}

// Config stores system configuration
var Config Configuration

// LoadConfiguration loads system configuration from a Reader
func LoadConfiguration(r io.Reader) error {
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&Config)
	if err != nil {
		return err
	}
	return nil
}
