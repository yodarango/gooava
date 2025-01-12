package main

import (
	"log"
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

	appRepo := Init()

	// close the db connection
	defer appRepo.DB.Connection.Close()

	// start the server
	log.Println("Starting server on port, ", PORT)
	server.ListenAndServe()

}
