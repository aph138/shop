package shared

import "go.mongodb.org/mongo-driver/bson/primitive"

type Cart struct {
	ItemID   primitive.ObjectID `bson:"item_id"`
	Quantity uint32             `bson:"quantity"`
}
