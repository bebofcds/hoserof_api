package config

import (
	"errors"
	"os"
	"strings"
	"sync"

	storage_go "github.com/supabase-community/storage-go"
)

type Supabase struct {
	Storage *storage_go.Client
}

var (
	supabaseInstance *Supabase
	Once             sync.Once
	initError        error
)

func NewSupabaseInstance() (*Supabase, error) {
	Once.Do(func() {
		supabaseInstance, initError = initializeSupabase()
	})
	return supabaseInstance, initError
}

func initializeSupabase() (*Supabase, error) {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_SERVICE_KEY")

	if url == "" || key == "" {
		return nil, errors.New("SUPABASE_URL and SUPABASE_SERVICE_KEY must be set")
	}
	url = strings.TrimRight(url, "/")

	client := storage_go.NewClient(url+"/storage/v1", key, nil)

	return &Supabase{
		Storage: client,
	}, nil
}
