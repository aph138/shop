package shared

type User struct {
	ID       string `bson:"_id,omitempty"`
	Username string `bson:"username,omitempty"`
	Email    string `bosn:"email,omitempty"`
	Password string `bson:"password,omitempty"`
	Role     uint32 `bson:"role,omitempty"`
	Status   bool   `bson:"status,omitempty"`
}
