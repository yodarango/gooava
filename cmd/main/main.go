package main

import (
	"fmt"
	"net/http"

	"github.com/yodarango/gooava/api"
	"github.com/yodarango/gooava/config"
	"github.com/yodarango/gooava/db"
	"github.com/yodarango/gooava/internal/repo"
)

var PORT string = ":8003"

// app singleton
var appConfig config.AppConfig

// db singleton
var dbConfig db.DbConfig

// AppRepo that holds all global app singletons
var appRepo repo.AppRepo

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
