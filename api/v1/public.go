package apiv1

import (
	"fmt"
	"html/template"
	"net/http"
)

func Public(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("web/templates/index.html", "web/templates/partials/base.html")

	if err != nil {
		fmt.Print(fmt.Errorf("error parsing template %q", err.Error()))
	}

	temp.Execute(w, nil)
}
