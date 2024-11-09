package models

import "github.com/yodarango/gooava/config"

// I export a singleton to the api configuration
var ModelConfig *ModelConfiguration

// I provide configuration struct to hold the data necessary across all the api endpoints
type ModelConfiguration struct {
	AppConfig *config.AppConfig
	DB        string
}

// I receive the initial values to create a brand new instance of the ModelConfig struct
func NewModelConfig(appConfig *config.AppConfig, db string) *ModelConfiguration {
	return &ModelConfiguration{
		AppConfig: appConfig,
		DB:        db,
	}
}

// I receive the configuration created by the NewModelConfig() above and set it to the variable
// holding the singleton
func SetModelConfig(modelConfig *ModelConfiguration) {
	ModelConfig = modelConfig
}
