package main

import (
	"fmt"
	"log"

	"github.com/yodarango/gooava/db"
	"github.com/yodarango/gooava/internal/constants"
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

	// There are two global singletons in, App and Db. If all went well,
	//pass them to all pars of the app

	appRepo.NewDbRepoSetup(&appConfig, &dbConfig)

	// TODO! I don't any of these. All I need is just to have the AppRpo that can be
	// TODO!  injected into ANY part of the app!

	// initialize the apiConfig
	// apiConfig := apiv1.NewApiConfig(&appConfig, &dbConfig)
	// apiv1.SetApiConfig(apiConfig)

	// // initialize the modelConfig
	// modelsConfig := models.NewModelConfig(&appConfig, "DB")
	// models.SetModelConfig(modelsConfig)

	// // initialize the utils config
	// utilsConfig := utils.NewTemplateConfig(&appConfig, "DB")
	// utils.SetTemplateConfig(utilsConfig)

}
