package utils

import (
	"os"

	"github.com/bytedance/sonic"
)

type Config struct {
	MaxTokens  float64 `json:"maxTokens"`
	RefillRate float64 `json:"refillRate"`
}

func (config *Config) LoadConfig(filePath string) error {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := sonic.Unmarshal(fileBytes, config); err != nil {
		return err
	}

	return nil
}
