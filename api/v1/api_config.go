package apiv1

import "github.com/yodarango/gooava/config"

// I export a singleton to the api configuration
var ApiConfig *ApiConfiguration

// I provide configuration struct to hold the data necessary across all the api endpoints
type ApiConfiguration struct {
	AppConfig *config.AppConfig
	DB        string
}

// I receive the initial values to create a brand new instance of the ApiConfiguration struct
func NewApiConfig(appConfig *config.AppConfig, db string) *ApiConfiguration {
	return &ApiConfiguration{
		AppConfig: appConfig,
		DB:        db,
	}
}

// I receive the configuration created by the NewApiConfig() above and set it to the variable
// holding the singleton
func SetApiConfig(apiConfig *ApiConfiguration) {
	ApiConfig = apiConfig
}
