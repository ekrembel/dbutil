package src

import (
	"context"
	"dbutil/src/config"
	logger "dbutil/src/logging"
	"dbutil/src/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func ConnectToDB() (*mongo.Client, error) {
	appConfig := config.GetConfig()
	if appConfig.ConnectionString == "" {
		return nil, errors.New("No connection string")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		appConfig.ConnectionString,
	))
	if err != nil {
		log.Fatal(err)
	}
	return client, err
}

func getDBCollection(collectionName string, client *mongo.Client) *mongo.Collection {
	if client != nil {
		collection := client.Database("CoinDB").Collection(collectionName)
		if collection != nil {
			return collection
		}
	}
	return nil
}

func GetUserHash(email string, client *mongo.Client) (string, error) {
	credentials := models.UserCredentials{}
	logger.Info("Searching user with email: " + email)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"email": bson.M{"$eq": email}}
	collection := getDBCollection("Users", client)
	opts := options.FindOne().SetProjection(bson.D{{Key: "email", Value: 1}, {Key: "hash", Value: 1}})

	err := collection.FindOne(ctx, filter, opts).Decode(&credentials)

	if err != nil {
		logger.Error("Unable to get user credentials: " + err.Error())
		return "", err
	}
	hash := credentials.Hash
	logger.Info("Retrieved user hash.")
	logger.Info("hash: " + hash)
	return hash, nil
}

func GetDbIdByEmail(email string, client *mongo.Client) (string, error) {
	objId := models.UserID{}

	logger.Info("Searching user with email: " + email)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"email": bson.M{"$eq": email}}
	collection := getDBCollection("Users", client)
	opts := options.FindOne().SetProjection(bson.D{{Key: "_id", Value: 1}})

	err := collection.FindOne(ctx, filter, opts).Decode(&objId)

	if err != nil {
		logger.Error("Unable to get Id: " + err.Error())
		return "", err
	}
	logger.Info("Object ID: " + objId.ID)

	return objId.ID, nil
}

func CheckIfEmailExists(email string, client *mongo.Client) (bool, error) {
	logger.Info("Looking up user with email: " + email)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"email": bson.M{"$eq": email}}
	collection := getDBCollection("Users", client)
	number, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Error("Encountered error while looking up email")
		return true, err
	}
	if number != 0 {
		logger.Info("Email already exists.")
		return true, nil
	}
	return false, nil
}

func SaveNewUser(user models.User, client *mongo.Client) (*mongo.InsertOneResult, error) {
	user.EmailConfimed = false

	collection := getDBCollection("Users", client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, user)

	if err != nil {
		logger.Error("Encountered error while saving user data. " + err.Error())
		return nil, err
	}
	byte, _ := json.Marshal(result)
	logger.Info("Successfully saved user data - " + string(byte))

	return result, nil
}

func GetUserData(email string, client *mongo.Client) (models.User, error) {
	logger.Info("Searching user with username: " + email)

	user := models.User{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	collection := getDBCollection("Users", client)
	filter := bson.M{"email": email}

	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		logger.Error("User does not exist. " + err.Error())
		return user, err
	}

	logger.Info("Successfully retrieved user data")
	return user, nil
}

func AuthenticateUserOnDB(email string, password string, client *mongo.Client) error {
	hashedPassword := []byte(password)
	hash, err := GetUserHash(email, client)
	if err != nil {
		logger.Error("Unable to get user hash.")
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), hashedPassword)
	if err != nil {
		logger.Error("Unable to authenticate the user")
		return err
	}
	logger.Info("User has been authenticated successfully.")
	return nil
}

func DeleteUserFromDB(email string, client *mongo.Client) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := getDBCollection("Users", client)
	filter := bson.M{"email": bson.M{"$eq": email}}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		logger.Error("Unable to delete user from db: " + err.Error())
		return nil, err
	}
	logger.Info("User has been deleted successfully.")

	return result, nil
}

func UpdateUserStatusOnDB(email string, status string, client *mongo.Client) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := getDBCollection("Users", client)
	filter := bson.M{"email": bson.M{"$eq": email}}
	update := bson.M{"$set": bson.M{"accountStatus": status}}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Unable to update the status of user " + err.Error())
		return nil, err
	}
	return result, nil
}

