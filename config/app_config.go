/**
* Global config for the app
**/

package config

import "html/template"

type AppConfig struct {
	TemplateCache map[string]*template.Template
	Session       map[string]interface{}
	Environment   string
}
