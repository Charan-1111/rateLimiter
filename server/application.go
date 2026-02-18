package server

import (
	"goapp/algorithms"
	"goapp/utils"
)

type Application struct {
	config      *utils.Config
	tokenBucket *algorithms.TokenBucket
}

func NewApplication(filePath string) (*Application, error) {
	// Load the configuration file
	config := &utils.Config{}
	if err := config.LoadConfig(filePath); err != nil {
		return nil, err
	}

	tokenBucket := algorithms.NewTokenBucket(config.MaxTokens, config.RefillRate)

	return &Application{
		config:      config,
		tokenBucket: tokenBucket,
	}, nil
}
