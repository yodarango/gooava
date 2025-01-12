package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yodarango/gooava/internal/constants"
)

type TemplateRenderer struct {
	Title  string          // title of the page
	Data   interface{}     // the data to render. Typically sourced by the DB
	Name   string          // name of the template
	UiMeta *TemplateUIMEta // things that might change in the UI like header icons, menu options, etc
}

type TemplateUIMEta struct {
	MenuIcon string `json:"main_icon"`
}

// set the template functions
var functions = template.FuncMap{}

const PATH_TO_TEMPLATES = "/web/templates"
const PATH_TO_PARTIALS = "partials"

/*
***********************************************************************************
* NOT USED FOR NOW. I thought about sending templates as response instead of JSON but
* have changed my mind since

*I return the given template paths as an array of TemplateRenders.
*primarily used to send in body request.
***************************************
 */
func (t *TemplateRenderer) ParseTemplates(paths []string) (string, error) {

	var functions = template.FuncMap{}

	templates := make(map[string]string)

	for _, path := range paths {

		templ, err := template.New(path).Funcs(functions).ParseFiles(path)

		if err != nil {
			return "", fmt.Errorf("error creating template %s. Error: %w", path, err)
		}

		buf := new(bytes.Buffer)

		err = templ.Execute(buf, t.Data)

		if err != nil {
			return "", fmt.Errorf("error executing %s to buffer. Error:  %w", path, err)
		}

		templates[path] = buf.String()

	}

	jsonString, err := json.Marshal(templates)

	if err != nil {
		return "", fmt.Errorf("error marshalling %w", err)
	}

	return string(jsonString), nil
}

// Add a new function to the functions Maps
func (t *TemplateRenderer) AddFunction(function template.FuncMap) {
}

// I execute the template TemplateRenderer.Name from the cache if in production
// or from the pared
func (t *TemplateRenderer) Render(w http.ResponseWriter) error {

	//If in dev mod, the cache must not be used to allow for hot reload
	if utilsConfig.AppRepo.App.Environment == constants.ENV_DEV {
		templateCache, err := CacheTemplates()
		if err != nil {
			return fmt.Errorf("there was an error getting the cache %w", err)
		}

		utilsConfig.AppRepo.App.TemplateCache = templateCache
	}

	//check that the template requested exists in the cache
	templ, ok := utilsConfig.AppRepo.App.TemplateCache[t.Name]

	if !ok {
		return fmt.Errorf("could not find %s in the cache ", t.Name)
	}

	buf := new(bytes.Buffer)

	err := templ.Execute(buf, t)

	if err != nil {
		return fmt.Errorf("error executing %s temple: %w", t.Name, err)
	}
	_, err = buf.WriteTo(w)

	if err != nil {
		return fmt.Errorf("error writing %s temple buffer: %w", t.Name, err)
	}

	return nil
}

// I iterate for all the files within pathToTemplates and parse them and return then in
// a map holding AppConfig.Templates
func CacheTemplates() (map[string]*template.Template, error) {

	var templates []string
	var partials []string

	templateCache := map[string]*template.Template{}

	pathDir, _ := os.Getwd()
	pathDir += PATH_TO_TEMPLATES

	// Extract all templates first
	err := filepath.Walk(pathDir+"/"+PATH_TO_PARTIALS, func(path string, info fs.FileInfo, err error) error {
		// check if this file exists
		if err != nil {
			if err == os.ErrNotExist {
				log.Printf("error walking the templates path: %v \n", err)
				return err
			}

			log.Printf("error walking the templates path: %v \n", err)
			return err
		}

		cleanPath, _ := strings.CutPrefix(path, "/app/")

		if filepath.Ext(path) == ".html" {
			partials = append(partials, cleanPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the templates path: %w", err)
	}

	// walk the template directory and extract all the html files
	err = filepath.Walk(pathDir, func(path string, info fs.FileInfo, err error) error {

		// check if this file exists
		if err != nil {
			if err == os.ErrNotExist {
				log.Printf("error walking the templates path: %v \n", err)
				return err
			}

			log.Printf("error walking the templates path: %v \n", err)
			return err
		}

		// skip all partials since those are set by glob
		if info.IsDir() && info.Name() == PATH_TO_PARTIALS {
			return filepath.SkipDir
		}

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
		templateName := strings.Replace(filepath.Base(templ), ".html", "", 1)
		templatePath, _ := strings.CutPrefix(templ, "/app/")
		templatesAndPartials := partials

		templatesAndPartials = append(templatesAndPartials, templatePath)

		parsedTemplate, err := template.New(templateName).Funcs(functions).ParseFiles(templatesAndPartials...)

		if err != nil {
			return nil, fmt.Errorf("error parsing %s template: %w", templateName, err)
		}

		// Make sure we've actually parsed the named template
		if parsedTemplate.Lookup(templateName) == nil {
			return nil, fmt.Errorf("template %s not found in parsed content", templateName)
		}

		templateCache[templateName] = parsedTemplate
	}

	return templateCache, nil
}
