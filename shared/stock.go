package shared

type Item struct {
	ID          string   `bson:"_id,omitempty"`
	Name        string   `bson:"name,omitempty"`
	Description string   `bson:"description,omitempty"`
	Number      int32    `bson:"number,omitempty"`
	Price       float32  `bson:"price,omitempty"`
	Photos      []string `bson:"photos,omitempty"`
}
