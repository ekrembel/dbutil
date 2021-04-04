package handlers

import (
	db "dbutil/src/database"
	logger "dbutil/src/logging"
	"dbutil/src/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func Register(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Adding new user")
		user := models.User{}

		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		checkResult, err := db.CheckIfEmailExists(user.Email, client)

		if err != nil {
			http.Error(w, "Error while checking if email already exists. "+err.Error(), http.StatusInternalServerError)
			return
		}

		if checkResult {
			http.Error(w, "Email already in use.", http.StatusBadRequest)
			return
		}

		password := user.Hash
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)

		if err != nil {
			http.Error(w, "Unable to hash the password", http.StatusInternalServerError)
			return
		}

		user.Hash = string(hashedPassword)

		result, err := db.SaveNewUser(user, client)
		if err != nil {
			http.Error(w, "Error while saving the member to db.", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(result)
	}
}

func GetUser(client *mongo.Client) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		email := params["email"]
		if email == "" {
			http.Error(rw, "username must be included in the search parameters", http.StatusBadRequest)
			return
		}

		user, err := db.GetUserData(email, client)
		if err != nil {
			http.Error(rw, "User does not exist.", http.StatusNotFound)
			return
		}
		rw.Header().Set("content-type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(user)
	}
}

func DeleteUser(client *mongo.Client) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		logger.Info("Attempting to delete user from db.")
		params := mux.Vars(r)
		email := params["email"]
		password := params["password"]

		if email == "" || password == "" {
			http.Error(rw, "Email or password is missing.", http.StatusBadRequest)
			return
		}
		err := db.AuthenticateUserOnDB(email, password, client)
		if err != nil {
			http.Error(rw, "Unable to authenticate: "+err.Error(), http.StatusUnauthorized)
			return
		}

		result, err := db.DeleteUserFromDB(email, client)
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(result)
	}
}

func UpdateUserStatus(client *mongo.Client) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		status := params["status"]
		email := params["email"]

		if status == "" || email == "" {
			http.Error(rw, "email or status is not present in the url", http.StatusBadRequest)
			return
		}

		checkUser, err := db.CheckIfEmailExists(email, client)
		if err != nil {
			http.Error(rw, "Error while checking the user", http.StatusInternalServerError)
			return
		}
		if !checkUser {
			http.Error(rw, "Unable to update. User does not exists", http.StatusBadRequest)
			return
		}

		result, err := db.UpdateUserStatusOnDB(email, status, client)

		if err != nil {
			http.Error(rw, "Unable to update active indicator of user. "+err.Error(), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(result)
	}
}

func AuthenticateUser(client *mongo.Client) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		email := params["email"]
		password := params["password"]

		if email == "" || password == "" {
			http.Error(rw, "Email or password is missing.", http.StatusBadRequest)
			return
		}
		logger.Info("Attempting to authenticate user.")
		err := db.AuthenticateUserOnDB(email, password, client)
		if err != nil {
			http.Error(rw, "Unable to authenticate: "+err.Error(), http.StatusUnauthorized)
			return
		}

		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode("User has been authenticated successfully.")
	}
}

func SaveShare(client *mongo.Client) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		share := models.Share{}
		params := mux.Vars(r)
		email := params["email"]
		transactionType := params["transactiontype"]
		if email == "" {
			http.Error(rw, "Email is missing.", http.StatusBadRequest)
			return
		}

		err := json.NewDecoder(r.Body).Decode(&share)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}

		result := &mongo.UpdateResult{}
		if transactionType == "buy" {
			result, err = db.SaveBaughtShare(email, share, client)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
		}
		if transactionType == "sell" {
			result, err = db.UpdateShareToSold(email, share, client)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
		}

		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(result)
	}
}

func ConfirmEmail(client *mongo.Client) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

	}
}

func AddToBalance(client *mongo.Client) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		email := params["email"]
		amountToAdd := params["amount"]
		if email == "" || amountToAdd == "" {
			http.Error(rw, "Email or amount is missing in the url", http.StatusBadRequest)
			return
		}

		amount, err := strconv.ParseFloat(amountToAdd, 64)
		if err != nil {
			http.Error(rw, "Failed while parsing the amount to float", http.StatusInternalServerError)
			return
		}

		err = db.UpdateBalance(client, email, amount, true, false)
		if err != nil {
			http.Error(rw, "Unable to add balance", http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode("Amount has been added to the balance successfully.")
	}
}
