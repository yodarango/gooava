package apiv1

import (
	"log"
	"net/http"

	"github.com/yodarango/gooava/internal/utils"
)

func (c *ApiConfiguration) Home(w http.ResponseWriter, r *http.Request) {
	var template utils.TemplateRenderer
	template.Name = "index"
	template.Title = "Home"

	err := template.Render(w)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}
