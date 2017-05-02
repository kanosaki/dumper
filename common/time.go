package common

import "time"

// ms from unix epoch
func Timestamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func FromTimestamp(n int64) time.Time {
	return time.Unix(0, n*int64(time.Millisecond))
}
