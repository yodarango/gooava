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

type Option func(*AppConfig)

func WithTemplateCache(tc map[string]*template.Template) Option {

	return func(c *AppConfig) {
		c.TemplateCache = tc
	}
}

func WithSession(s map[string]interface{}) Option {

	return func(c *AppConfig) {
		c.Session = s
	}
}

func WithEnvironment(env string) Option {

	return func(c *AppConfig) {
		c.Environment = env
	}
}

func NewAppConfig(opts ...Option) *AppConfig {

	// declare any default values here if any
	config := &AppConfig{}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
