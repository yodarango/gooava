package main

import (
	apiv1 "github.com/yodarango/gooava/api/v1"
	"github.com/yodarango/gooava/internal/constants"
	"github.com/yodarango/gooava/internal/models"
	"github.com/yodarango/gooava/internal/utils"
)

func Init() {

	// initialize the appConfig
	appConfig.Environment = constants.ENV_DEV
	appConfig.Session = map[string]interface{}{
		"test": "test",
	}
	templates, err := utils.CacheTemplates()
	if err != nil {
		panic("I could not cache templates")
	}
	appConfig.TemplateCache = templates

	// initliazie the apiConfig
	apiConfig := apiv1.NewApiConfig(&appConfig, "DB")
	apiv1.SetApiConfig(apiConfig)

	// initialize the modelConfig
	modelsConfig := models.NewModelConfig(&appConfig, "DB")
	models.SetModelConfig(modelsConfig)
}
