package shared

type User struct {
	ID       string `bson:"_id,omitempty"`
	Username string `bson:"username"`
	Email    string `bosn:"email"`
	Password string `bson:"password"`
	Role     uint32 `bson:"role"`
	Status   bool   `bson:"status"`
}
