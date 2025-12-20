package models

import (
	"keewee/repo"
)

var ModelsRepo *repo.AppRepo

func SetModelsConfig(ar *repo.AppRepo) {
	ModelsRepo = ar
}