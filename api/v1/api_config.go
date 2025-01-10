package apiv1

import (
	"github.com/yodarango/gooava/internal/repo"
)

// I export a singleton to the api configuration
var ApiConfig *ApiConfiguration

// I provide configuration struct to hold the data necessary across all the api endpoints
type ApiConfiguration struct {
	AppRepo *repo.AppRepo
}

// I receive the initial values to create a brand new instance of the ApiConfiguration struct
func NewApiConfig(appRepo *repo.AppRepo) *ApiConfiguration {
	return &ApiConfiguration{
		AppRepo: appRepo,
	}
}

// I receive the configuration created by the NewApiConfig() above and set it to the variable
// holding the singleton
func SetApiConfig(apiConfig *ApiConfiguration) {
	ApiConfig = apiConfig
}
