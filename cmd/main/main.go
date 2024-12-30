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

	handler := api.Routes()

	server := &http.Server{
		Addr:    PORT,
		Handler: handler,
	}

	Init()
	// start the server
	fmt.Println("Starting server on port, ", PORT)
	server.ListenAndServe()

}
