package utils

import (
	"fmt"
	"html/template"
	"net/http"
)

type TemplateRenderer struct {
	Title string      // title of the page
	Data  interface{} // the data to render. Typically sourced by the DB
	Files []string    // the template files to parse
	Name  string      // name of the template
}

func (t *TemplateRenderer) Render(w http.ResponseWriter) error {
	newTmpl := template.New(t.Name)

	templ, err := newTmpl.ParseFiles(append([]string{"web/templates/partials/base.html", "web/templates/partials/header.html"}, t.Files...)...)

	if err != nil {
		return fmt.Errorf("error parsing %s temple: %w", t.Name, err)
	}

	err = templ.Execute(w, t.Data)
	if err != nil {
		return fmt.Errorf("error executing %s temple: %w", t.Name, err)
	}

	return nil
}
