package dto

import "github.com/yodarango/gooava/config"

// I export a singleton to the api configuration
var DtoConfig *DtoConfiguration

// I provide configuration struct to hold the data necessary across all the api endpoints
type DtoConfiguration struct {
	AppConfig *config.AppConfig
	DB        string
}

// I receive the initial values to create a brand new instance of the DtoConfiguration struct
func NewDtoConfig(appConfig *config.AppConfig, db string) *DtoConfiguration {
	return &DtoConfiguration{
		AppConfig: appConfig,
		DB:        db,
	}
}

// I receive the configuration created by the NewApiConfig() above and set it to the variable
// holding the singleton
func SetApiConfig(dtoConfig *DtoConfiguration) {
	DtoConfig = dtoConfig
}
