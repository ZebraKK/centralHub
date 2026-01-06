package store

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"centralHub/logger"
	models "centralHub/model"
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
		logger.RunLogger.Error().Err(err).Msg("Failed to connect to MongoDB")
		return nil
	}

	logger.RunLogger.Info().Msg("Connected to MongoDB successfully")
	return &DomainStore{
		DB: *client.Database("your_database_name").Collection("your_collection_name"),
	}
}

func (ds *DomainStore) Close() error {
	if err := ds.DB.Database().Client().Disconnect(context.TODO()); err != nil {
		logger.RunLogger.Error().Err(err).Msg("Failed to disconnect MongoDB client")
		return err
	}
	logger.RunLogger.Info().Msg("MongoDB client disconnected")
	return nil
}

func (ds *DomainStore) Insert(ctx context.Context, domain models.XLDomain) error {
	logger.RunLogger.Info().Str("domain", domain.Name).Msg("Inserting domain")
	_, err := ds.DB.InsertOne(ctx, domain)
	if err != nil {
		logger.RunLogger.Error().Err(err).Str("domain", domain.Name).Msg("Insert domain failed")
	}
	return err
}

func (ds *DomainStore) FindByID(ctx context.Context, id string) (*models.XLDomain, error) {
	logger.RunLogger.Info().Str("id", id).Msg("Finding domain by ID")
	var domain models.XLDomain
	err := ds.DB.FindOne(ctx, bson.M{"_id": id}).Decode(&domain)
	if err != nil {
		logger.RunLogger.Error().Err(err).Str("id", id).Msg("Find domain failed")
		return nil, err
	}
	logger.RunLogger.Info().Str("id", id).Msg("Domain found")
	return &domain, nil
}

func (ds *DomainStore) Update(ctx context.Context, id string, update bson.M) error {
	logger.RunLogger.Info().Str("id", id).Interface("update", update).Msg("Updating domain")
	_, err := ds.DB.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		logger.RunLogger.Error().Err(err).Str("id", id).Msg("Update domain failed")
	}
	return err
}

func (ds *DomainStore) Delete(ctx context.Context, id string) error {
	logger.RunLogger.Info().Str("id", id).Msg("Deleting domain")
	_, err := ds.DB.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		logger.RunLogger.Error().Err(err).Str("id", id).Msg("Delete domain failed")
	}
	return err
}
