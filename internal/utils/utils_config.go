package utils

import (
	"github.com/yodarango/gooava/internal/repo"
)

// TODO! Secondo AI questo e una pratica cativa, ricercare (magari no)
// PRIVATE: Should not use singleton outside its scope
var utilsConfig *UtilsConfig

// singleton to the template configuration
type UtilsConfig struct {
	AppRepo *repo.AppRepo
}

// I create a brand new instance of the TemplateConfiguration
func NewUtilsConfig(ar *repo.AppRepo) *UtilsConfig {
	return &UtilsConfig{
		AppRepo: ar,
	}
}

// I set the TemplateConfig variable holding the TemplateConfiguration singleton
// to the TemplateConfiguration instance returned by NewTemplateConfig()
func SetUtilsConfig(uc *UtilsConfig) {
	utilsConfig = uc
}
