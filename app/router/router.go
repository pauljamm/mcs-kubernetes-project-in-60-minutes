package router

import (
	"mcs-kubernetes-project/middleware"
	"github.com/gorilla/mux"
)

func Router() *mux.Router {

	router := mux.NewRouter()

	router.HandleFunc("/health", middleware.HealthCheck).Methods("GET")
	router.HandleFunc("/users/{id}", middleware.GetUser).Methods("GET", "OPTIONS")
	router.HandleFunc("/users", middleware.GetUsers).Methods("GET", "OPTIONS")
	router.HandleFunc("/users", middleware.CreateUser).Methods("POST", "OPTIONS")
	router.HandleFunc("/users/{id}", middleware.UpdateUser).Methods("PUT", "OPTIONS")
	router.HandleFunc("/users/{id}", middleware.DeleteUser).Methods("DELETE", "OPTIONS")

	return router
}
