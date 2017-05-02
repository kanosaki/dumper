package shelf

import (
	"fmt"
	"time"
	"github.com/ugorji/go/codec"
)

type BlobMeta struct {
	DirID     int64 `codec:"dir"`
	Filename  []byte `codec:"name"`
	CreatedAt time.Time `codec:"ctime"`
}

func (b *BlobMeta) StoragePath() string {
	tsMillisec := b.CreatedAt.UnixNano() / 1000 / 1000
	return fmt.Sprintf("%d/%d-%s", b.DirID, tsMillisec, b.Filename)
}

func (b *BlobMeta) Encode(out *[]byte) error {
	enc := codec.NewEncoderBytes(out, &mh)
	return enc.Encode(b)
}

func (b *BlobMeta) Decode(data []byte) error {
	dec := codec.NewDecoderBytes(data, &mh)
	return dec.Decode(b)
}

func ParseBlobMeta(data []byte) (*BlobMeta, error) {
	b := new(BlobMeta)
	if err := b.Decode(data); err != nil {
		return nil, err
	}
	return b, nil
}

type ChunkDir struct {
	MinTime time.Time `codec:"max_time"`
	MaxTime time.Time `codec:"min_time"`
	Count   int `codec:"count"`
}
