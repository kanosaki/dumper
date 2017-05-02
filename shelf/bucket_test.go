package shelf

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestBasicOperation(t *testing.T) {
	testDir := path.Join(os.TempDir(), "shelfroot")
	metaRoot := path.Join(testDir, "meta")
	storageRoot := path.Join(testDir, "storage")
	os.RemoveAll(testDir)
	defer func() {
		os.RemoveAll(testDir)
	}()
	fmt.Println(testDir)
	if err := os.MkdirAll(metaRoot, 0700); err != nil {
		t.Error(err)
		return
	}
	if err := os.MkdirAll(storageRoot, 0700); err != nil {
		t.Error(err)
		return
	}

	s := New(metaRoot, storageRoot)
	t.Run("Test Put/Get", func(t *testing.T) {
		firstBucket := "sample"
		firstKey := "key1"
		firstData := "hogehoge"
		b, err := s.Bucket(firstBucket)
		if err != nil {
			t.Error(err)
			return
		}
		buf := bytes.NewBufferString("hogehoge")
		if err := b.Put([]byte(firstKey), buf); err != nil {
			t.Error("Put failed:", err)
			return
		}
		rs, err := b.Get([]byte(firstKey))
		if err != nil {
			t.Error("GET faield:", err)
			return
		}
		str, err := ioutil.ReadAll(rs)
		if err != nil {
			t.Error("Read failed:", err)
			return
		}
		if string(str) != firstData {
			t.Errorf("AssertionError: expected: %v, got %v", firstData, string(str))
			return
		}
	})
}
