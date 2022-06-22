package gcs

import (
	"context"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

var BUCKET = "yirzhou"
var MNT_DIR = "test_gcs"

type GCSNode struct {
	fs.Inode
	bucket *storage.BucketHandle
	mu     sync.Mutex
}

// We need a dynamically discovered file system.
// See https://pkg.go.dev/github.com/hanwen/go-fuse/v2/fs#hdr-Dynamically_discovered_file_systems
type FileNode struct {
	fs.Inode
	bucket  *storage.BucketHandle
	mu      sync.Mutex
	info    FileInfo
	content []byte
}

var _ = (fs.InodeEmbedder)((*FileNode)(nil))

var _ = (fs.NodeOnAdder)((*GCSNode)(nil))

var _ = (fs.InodeEmbedder)((*GCSNode)(nil))

var _ = (fs.NodeLookuper)((*GCSNode)(nil))

var _ = (fs.NodeReaddirer)((*GCSNode)(nil))

var _ = (fs.NodeGetattrer)((*FileNode)(nil))

// TODO
var _ = (fs.NodeOpener)((*FileNode)(nil))

// TODO
var _ = (fs.NodeReader)((*FileNode)(nil))

// TODO
var _ = (fs.NodeWriter)((*FileNode)(nil))

// TODO
var _ = (fs.NodeSetattrer)((*FileNode)(nil))

func (bn *FileNode) getattr(out *fuse.AttrOut) {
	out.Size = uint64(bn.info.size)
	out.SetTimes(nil, &bn.info.mtime, nil)
}

func (bn *FileNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	if sz, ok := in.GetSize(); ok {
		bn.resize(sz)
	}
	bn.getattr(out)
	return 0
}

func (bn *FileNode) resize(sz uint64) {
	if sz > uint64(cap(bn.content)) {
		n := make([]byte, sz)
		copy(n, bn.content)
		bn.content = n
	} else {
		bn.content = bn.content[:sz]
	}
	bn.info.mtime = time.Now()
}

// TODO: Ignore this for now. Focus is on read.
func (bn *FileNode) Write(ctx context.Context, fh fs.FileHandle, buf []byte, off int64) (uint32, syscall.Errno) {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	sz := int64(len(buf))
	if off+sz > int64(len(bn.content)) {
		bn.resize(uint64(off + sz))
	}
	copy(bn.content[off:], buf)
	bn.info.mtime = time.Now()
	return uint32(sz), 0
}

// Load data from GCS to data.
func (node *FileNode) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	// Load data from GCS
	rc, err := node.bucket.Object(node.info.name).NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()
	body, err := io.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}
	node.content = body
	return nil, 0, 0
}

// TODO
func (node *FileNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	node.mu.Lock()
	defer node.mu.Unlock()

	// TODO: Temporary fix for read
	if off >= int64(len(node.content)) {
		return fuse.ReadResultData(node.content), 0
	}

	end := off + int64(len(dest))
	if end > int64(len(node.content)) {
		end = int64(len(node.content))
	}
	resRead := node.content[off:end]
	log.Printf("Read [%d] bytes within [%d:%d]", len(resRead), off, end)
	// We could copy to the `dest` buffer, but since we have a
	// []byte already, return that.
	return fuse.ReadResultData(resRead), 0
}

func (root *GCSNode) EmbeddedInode() *fs.Inode {
	return &root.Inode
}

// Lookup implements returns the inode based on the name.
func (root *GCSNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	log.Printf("Looking up for [%s]\n", name)
	dir, base := filepath.Split(name)
	root.mu.Lock()
	defer root.mu.Unlock()
	node := &root.Inode
	var ch *fs.Inode
	for _, comp := range strings.Split(dir, "/") {
		if len(comp) == 0 {
			continue
		}
		ch = node.GetChild(comp)
		if ch == nil {
			// It is displayed if there is no such file or directory exists.
			return nil, syscall.ENOENT
		}
		node = ch
	}

	ch = node.GetChild(base)
	if ch == nil {
		return nil, syscall.ENOENT
	}
	return ch, 0
}

func (n *GCSNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	n.mu.Lock()
	defer n.mu.Unlock()
	node := &n.Inode
	entries := make([]fuse.DirEntry, 0, len(node.Children()))
	for name, inode := range node.Children() {
		entry := fuse.DirEntry{
			Name: name,
			Ino:  inode.StableAttr().Ino,
			Mode: inode.StableAttr().Mode,
		}
		entries = append(entries, entry)
	}

	return fs.NewListDirStream(entries), 0
}

// OnAdd initializes the fs when being mounted.
func (root *GCSNode) OnAdd(ctx context.Context) {
	log.Println("Mounting...")
	// Create new client
	client, _ := NewGCSClient()
	bkt := client.Bucket(BUCKET)
	root.bucket = bkt
	// List Object in bucket
	fileInfoList := ListObjects(bkt, ctx)

	for _, fileInfo := range fileInfoList {
		p := &root.Inode
		dir, base := filepath.Split(fileInfo.name)
		for _, comp := range strings.Split(dir, "/") {
			if len(comp) == 0 {
				continue
			}

			ch := p.GetChild(comp)
			if ch == nil {
				// Create a directory
				ch = p.NewPersistentInode(
					ctx, &fs.Inode{},
					fs.StableAttr{Mode: syscall.S_IFDIR},
				)
				// Always overwrite
				_ = p.AddChild(comp, ch, true)
			}
			p = ch
		}

		// Custom FileNode
		embedded := &FileNode{
			Inode:  fs.Inode{},
			bucket: bkt,
			info: FileInfo{
				name:  base,
				size:  fileInfo.size,
				mtime: fileInfo.mtime,
			},
		}

		ch := p.NewPersistentInode(ctx, embedded, fs.StableAttr{})
		p.AddChild(base, ch, true)

	}
}

func (root *FileNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	root.mu.Lock()
	defer root.mu.Unlock()
	root.getattr(out)
	return 0
}
