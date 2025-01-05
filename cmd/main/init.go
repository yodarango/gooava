package main

import (
	"fmt"
	"log"

	apiv1 "github.com/yodarango/gooava/api/v1"
	"github.com/yodarango/gooava/db"
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
		panic(fmt.Errorf("i could not cache templates %w", err))

	}
	appConfig.TemplateCache = templates

	// Initialize the DB config
	dbConn, err := db.Connect()
	dbConfig.DB = dbConn

	if err != nil {
		// if the db connection failed, there is no point in continuing, crash te app.
		log.Panicf("failed to connect to DB %v \n", err)
	}

	// initialize the apiConfig
	apiConfig := apiv1.NewApiConfig(&appConfig, &dbConfig)
	apiv1.SetApiConfig(apiConfig)

	// initialize the modelConfig
	modelsConfig := models.NewModelConfig(&appConfig, "DB")
	models.SetModelConfig(modelsConfig)

	// initialize the utils config
	utilsConfig := utils.NewTemplateConfig(&appConfig, "DB")
	utils.SetTemplateConfig(utilsConfig)

}
