package main

import (
	db "dbutil/src/database"
	"dbutil/src/handlers"
	logger "dbutil/src/logging"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	client, err := db.ConnectToDB()
	if err != nil {
		logger.Error(err)
	}
	logger.Info(client)
	logger.Info("Connected to mongodb...")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/user/register", handlers.Register(client)).Methods("POST")
	router.HandleFunc("/user/{email}", handlers.GetUser(client)).Methods("GET")
	router.HandleFunc("/user/update/{email}/{status}", handlers.UpdateUserStatus(client)).Methods("PUT")
	router.HandleFunc("/user/delete/{email}/{password}", handlers.DeleteUser(client)).Methods("DELETE")
	router.HandleFunc("/user/authenticate/{email}/{password}", handlers.AuthenticateUser(client)).Methods("GET")
	router.HandleFunc("/user/share/{email}/{transactiontype}", handlers.SaveShare(client)).Methods("PUT")
	router.HandleFunc("/user/update/emailconfirmation/{email}", handlers.ConfirmEmail(client)).Methods("PUT")
	router.HandleFunc("/user/update/addbalance/{email}/{amount}", handlers.AddToBalance(client)).Methods("PUT")

	logger.Info("dbutil is running")
	log.Fatal(http.ListenAndServe(":8080", router))

}
