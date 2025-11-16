package services

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var SecretKey = []byte(os.Getenv("JWT_KEY"))

func jwtGenerator(id, class, name string) (string, error) {
	fmt.Println("generating jwt")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading.env file")
		log.Fatal("Error loading .env file")
	}

	// Access environment variables
	var (
		t *jwt.Token
		s string
	)

	t = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_ID":    id,
			"user_class": class,
			"user_name":  name,
		})
	s, err = t.SignedString(SecretKey)
	if err != nil {
		fmt.Println("failed to generate token")
		fmt.Println(err)
		return "", errors.New("failed to generate token")
	}
	fmt.Println(s)
	return s, nil
}
