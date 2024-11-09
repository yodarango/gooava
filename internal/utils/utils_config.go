package utils

import "github.com/yodarango/gooava/config"

var TemplateConfig *TemplateConfiguration

// singleton to the template configuration
type TemplateConfiguration struct {
	AppConfig *config.AppConfig
	DB        string
}

// I create a brand new instance of the TemplateConfiguration
func NewTemplateConfig(appConfig *config.AppConfig, DB string) *TemplateConfiguration {
	return &TemplateConfiguration{
		AppConfig: appConfig,
		DB:        DB,
	}
}

// I set the TemplateConfig variable holding the TemplateConfiguration singleton
// to the TemplateConfiguration instance returned by NewTemplateConfig()
func SetTemplateConfig(templateConfig *TemplateConfiguration) {
	TemplateConfig = templateConfig
}
