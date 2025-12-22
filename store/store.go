package store

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type DomainStore struct {
	DB mongo.Collection
}

func NewDomainStore() *DomainStore {
	return &DomainStore{
		DB: *mongo.Collection,
	}
}
