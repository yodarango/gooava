/**
* Global config for the app
**/

package config

import "text/template"

type AppConfig struct {
	TemplateCache map[string]*template.Template
	Session       map[string]interface{}
	Environment   string
}
