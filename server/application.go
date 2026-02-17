package server

type Application struct {
	config *utils.Config
}

func NewApplication(filePath) (*Application, error) {
	// Load the configuration file
	config := &utils.Config{}
	if err := utils.LoadConfig(filePath); err != nil {
		return nil, err
	}

	tokenBucket := algorithms.NewTokenBucket(config.MaxTokens)

	

	return &Application{
		config: config,
	}, nil
}