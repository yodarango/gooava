package repo

import (
	"github.com/yodarango/gooava/config"
	"github.com/yodarango/gooava/db"
)

// Global repo available to all the app
type AppRepo struct {
	App *config.AppConfig
	DB  *db.DbConfig
}

// Set up the values of the global repo
func (ar *AppRepo) NewDbRepoSetup(app *config.AppConfig, db *db.DbConfig) {
	ar.App = app
	ar.DB = db
}
