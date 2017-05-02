package shelf

import (
	"net/http"
	"strings"
	"errors"
	"os"
	"path"
)

var (
	ErrBucketNotFound = errors.New("Bucket not found")
)

// Shelf is collection of Bucket
// Bucket has responsibility for storage files.
// Shelf will manage buckets. (for example, periodically preform cleanup and consistency check)
// And also, perform as static file handler
type Shelf struct {
	buckets    map[string]*Bucket
	metaDir    string
	storageDir string
	metaPrefix string
}

func (s *Shelf) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		// Only get is supported
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p := strings.TrimPrefix(r.URL.EscapedPath(), "/")
	firstSlash := strings.Index(p, "/")
	if firstSlash < 0 {
		// GetBucket is not supported
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bucketName := p[:firstSlash]
	key := p[firstSlash+1:]
	if len(key) == 0 {
		// key == "" is defined, but forbid to avoid ambiguous path
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// GET Only
	bucket, err := s.Bucket(bucketName)
	if err != nil {
		if err == ErrBucketNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	var meta BlobMeta
	if err := bucket.LoadMeta([]byte(key), &meta); err != nil {
		if err == ErrKeyNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	f, err := os.Open(path.Join(bucket.storageDir, meta.StoragePath()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	http.ServeContent(w, r, p, meta.CreatedAt, f)
}

func (s *Shelf) Bucket(key string) (*Bucket, error) {
	if bkt, ok := s.buckets[key]; ok {
		return bkt, nil
	}
	metaName := key
	if len(s.metaPrefix) > 0 {
		metaName = s.metaPrefix + key
	}
	bkt, err := NewBucket(path.Join(s.metaDir, metaName), path.Join(s.storageDir, key))
	if err != nil {
		return nil, err
	}
	s.buckets[key] = bkt
	return bkt, nil
}

// Create new shelf.
// To achieve better performance,
// metaRoot should be on SSD,
// and storageRoot should be on HDD
func New(metaRoot, storageRoot string) *Shelf {
	slf := &Shelf{
		buckets:    make(map[string]*Bucket),
		metaDir:    metaRoot,
		storageDir: storageRoot,
		metaPrefix: "",
	}
	if metaRoot == storageRoot {
		slf.metaPrefix = "meta_"
	}
	return slf
}
