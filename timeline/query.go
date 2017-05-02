package timeline

import (
	"fmt"
	"strings"
	"time"
)

type Query struct {
	Topics []string
	MaxID  int
	MinID  int
	Limit  int
	Before time.Time
	After  time.Time
}

func (q *Query) ToWhereClause() (string, []interface{}, bool) {
	ascend := false
	orderClause := " ORDER BY timeline.id DESC"
	terms := []string{}
	params := []interface{}{}
	if len(q.Topics) > 0 {
		topicTerms := []string{}
		for _, topic := range q.Topics {
			topicTerms = append(topicTerms, "topic.key = ?")
			params = append(params, topic)
		}
		terms = append(terms,
			fmt.Sprintf("(%s)", strings.Join(topicTerms, " OR ")),
		)
	}
	if q.After.Nanosecond() > 0 {
		ascend = true
		orderClause = " ORDER BY timeline.timestamp ASC"
		afterMillisec := q.After.UnixNano() / int64(time.Millisecond)
		params = append(params, afterMillisec)
		terms = append(terms, "timeline.timestamp >= ?")
	}
	if q.Before.Nanosecond() > 0 {
		ascend = false
		orderClause = " ORDER BY timeline.timestamp DESC"
		beforeMillisec := q.Before.UnixNano() / int64(time.Millisecond)
		params = append(params, beforeMillisec)
		terms = append(terms, "timeline.timestamp <= ?")
	}
	if q.MinID > 0 {
		ascend = true
		orderClause = " ORDER BY timeline.id ASC"
		params = append(params, q.MinID)
		terms = append(terms, "timeline.id >= ?")
	}
	if q.MaxID > 0 {
		orderClause = " ORDER BY timeline.id DESC"
		ascend = false
		params = append(params, q.MaxID)
		terms = append(terms, "timeline.id <= ?")
	}
	where := strings.Join(terms, " AND ") + orderClause
	if len(terms) > 0 {
		where = "WHERE " + where
	}
	if q.Limit > 0 {
		return where + " LIMIT ?", append(params, q.Limit), ascend
	}
	return where, params, ascend
}
