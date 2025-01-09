package apiv1

import (
	"fmt"
	"net/http"

	"github.com/yodarango/gooava/internal/utils"
)

func Home(w http.ResponseWriter, r *http.Request) {
	var template utils.TemplateRenderer
	template.Name = "index"
	template.Title = "Home"

	err := template.Render(w)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}
