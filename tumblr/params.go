package tumblr

import "strconv"

type DashboardParams struct {
	Limit      int
	Offset     int
	Type       string
	SinceID    int64
	ReblogInfo bool
	NotesInfo  bool
}

func (dp *DashboardParams) Params() map[string]string {
	ret := make(map[string]string)
	if dp.Limit > 0 {
		ret["limit"] = strconv.Itoa(dp.Limit)
	}
	if dp.Offset > 0 {
		ret["offset"] = strconv.Itoa(dp.Offset)
	}
	dp.Type = dp.Type
	if dp.SinceID > 0 {
		ret["since_id"] = strconv.FormatInt(dp.SinceID, 10)
	}
	if dp.ReblogInfo {
		ret["reblog_info"] = "true"
	}
	if dp.NotesInfo {
		ret["notes_info"] = "true"
	}
	return ret
}
