package auth

import (
	"crypto"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	private crypto.PrivateKey
}

func NewIssuer(path string) (*Issuer, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p, err := jwt.ParseEdPrivateKeyFromPEM(file)
	if err != nil {
		return nil, err
	}
	return &Issuer{private: p}, nil
}

type myClaims struct {
	Value map[string]string `json:"value,omitempty"`
	jwt.RegisteredClaims
}

// expireDate in minute
func (i *Issuer) Token(v map[string]string, expireDate int) (string, error) {
	claims := myClaims{
		Value: v,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(expireDate))),
		},
	}
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, claims, nil)
	t, err := token.SignedString(i.private)
	if err != nil {
		return "", err
	}
	return t, nil

}

type Validator struct {
	public crypto.PublicKey
}

func NewValidator(path string) (*Validator, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p, err := jwt.ParseEdPublicKeyFromPEM(file)
	if err != nil {
		return nil, err
	}
	return &Validator{public: p}, err
}

func (v *Validator) Validate(tokenString string) (map[string]string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &myClaims{}, func(t *jwt.Token) (interface{}, error) {
		// check if the token use the correct method
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("jwt: wrong signing method")
		}
		return v.public, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*myClaims)
	if !ok {
		return nil, nil
	}
	return claims.Value, nil
}
