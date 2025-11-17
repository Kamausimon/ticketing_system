package main

import (
	"fmt"
	"net/http"
	"ticketing_system/internal/auth"
	"ticketing_system/internal/database"

	"github.com/gorilla/mux"
)

func main() {
	DB := database.Init()
	authHandler := auth.NewAuthHandler(DB)
	router := mux.NewRouter()

	//routes
	router.HandleFunc("/register", authHandler.RegisterUser).Methods(http.MethodPost)

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("\nserver starting on port 8080")
	server.ListenAndServe()

}
