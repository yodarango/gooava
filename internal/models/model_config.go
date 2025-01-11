package models

import (
	"github.com/yodarango/gooava/internal/repo"
)

var ModelConfig *ModelConfiguration

// I provide configuration struct to hold the data necessary across all the api endpoints
type ModelConfiguration struct {
	AppRepo *repo.AppRepo
}

// I receive the initial values to create a brand new instance of the ModelConfig struct
func NewModelConfig(ar *repo.AppRepo) *ModelConfiguration {
	return &ModelConfiguration{
		AppRepo: ar,
	}
}

// I receive the configuration created by the NewModelConfig() above and set it to the variable
// holding the singleton
func SetModelConfig(modelConfig *ModelConfiguration) {
	ModelConfig = modelConfig
}
