package timeline

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func mapTopicKeys(ts []*Topic) []string {
	ret := make([]string, 0, len(ts))
	for _, t := range ts {
		ret = append(ret, t.Key)
	}
	return ret
}

func mapItemCaption(ts []*Item) []string {
	ret := make([]string, 0, len(ts))
	for _, t := range ts {
		ret = append(ret, t.Caption)
	}
	return ret
}

func simpleItem(caption string) *Item {
	return &Item{
		Caption: caption,
	}
}

func TestListTopic(t *testing.T) {
	a := assert.New(t)
	storage, _ := NewStorage("memory", "")
	s := NewService(storage)
	s.NewTopic("", "/foo/bar/baz/piyo")
	s.NewTopic("", "/foo/bar/baz")
	a.Equal(mapTopicKeys(s.Topics("/foo")), []string{"/foo/bar/baz", "/foo/bar/baz/piyo"})
	// modification
	s.NewTopic("", "/foo")
	a.Equal(mapTopicKeys(s.Topics("/")), []string{"/foo", "/foo/bar/baz", "/foo/bar/baz/piyo"})
	// duplicated name will not override
	s.NewTopic("", "/foo")
	a.Equal(mapTopicKeys(s.Topics("/")), []string{"/foo", "/foo/bar/baz", "/foo/bar/baz/piyo"})
	s.NewTopic("", "/hoge")
	a.Equal(mapTopicKeys(s.Topics("/foo")), []string{"/foo", "/foo/bar/baz", "/foo/bar/baz/piyo"})
	a.Equal(mapTopicKeys(s.Topics("/")), []string{"/foo", "/foo/bar/baz", "/foo/bar/baz/piyo", "/hoge"})

	root, _ := s.Listen("/")
	hoge, _ := s.Listen("/hog")
	foo, _ := s.Listen("/fo")
	foobarbaz, _ := s.Listen("/foo/bar/baz")
	foobarbazp, _ := s.Listen("/foo/bar/baz/p")

	s.Publish("/foo/bar/baz", simpleItem("X"))
	s.Publish("/foo/bar/baz", simpleItem("Y"))
	s.Publish("/foo/bar/baz", simpleItem("Z"))
	a.Equal([]string{"X", "Y", "Z"}, mapItemCaption(root.Fetch(0)))
	a.Equal([]string{"X", "Y", "Z"}, mapItemCaption(foo.Fetch(0)))
	a.Equal([]string{"X", "Y", "Z"}, mapItemCaption(foobarbaz.Fetch(0)))
	a.Equal([]string{}, mapItemCaption(hoge.Fetch(0)))
	a.Equal([]string{}, mapItemCaption(foobarbaz.Fetch(0)))
	a.Equal([]string{}, mapItemCaption(foobarbazp.Fetch(0)))

	s.Publish("/foo/bar/baz/piyo", simpleItem("X"))
	s.Publish("/foo/bar/baz", simpleItem("Y"))
	s.Publish("/foo", simpleItem("Z"))
	a.Equal([]string{"X", "Y", "Z"}, mapItemCaption(root.Fetch(0)))
	a.Equal([]string{}, mapItemCaption(hoge.Fetch(0)))
	a.Equal([]string{"X", "Y", "Z"}, mapItemCaption(foo.Fetch(0)))
	a.Equal([]string{"X", "Y"}, mapItemCaption(foobarbaz.Fetch(0)))
	a.Equal([]string{"X"}, mapItemCaption(foobarbazp.Fetch(0)))
}
