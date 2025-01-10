package repo

import (
	"github.com/yodarango/gooava/config"
	"github.com/yodarango/gooava/internal/db"
)

type AppRepo struct {
	App *config.AppConfig
	DB  *db.DbConfig
}

func NewAppRepo(app *config.AppConfig, db *db.DbConfig) *AppRepo {
	return &AppRepo{
		App: app,
		DB:  db,
	}
}
