package gcs

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

var TYPE_DIR = 0
var TYPE_FILE = 1

type FileInfo struct {
	name  string
	ftype int
	size  int64
	mtime time.Time
}

// ListObjects lists all objects within a bucket.
func ListObjects(bkt *storage.BucketHandle, ctx context.Context) []FileInfo {
	query := &storage.Query{
		Prefix: "",
	}

	var files []FileInfo

	iter := bkt.Objects(ctx, query)
	for {
		attrs, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		if attrs.Name[len(attrs.Name)-1] != '/' {
			files = append(files, FileInfo{attrs.Name, TYPE_FILE, attrs.Size, attrs.Updated})
		}

	}
	return files
}
