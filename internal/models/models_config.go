package models

import (
	"gooava/repo"
)

var ModelsRepo *repo.AppRepo

func SetModelsConfig(ar *repo.AppRepo) {
	ModelsRepo = ar
}