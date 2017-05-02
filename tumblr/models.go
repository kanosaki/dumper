package tumblr

import (
	"encoding/json"
	"fmt"
)

type Post struct {
	//
	// Base Fields
	//

	BlogName string `json:"blog_name"`
	ID       int64 `json:"id"`
	PostUrl  string `json:"post_url"`
	// text, quote, link, answer, video, audio, photo, chat
	Type string `json:"type"`
	// Seconds from epoch
	Timestamp int64 `json:"timestamp"`
	Date      string `json:"date"`
	// html or markdown
	Format      string `json:"format"`
	ReblogKey   string `string:"reblog_key"`
	Tags        []string `json:"tags"`
	Bookmarklet bool `json:"bookmarklet,omitempty"`    // appears if true
	Mobile      bool `json:"mobile,omitempty"`         // appears if true
	SourceUrl   string `json:"source_url,omitempty"`   // appears if non-empty
	SourceTitle string `json:"source_title,omitempty"` // appears if non-empty
	Liked       bool `json:"liked,omitempty"`          // appears if user is authenticated with OAuth
	// published, queued, draft, private
	State      string `json:"state"`
	TotalPosts int64 `json:"total_posts"`

	// type == 'photo'
	Photos  []Photo `json:"photos,omitmpty"`
	Caption string `json:"caption,omitempty"`
	Width   int64 `json:"width,omitempty"`
	Height  int64 `json:"height,omitempty"`

	// type == 'text'
	// not implemented
	// type == 'quote'
	// not implemented
	// type == 'link'
	// not implemented
	// type == 'answer'
	// not implemented
	// type == 'video'
	// not implemented
	// type == 'audio'
	// not implemented
	// type == 'chat'
	// not implemented
}

type Photo struct {
	Caption  string `json:"string"`
	AltSizes []AltSize `json:"alt_sizes"`
}

type AltSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	URL    string `json:"url"`
}

type Info struct {
	Name              string `json:"name"`
	Likes             int64 `json:"likes"`
	Following         int64 `json:"following"`
	DefaultPostFormat string `json:"default_post_format"`
	Blogs             []Blog `json:"blogs"`
}

type Blog json.RawMessage

type BaseResponse struct {
	Meta     MetaResponse `json:"meta"`
	Response json.RawMessage `json:"response"`
}

type MetaResponse struct {
	Status  int `json:"status"`
	Message string `json:"msg"`
}

func (mr MetaResponse) Err() error {
	if mr.Status != 200 {
		return fmt.Errorf("%v (code=%d)", mr.Message, mr.Status)
	}
	return nil
}

type LikesResponse struct {
	LikedPosts []Post `json:"liked_posts"`
	LikedCount int64 `json:"liked_count"`
}

type TagResponse []Post

type DashboardResponse struct {
	Posts []Post `json:"posts"`
}
