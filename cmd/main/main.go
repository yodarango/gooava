package main

import (
	"fmt"
	"net/http"

	"github.com/yodarango/gooava/api"
	apiv1 "github.com/yodarango/gooava/api/v1"
	"github.com/yodarango/gooava/config"
	"github.com/yodarango/gooava/internal/models"
)

var PORT string = ":8003"

// app singleton
var appConfig config.AppConfig

func main() {

	handler := api.Routes()

	server := &http.Server{
		Addr:    PORT,
		Handler: handler,
	}
	// start the server
	fmt.Println("Starting server on port, ", PORT)
	server.ListenAndServe()

}

func init() {

	// initialize the appConfig
	appConfig.Environment = "DEV"
	appConfig.Session = map[string]interface{}{
		"test": "test",
	}
	appConfig.TemplateCache = nil

	// initliazie the apiConfig
	apiConfig := apiv1.NewApiConfig(&appConfig, "DB")
	apiv1.SetApiConfig(apiConfig)

	// initialize the modelConfig
	modelsConfig := models.NewModelConfig(&appConfig, "DB")
	models.SetModelConfig(modelsConfig)
}
