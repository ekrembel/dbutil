package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Username      string  `bson:"username" json:"username"`
	Email         string  `bson:"email" json:"email"`
	EmailConfimed bool    `bson:"emailConfirmed" json:"emailConfirmed"`
	Phone         string  `bson:"phone" json:"phone"`
	Hash          string  `bson:"hash" json:"hash"`
	FirstName     string  `bson:"firstName" json:"firstName"`
	MiddleName    string  `bson:"middleName" json:"middleName"`
	LastName      string  `bson:"lastName" json:"lastName"`
	AccountStatus string  `bson:"accountStatus" json:"accountStatus"`
	Balance       float64 `bson:"balance" json:"balance"`
	CreatedDate   string  `bson:"createdDate" json:"createdDate"`
	Shares        []Share `bson:"shares" json:"shares"`
}

type UserID struct {
	ID string `bson:"_id" json:"_id"`
}

type UserCredentials struct {
	Email string `bson:"email" json:"email"`
	Hash  string `bson:"hash" json:"hash"`
}

type Balance struct {
	Balance float64 `bson:"balance" json:"balance"`
}

type Shares struct {
	ID     primitive.ObjectID `bson:"_id" json:"_id"`
	Shares []Share            `bson:"shares" json:"shares"`
}

type Share struct {
	ShareID       string  `bson:"shareID" json:"shareID"`
	Symbol        string  `bson:"symbol" json:"symbol"`
	Company       string  `bson:"company" json:"company"`
	Quantity      int     `bson:"quantity" json:"quantity"`
	PriceBaught   float64 `bson:"priceBaught" json:"priceBaught"`
	PriceSold     float64 `bson:"priceSold" json:"priceSold"`
	SoldIndicator string  `bson:"soldIndicator" json:"soldIndicator"`
	DateBaught    string  `bson:"dateBaught" json:"dateBaught"`
	DateSold      string  `bson:"dateSold" json:"dateSold"`
}
