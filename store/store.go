package store

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DomainStore struct {
	DB mongo.Collection
}

func NewDomainStore() *DomainStore {
	uri := "mongodb+srv://<username>:<password>@cluster0.example.mongodb.net/?retryWrites=true&w=majority"
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &DomainStore{
		DB: *client.Database("your_database_name").Collection("your_collection_name"),
	}
}

func (ds *DomainStore) Close() error {
	// disconnect
	if err := ds.DB.Database().Client().Disconnect(context.TODO()); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
