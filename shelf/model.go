package shelf

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"time"

	"github.com/ugorji/go/codec"
)

var (
	mh codec.MsgpackHandle
)

func init() {
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.SetBytesExt(reflect.TypeOf(time.Time{}), 1, timeExt{})
}

type timeExt struct {
}

func (timeExt) WriteExt(v interface{}) (b []byte) {
	b = make([]byte, binary.MaxVarintLen64)
	switch t := v.(type) {
	case time.Time:
		binary.PutVarint(b, t.UnixNano())
		return
	case *time.Time:
		binary.PutVarint(b, t.UnixNano())
		return
	default:
		panic("Bug")
	}
}

func (timeExt) ReadExt(dst interface{}, src []byte) {
	tt := dst.(*time.Time)
	r := bytes.NewBuffer(src)
	v, err := binary.ReadVarint(r)
	if err != nil {
		panic("BUG")
	}
	*tt = time.Unix(0, v).UTC()
}
