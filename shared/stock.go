package shared

type Item struct {
	ID          string   `bson:"_id,omitempty"`
	Name        string   `bson:"name,omitempty"`
	Link        string   `bson:"link,omitempty"`
	Description string   `bson:"description,omitempty"`
	Number      int32    `bson:"number,omitempty"`
	Price       float32  `bson:"price,omitempty"`
	Poster      string   `bson:"poster,omitempty"`
	Photos      []string `bson:"photos,omitempty"`
}
