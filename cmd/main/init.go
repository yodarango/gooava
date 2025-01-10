package main

import (
	"fmt"
	"log"

	apiv1 "github.com/yodarango/gooava/api/v1"
	"github.com/yodarango/gooava/config"
	"github.com/yodarango/gooava/internal/constants"
	"github.com/yodarango/gooava/internal/models"
	"github.com/yodarango/gooava/internal/repo"
	"github.com/yodarango/gooava/internal/utils"
	"github.com/yodarango/gooavainternal/db"
)

func Init() *repo.AppRepo {

	// cache the templates so I don't have to do parse them every time
	templates, err := utils.CacheTemplates()
	if err != nil {
		panic(fmt.Errorf("i could not cache templates %w", err))

	}

	// initialize the appConfig
	appConfig := config.NewAppConfig(
		config.WithEnvironment(constants.ENV_DEV),
		config.WithSession(map[string]interface{}{}),
		config.WithTemplateCache(templates),
	)

	// Initialize the DB config
	dbConn, err := db.Connect()
	if err != nil {
		// if the db connection failed, there is no point in continuing, crash te app.
		log.Panicf("failed to connect to DB %v \n", err)
	}

	dbConfig := db.NewDBConfig(dbConn)

	// There are two global singletons, the db singleton and the app singleton,
	// set them all into a global app singleton to pas around the app
	appRepo := repo.NewAppRepo(appConfig, dbConfig)

	// Pass singleton to the app
	apiConfig := apiv1.NewApiConfig(appRepo) // ricercare pointers
	apiv1.SetApiConfig(apiConfig)

	// initialize the modelConfig
	modelsConfig := models.NewModelConfig(appRepo)
	models.SetModelConfig(modelsConfig)

	// initialize the utils config
	utilsConfig := utils.NewUtilsConfig(appRepo)
	utils.SetUtilsConfig(utilsConfig)

	return appRepo

}
