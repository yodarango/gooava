package main

import (
	"fmt"
	"net/http"

	"github.com/yodarango/gooava/api"
	"github.com/yodarango/gooava/config"
	"github.com/yodarango/gooava/db"
)

var PORT string = ":8003"

// app singleton
var appConfig config.AppConfig

// db singleton
var dbConfig db.SQLConfig

func main() {

	handler := api.Routes()

	server := &http.Server{
		Addr:    PORT,
		Handler: handler,
	}

	Init()

	// close the db connection
	dbConfig.DB.Close()

	// start the server
	fmt.Println("Starting server on port, ", PORT)
	server.ListenAndServe()

}
