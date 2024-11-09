package utils

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/yodarango/gooava/internal/constants"
)

type TemplateRenderer struct {
	*TemplateConfiguration
	Title string      // title of the page
	Data  interface{} // the data to render. Typically sourced by the DB
	Files []string    // the template files to parse
	Name  string      // name of the template
}

// set the template functions
var functions = template.FuncMap{}

const pathToTemplates = "../../web/templates"

// I execute the template TemplateRenderer.Name from the cache if in production
// or from the pared
func (t *TemplateRenderer) Render(w http.ResponseWriter) error {

	// If in dev mod, the cache must not be used to allow for hot reload
	if t.AppConfig.Environment == constants.ENV_DEV {
		templateCache, err := CacheTemplates()
		if err != nil {
			return fmt.Errorf("there was an error getting the ")
		}

		t.AppConfig.TemplateCache = templateCache
	}

	// check that the template requested exists in the cache
	templ, ok := t.AppConfig.TemplateCache[t.Name]

	if !ok {
		return fmt.Errorf("could not find %s in the cache ", t.Name)
	}

	err := templ.Execute(w, t.Data)

	if err != nil {
		return fmt.Errorf("error executing %s temple: %w", t.Name, err)
	}

	return nil
}

// I iterate for all the files within pathToTemplates and parse them and return then in
// a map holding AppConfig.Templates
func CacheTemplates() (map[string]*template.Template, error) {

	var templates []string

	templateCache := map[string]*template.Template{}

	// walk the template directory and extract all the html files
	err := filepath.Walk(pathToTemplates, func(path string, info fs.FileInfo, err error) error {

		if filepath.Ext(path) == ".html" {
			templates = append(templates, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the templates path: %w", err)
	}

	// iterate for each template file and set them in map holding the templates
	for _, templ := range templates {
		templateName := filepath.Base(templ)
		parsedTemplate, err := template.New(templateName).Funcs(functions).Parse(templ)

		if err != nil {
			return nil, fmt.Errorf("error parsing %s template: %w", templateName, err)
		}

		templateCache[templateName] = parsedTemplate
	}

	return templateCache, nil
}
