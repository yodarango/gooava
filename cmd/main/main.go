package main

import (
	"fmt"
	"net/http"

	"github.com/yodarango/gooava/api"
	"github.com/yodarango/gooava/config"
)

var PORT string = ":8003"

// app singleton
var appConfig config.AppConfig

func main() {
	// initialize the configuration for all parts of the app.
	Init()

	handler := api.Routes()

	server := &http.Server{
		Addr:    PORT,
		Handler: handler,
	}
	// start the server
	fmt.Println("Starting server on port, ", PORT)
	server.ListenAndServe()

}
