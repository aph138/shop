package shared

type User struct {
	ID       string  `bson:"_id,omitempty"`
	Username string  `bson:"username,omitempty"`
	Email    string  `bosn:"email,omitempty"`
	Password string  `bson:"password,omitempty"`
	Phone    string  `bson:"phone,omitempty"`
	Role     uint32  `bson:"role,omitempty"`
	Status   bool    `bson:"status,omitempty"`
	Address  Address `bson:"address,omitempty"`
}
type Address struct {
	Address string `bson:"address,omitempty"`
	Phone   string `bson:"phone,omitempty"`
}
