package config

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var ctx context.Context
var App *firebase.App
var DB *firestore.Client

func InitFirebase() {
	ctx = context.Background()

	jsonCredentials := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	if jsonCredentials == "" {
		log.Fatal("FIREBASE_CREDENTIALS_JSON environment variable not set")
	}

	opt := option.WithCredentialsJSON([]byte(jsonCredentials))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Firebase initialization failed: %v", err)
	}
	App = app

	client, err := App.Firestore(ctx)
	if err != nil {
		log.Fatalf("Firestore connection failed: %v", err)
	}
	DB = client

	log.Println("Firebase initialized successfully!")
}
