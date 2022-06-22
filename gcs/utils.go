package gcs

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
)

// NewGCSClient creates a new GCS client.
func NewGCSClient() (*storage.Client, context.Context) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return client, ctx
}
