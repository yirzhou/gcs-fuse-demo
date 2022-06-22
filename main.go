package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"yirzhou/gcs/gcs"

	"cloud.google.com/go/storage"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Deprecated
func read(client *storage.Client, ctx context.Context) {
	rc, err := client.Bucket(gcs.BUCKET).Object("randfile").NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()
	body, err := io.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("File size: %d\n", len(body))
}

func preStart(mntDir string) {
	// Umount if mounted
	_ = syscall.Unmount(mntDir, 0)
	// Remove
	_ = os.RemoveAll(mntDir)
}

func main() {
	mntDir := fmt.Sprintf("./%s", gcs.MNT_DIR)
	preStart(mntDir)
	err := os.Mkdir(mntDir, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	// Create root
	root := &gcs.GCSNode{}
	server, err := fs.Mount(mntDir, root, &fs.Options{
		MountOptions: fuse.MountOptions{Debug: true},
	})
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Mounted on %s", mntDir)
	log.Printf("Unmount by calling 'fusermount -u %s'", mntDir)

	// Wait until unmount before exiting
	server.Wait()

}
