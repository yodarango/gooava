package apiv1

import (
	"github.com/yodarango/gooava/config"
	"github.com/yodarango/gooava/db"
)

// I export a singleton to the api configuration
var ApiConfig *ApiConfiguration

// I provide configuration struct to hold the data necessary across all the api endpoints
type ApiConfiguration struct {
	AppConfig *config.AppConfig
	DB        *db.DbConfig
}

// I receive the initial values to create a brand new instance of the ApiConfiguration struct
func NewApiConfig(appConfig *config.AppConfig, db *db.DbConfig) *ApiConfiguration {
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
