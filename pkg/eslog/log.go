package eslog

import (
	"context"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/olivere/elastic.v5"
)

var (
	FlushInterval     = 5 * time.Second
	IndexSuffixFormat = "-2006.01"
	BufferSize        = 10
)

// Logger is buffers log message
// And send it to Elasticsearch asynchronously.
type Logger struct {
	index       string
	es          *elastic.Client
	writeBuffer chan *elastic.BulkIndexRequest
	pool        sync.Pool
	log         *logrus.Entry

	// this is not atomic, but maybe enough
	closed bool
}

func New(index string, es *elastic.Client) *Logger {
	l := &Logger{
		index:       index,
		es:          es,
		writeBuffer: make(chan *elastic.BulkIndexRequest, 100),
	}
	l.pool.New = func() interface{} {
		return elastic.NewBulkIndexRequest()
	}
	go l.bufferPump()
	return l
}

type LogMessage struct {
	Message   string `json:"message"`
	Level     string `json:"level"`
	Tag       string `json:"tag"`
	Timestamp int64 `json:"timestamp"`
	Meta      map[string]interface{} `json:"data"`
}

func (l *Logger) bufferPump() {
	ticker := time.NewTicker(FlushInterval)
	buf := make([]elastic.BulkableRequest, 0, BufferSize)
	defer ticker.Stop()
LOOP:
	for !l.closed {
		select {
		case le := <-l.writeBuffer:
			buf = append(buf, le)
			if len(buf) != cap(buf) {
				continue LOOP
			}
		case <-ticker.C:
		}
		now := time.Now()
		suffix := now.Format(IndexSuffixFormat)
		br, err := l.es.Bulk().
			Index(l.index + suffix).
			Type("dumper-log").
			Add(buf...).
			Do(context.Background())
		if err != nil {
			l.log.Errorf("LogFailed: %v", err)
		}
		if br.Errors {
			l.log.Errorf("ElasticSearch Failed: %v", br.Failed())
		}
		for _, i := range buf {
			l.pool.Put(i)
		}
		buf = buf[:0]
	}
}

func (l *Logger) Entry() LogEntry {
	return LogEntry{
		logger: l,
	}
}

func (l *Logger) Append(meta map[string]interface{}, level Level, tag, s string) {
	le := l.pool.Get().(*elastic.BulkIndexRequest)
	le.Doc(LogMessage{
		Level:     level.String(),
		Tag:       tag,
		Message:   s,
		Timestamp: time.Now().UnixNano(),
		Meta:      meta,
	})
	l.writeBuffer <- le
}

func (l *Logger) Close() {
	l.closed = true
}
