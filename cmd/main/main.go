package main

import (
	"fmt"
	"net/http"

	"github.com/yodarango/gooava/api"
)

var PORT string = ":8003"

func main() {

	handler := api.Routes()

	server := &http.Server{
		Addr:    PORT,
		Handler: handler,
	}
	// start the server
	fmt.Println("Starting server on port, ", PORT)
	server.ListenAndServe()

}
