// Package config provides initialization logic for Firebase and Firestore.
//
// This package is responsible for:
//   - Loading Firebase credentials from environment variables
//   - Initializing Firebase App
//   - Creating Firestore client
//   - Managing lifecycle and cleanup
//
// Usage:
//
//	ctx := context.Background()
//	fb, err := config.NewFirebase(ctx)
//	if err != nil {
//	    log.Fatalf("failed to init firebase: %v", err)
//	}
//	defer fb.Close()

package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var ctx context.Context

type Firebase struct {
	App *firebase.App
	DB  *firestore.Client
}

var (
	instance *Firebase
	once     sync.Once
	initErr  error
)

func NewFirebaseInstance(ctx context.Context) (*Firebase, error) {
	once.Do(func() {
		instance, initErr = initializeFirebase(ctx)
	})
	return instance, initErr
}

func initializeFirebase(ctx context.Context) (*Firebase, error) {

	credentials := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	if credentials == "" {
		return nil, errors.New("FIREBASE_CREDENTIALS_JSON environment variable not set")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	opt := option.WithCredentialsJSON([]byte(credentials))
	app, err := firebase.NewApp(timeoutCtx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase initialization failed: %w", err)
	}
	client, err := app.Firestore(timeoutCtx)
	if err != nil {
		return nil, fmt.Errorf("firestore connection failed: %w", err)
	}

	return &Firebase{
		App: app,
		DB:  client,
	}, nil

}
func (f *Firebase) Close() error {
	if f.DB != nil {
		return f.DB.Close()
	}
	return nil
}
