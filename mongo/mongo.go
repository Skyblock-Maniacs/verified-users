package mongo

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type GlobalUser struct {
	Id string `json:"_id" bson:"_id"`
	Uuid string `json:"uuid" bson:"uuid"`
}
type UserKeyData struct {
	Id string    `json:"_id" bson:"_id"`
	ApiKey apiKey `json:"apiKey" bson:"apiKey"`
}
type apiKey struct {
	Key string            `json:"key" bson:"key"`
	Permissions []string  `json:"permissions" bson:"permissions"`
}

var MongoClient *mongo.Client

func Init() {
	c, err := mongo.Connect(
		context.TODO(),
		options.Client().ApplyURI(os.Getenv("MONGO_URI")),
	)

	if err != nil {
		log.Fatal(err)
	}

	MongoClient = c

	if err := MongoClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
	log.Println("Successfully connected to MongoDB")
}

func GetMongoClient() *mongo.Client {
	return MongoClient
}

func GetApiKeyData(key string) (UserKeyData, error) {
	database := MongoClient.Database(os.Getenv("MONGO_DATABASE"))
	collection := database.Collection(os.Getenv("MONGO_KEY_COLLECTION"))

	opts := options.FindOne().
		SetProjection(bson.D{{Key: "apiKey", Value: 1}, {Key: "_id", Value: 1}})

	var result UserKeyData
	err := collection.FindOne(
		context.TODO(),
		bson.M{"apiKey.key": key},
		opts,
	).Decode(&result)

	return result, err
}

func GetUserByUUID(uuid string) (GlobalUser, error) {
	database := MongoClient.Database(os.Getenv("MONGO_DATABASE"))
	collection := database.Collection(os.Getenv("MONGO_COLLECTION"))

	var result GlobalUser
	err := collection.FindOne(
		context.TODO(),
		bson.M{"uuid": uuid},
	).Decode(&result)

	return result, err
}

func GetUserByDiscordID(id string) (GlobalUser, error) {
	database := MongoClient.Database(os.Getenv("MONGO_DATABASE"))
	collection := database.Collection(os.Getenv("MONGO_COLLECTION"))

	var result GlobalUser
	err := collection.FindOne(
		context.TODO(),
		bson.M{"_id": id},
	).Decode(&result)

	return result, err
}

func InsertUser(uuid string, discordId string) error {
	database := MongoClient.Database(os.Getenv("MONGO_DATABASE"))
	collection := database.Collection(os.Getenv("MONGO_COLLECTION"))

	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{
			"_id": discordId,
		},
		bson.M{
			"$set": bson.M{
				"_id": discordId,
				"uuid": uuid,
			},
		},
		opts,
	)

	return err
}

func DeleteUserViaUUID(uuid string) error {
	database := MongoClient.Database(os.Getenv("MONGO_DATABASE"))
	collection := database.Collection(os.Getenv("MONGO_COLLECTION"))

	_, err := collection.DeleteOne(
		context.TODO(),
		bson.M{"uuid": uuid},
	)

	return err
}

func DeleteUserViaDiscordID(id string) error {
	database := MongoClient.Database(os.Getenv("MONGO_DATABASE"))
	collection := database.Collection(os.Getenv("MONGO_COLLECTION"))

	_, err := collection.DeleteOne(
		context.TODO(),
		bson.M{"_id": id},
	)

	return err
}