package shelf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

const (
	FilesNumberPerDir = 10000
	DefaultPermission = 0600
)

// Bucket DB Structure
// Specials
//   _tip :: Tip chunk ID
// ChunkDir info
//   c_<CID>
// Blobs
//   b_<key> ::
var (
	RootBucket        = []byte("_root")
	TipKey            = []byte("_tip")
	BlobBucket        = []byte("_blobs")
	TombBucket        = []byte("_tomb")
	TombTimeSeparator = '_'
	ErrKeyNotFound    = errors.New("KeyNotFound")
	ErrEmptyTip       = errors.New("ErrEmptyTip")
)

type Bucket struct {
	meta        *bolt.DB
	storageDir  string
	metaPath    string
	tipMu       sync.Mutex
	tipID       int64
	tipDirCount int
}

func NewBucket(metaPath, storageDir string) (*Bucket, error) {
	if err := os.MkdirAll(storageDir, 0700); err != nil {
		return nil, err
	}
	meta, err := bolt.Open(metaPath, 0700, nil)
	if err != nil {
		return nil, err
	}
	bkt := &Bucket{
		meta:       meta,
		storageDir: storageDir,
		metaPath:   metaPath,
	}
	err = meta.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(RootBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(BlobBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(TombBucket); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := bkt.initTip(); err != nil {
		return nil, err
	}

	if err := bkt.refreshTipDirStatus(); err != nil {
		return nil, err
	}
	return bkt, nil
}

func (b *Bucket) Get(key []byte) (io.ReadSeeker, error) {
	var meta BlobMeta
	if err := b.LoadMeta(key, &meta); err != nil {
		return nil, err
	}
	p := path.Join(b.storageDir, meta.StoragePath())
	return os.Open(p)
}

func (b *Bucket) LoadMeta(blobKey []byte, meta *BlobMeta) error {
	return b.meta.View(func(tx *bolt.Tx) error {
		bb := tx.Bucket(BlobBucket)
		metaBytes := bb.Get(blobKey)
		if metaBytes == nil {
			return ErrKeyNotFound
		}
		if err := meta.Decode(metaBytes); err != nil {
			return err
		}
		return nil
	})
}

func (b *Bucket) Put(key []byte, data io.Reader) error {
	now := time.Now()
	newMeta := BlobMeta{
		Filename:  bytes.Replace(key, []byte("/"), []byte("_"), -1),
		DirID:     b.tipID,
		CreatedAt: now,
	}
	// Write data first
	if err := b.putBlob(&newMeta, data); err != nil {
		return err
	}
	var buf []byte
	if err := newMeta.Encode(&buf); err != nil {
		return err
	}
	// Update metadata file has successfully written
	return b.meta.Update(func(tx *bolt.Tx) error {
		bb := tx.Bucket(BlobBucket)
		prevMeta := bb.Get(key)
		if err := bb.Put(key, buf); err != nil {
			return err
		}
		if prevMeta != nil {
			// delete previous file and meta
			tmb := tx.Bucket(TombBucket)
			tombKey := bytes.NewBuffer(make([]byte, 0, len(key)+16))
			tombKey.Write(key)
			tombKey.WriteRune(TombTimeSeparator)
			tombKey.WriteString(strconv.FormatInt(now.UnixNano(), 36))
			if err := tmb.Put(tombKey.Bytes(), prevMeta); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *Bucket) putBlob(meta *BlobMeta, data io.Reader) error {
	p := path.Join(b.storageDir, meta.StoragePath())
	d := path.Base(p)
	_, err := os.Stat(d)
	if os.ErrNotExist == err {
		if err := os.MkdirAll(d, 0700); err != nil {
			return err
		}
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, data); err != nil {
		return err
	}
	b.tipDirCount += 1
	return nil
}

func (b *Bucket) tipDirPath() string {
	return path.Join(b.storageDir, strconv.FormatInt(b.tipID, 10))
}

func (b *Bucket) refreshTipDirStatus() error {
	b.tipMu.Lock()
	defer b.tipMu.Unlock()
	d, err := os.Open(b.tipDirPath())
	if err != nil {
		if os.IsNotExist(err) {
			b.tipDirCount = 0
			return os.MkdirAll(b.tipDirPath(), 0700)
		} else {
			return err
		}
	}
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	b.tipDirCount = len(names)
	return nil
}

func (b *Bucket) Rollover() error {
	b.tipMu.Lock()
	defer b.tipMu.Unlock()
	b.tipID = time.Now().UnixNano() / 1000 / 1000
	var buf [binary.MaxVarintLen64]byte
	if n := binary.PutVarint(buf[:], b.tipID); n <= 0 {
		return errors.New("Failed to put variant")
	}
	return b.meta.Update(func(tx *bolt.Tx) error {
		rb := tx.Bucket(RootBucket)
		if err := rb.Put(TipKey, buf[:]); err != nil {
			return err
		}
		return nil
	})
}

func (b *Bucket) initTip() error {
	// initTip only called from initializer, so need not to acquire lock here.
	e := b.meta.View(func(tx *bolt.Tx) error {
		rb := tx.Bucket(RootBucket)
		binTip := rb.Get(TipKey)
		if binTip == nil {
			return ErrEmptyTip
		} else {
			var n int
			b.tipID, n = binary.Varint(binTip)
			if n <= 0 {
				return errors.New("Failed to parse tip")
			}
		}
		return nil
	})
	if e != nil {
		if e == ErrEmptyTip {
			return b.Rollover()
		} else {
			return e
		}
	}
	return nil
}

// TODO: add clean tomb files
// TODO: add delete method
// TODO: add lifecycle method