func SaveBaughtShare(email string, share models.Share, client *mongo.Client) (*mongo.UpdateResult, error) {
	balance, err := GetBalance(email, client)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	if balance < share.PriceBaught*float64(share.Quantity) {
		logger.Error("Insufficient balance to purchase the shares")
		return nil, fmt.Errorf("Insufficient balance to purchase the shares.")
	}

	cost := share.PriceBaught * float64(share.Quantity)
	err = UpdateBalance(client, email, cost, false, true)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := GetUserData(email, client)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	share.ShareID = primitive.NewObjectID().Hex()
	share.SoldIndicator = "N"
	share.DateBaught = time.Now().String()

	if user.Shares == nil {
		shares := []models.Share{}
		shares = append(shares, share)

		collection := getDBCollection("Users", client)
		filter := bson.M{"email": bson.M{"$eq": email}}
		update := bson.M{"$set": bson.M{"shares": shares}}
		result, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			logger.Error("Unable to save baught share" + err.Error())
			_ = UpdateBalance(client, email, cost, true, false)
			return nil, err
		}
		return result, nil

	}
	collection := getDBCollection("Users", client)
	filter := bson.M{"email": bson.M{"$eq": email}}
	update := bson.M{"$push": bson.M{"shares": share}}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Unable to save baught share" + err.Error())
		_ = UpdateBalance(client, email, cost, true, false)
		return nil, err
	}

	return result, nil
}

func UpdateShareToSold(email string, share models.Share, client *mongo.Client) (*mongo.UpdateResult, error) {
	shareID := share.ShareID
	result := &mongo.UpdateResult{}
	soldIndicator, err := GetSoldIndicator(email, shareID, client)
	if err != nil {
		logger.Error(err.Error())
		return result, err
	}
	if soldIndicator == "N" {
		cost := share.PriceSold * float64(share.Quantity)
		err = UpdateBalance(client, email, cost, true, false)
		if err != nil {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		date := time.Now().String()
		collection := getDBCollection("Users", client)
		filter := bson.M{"email": email, "shares.shareID": shareID}
		update := bson.M{"$set": bson.M{"shares.$.ownedOrSold": "Sold", "shares.$.dateSold": date, "shares.$.soldIndicator": "Y", "shares.$.priceSold": share.PriceSold}}
		result, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			logger.Error("Unable to save sold share" + err.Error())
			_ = UpdateBalance(client, email, cost, false, true)
			return nil, err
		}

		return result, nil
	}
	return result, fmt.Errorf("Unable to complete the transaction. User does not own the shares.")
}

func GetSoldIndicator(email string, shareID string, client *mongo.Client) (string, error) {
	shares := models.Shares{}
	soldIndicator := ""
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := getDBCollection("Users", client)
	filter := bson.M{"email": email}
	opts := options.FindOne().SetProjection(bson.M{"shares": 1})
	err := collection.FindOne(ctx, filter, opts).Decode(&shares)
	if err != nil {
		logger.Error("Unable to get sold indicator for the share " + err.Error())
		return soldIndicator, err
	}
	for i := 0; i < len(shares.Shares); i++ {
		if shares.Shares[i].ShareID == shareID {
			soldIndicator = shares.Shares[i].SoldIndicator
		}
	}
	logger.Info(shares)
	return soldIndicator, nil
}

func GetBalance(email string, client *mongo.Client) (float64, error) {
	balance := models.Balance{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := getDBCollection("Users", client)
	filter := bson.M{"email": bson.M{"$eq": email}}
	opts := options.FindOne().SetProjection(bson.D{{Key: "balance", Value: 1}})
	err := collection.FindOne(ctx, filter, opts).Decode(&balance)
	if err != nil {
		logger.Error("Unable to update the status of user " + err.Error())
		return balance.Balance, err
	}
	return balance.Balance, nil
}

func UpdateBalance(client *mongo.Client, email string, amountToAddOrDeduct float64, toAdd bool, toDeduct bool) error {
	currentBalance, err := GetBalance(email, client)
	newBalance := float64(0)
	if err != nil {
		logger.Error("Unable to get current balance " + err.Error())
		return err
	}
	if toAdd {
		newBalance = currentBalance + amountToAddOrDeduct
	}
	if toDeduct {
		newBalance = currentBalance - amountToAddOrDeduct
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := getDBCollection("Users", client)
	filter := bson.M{"email": bson.M{"$eq": email}}
	update := bson.M{"$set": bson.M{"balance": newBalance}}
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Unable to save baught share" + err.Error())
		return err
	}

	logger.Info("Balance has been updated successfully.")
	return nil
}
