package models

// reference:

type Domain struct {
	ID     string `bson:"_id,omitempty"`
	Name   string `bson:"name"`
	Owner  string `bson:"owner"`
	Status string `bson:"status"`
}
