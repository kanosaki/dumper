package timeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	defaultSelectBufferCap = 32
	ErrNotFound            = errors.New("Not found")
)

type Storage interface {
	Insert(ctx context.Context, item ... *Item) (int64, error)
	Select(ctx context.Context, q *Query) ([]Publishing, error)
	OriginID(ctx context.Context, originName string, createIfMissing bool) (int, error)
	TopicID(ctx context.Context, key string, originID int, createIfMissing bool) (int, error)
	DB() *sql.DB
}

func NewStorage(dbType, param string) (Storage, error) {
	switch dbType {
	case "sqlite3":
		s, err := sql.Open("sqlite3", param)
		if err != nil {
			return nil, err
		}
		return NewSQLiteStorage(s)
	case "mysql":
		return nil, nil
	case "memory":
		s, err := sql.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
		if err != nil {
			return nil, err
		}
		return NewSQLiteStorage(s)
	default:
		return nil, fmt.Errorf("Unsupoorted dbType: %v", dbType)
	}
}

// Column structure is shared between sql dialects.
type SQLiteStorage struct {
	db        *sql.DB
	origins   map[string]int
	originsMu sync.Mutex
	topics    map[string]topicMeta
	topicsMu  sync.Mutex
}

type topicMeta struct {
	ID       int
	OriginID int
}

func scanOrigins(db *sql.DB) (map[string]int, error) {
	rows, err := db.Query(`SELECT id, name FROM origin`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ret := make(map[string]int)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		ret[name] = id
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return ret, nil
}

func scanTopics(db *sql.DB) (map[string]topicMeta, error) {
	rows, err := db.Query(`SELECT id, key, origin_id FROM topic`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ret := make(map[string]topicMeta)
	for rows.Next() {
		var id, originID int
		var name string
		if err := rows.Scan(&id, &name, &originID); err != nil {
			return nil, err
		}
		ret[name] = topicMeta{
			ID:       id,
			OriginID: originID,
		}
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return ret, nil
}

var sqliteInitDDLs = []string{
	`PRAGMA foreign_keys = ON`,
	`
	CREATE TABLE IF NOT EXISTS origin (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	)`,
	`
	CREATE TABLE IF NOT EXISTS topic(
		id INTEGER PRIMARY KEY,
		key TEXT NOT NULL UNIQUE,
		origin_id INTEGER NOT NULL,
		FOREIGN KEY(origin_id) REFERENCES origin(id)
	)`,
	`
	CREATE TABLE IF NOT EXISTS timeline (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		topic_id INTEGER NOT NULL,
		caption TEXT NOT NULL,
		thumbnail TEXT NOT NULL,
		origin_key INTEGER NOT NULL,
		timestamp INTEGER NOT NULL,
		meta BLOB,
		FOREIGN KEY(topic_id) REFERENCES timeline(id)
	)`,
}

func NewSQLiteStorage(s *sql.DB) (*SQLiteStorage, error) {
	for _, statement := range sqliteInitDDLs {
		_, err := s.Exec(statement)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
	}
	originMap, err := scanOrigins(s)
	if err != nil {
		return nil, err
	}
	topicMap, err := scanTopics(s)
	if err != nil {
		return nil, err
	}
	return &SQLiteStorage{
		db:      s,
		origins: originMap,
		topics:  topicMap,
	}, nil
}

func (s *SQLiteStorage) DB() *sql.DB {
	return s.db
}

func (s *SQLiteStorage) OriginID(ctx context.Context, originName string, createIfMissing bool) (int, error) {
	s.originsMu.Lock()
	defer s.originsMu.Unlock()
	originID, ok := s.origins[originName]
	if !ok {
		if !createIfMissing {
			return 0, ErrNotFound
		}
		_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO origin(name) VALUES (?)`, originName)
		if err != nil {
			return 0, err
		}
		var oid int
		row := s.db.QueryRowContext(ctx, `SELECT id from origin WHERE name = ?`, originName)
		if err := row.Scan(&oid); err != nil {
			return 0, err
		}
		originID = oid
		s.origins[originName] = originID
	}
	return originID, nil
}

func (s *SQLiteStorage) TopicID(ctx context.Context, key string, originID int, createIfMissing bool) (int, error) {
	s.topicsMu.Lock()
	defer s.topicsMu.Unlock()
	tMeta, ok := s.topics[key]
	if !ok {
		if !createIfMissing {
			return 0, ErrNotFound
		}
		_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO topic(key, origin_id) VALUES (?, ?)`, key, originID)
		if err != nil {
			return 0, err
		}
		var tid int
		row := s.db.QueryRowContext(ctx, `SELECT id from topic WHERE key = ?`, key)
		if err := row.Scan(&tid); err != nil {
			return 0, err
		}
		tMeta = topicMeta{
			ID: tid, OriginID: originID,
		}
		s.topics[key] = tMeta
	}
	return tMeta.ID, nil
}

func (s *SQLiteStorage) Insert(ctx context.Context, item ... *Item) (int64, error) {
	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO timeline(topic_id, caption, thumbnail, origin_key, timestamp, meta) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	var r sql.Result
	for _, it := range item {
		metaBytes, err := json.Marshal(it.Meta)
		timestamp := it.Timestamp.UnixNano() / int64(time.Millisecond)
		r, err = stmt.ExecContext(ctx, it.TopicID, it.Caption, it.Thumbnail, it.OriginKey, timestamp, metaBytes)
		if err != nil {
			return 0, err
		}
		if lid, err := r.LastInsertId(); err == nil {
			it.ID = lid
		}
	}
	return r.LastInsertId()
}

func (s *SQLiteStorage) Select(ctx context.Context, q *Query) ([]Publishing, error) {
	where, params, ascend := q.ToWhereClause()
	query := `SELECT
		timeline.id, timeline.topic_id, timeline.caption, timeline.thumbnail, timeline.origin_key, timeline.timestamp, timeline.meta,
		topic.key
		FROM timeline JOIN topic on timeline.topic_id = topic.id ` + where
	rows, err := s.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	var ret []Publishing
	if q.Limit > 0 {
		ret = make([]Publishing, 0, q.Limit)
	}
	defer rows.Close()
	for rows.Next() {
		var id, originKey, timestamp int64
		var topicID int
		var caption, thumbnail, topicName string
		var metaBytes []byte
		var meta map[string]interface{}
		if err := rows.Scan(&id, &topicID, &caption, &thumbnail, &originKey, &timestamp, &metaBytes, &topicName); err != nil {
			return nil, err
		}
		if len(metaBytes) > 0 {
			meta = make(map[string]interface{})
			if err := json.Unmarshal(metaBytes, &meta); err != nil {
				return nil, err
			}
		}
		json.Unmarshal(metaBytes, &meta)
		ret = append(ret, Publishing{
			Topic: topicName,
			Item: &Item{
				ID:        id,
				Caption:   caption,
				Thumbnail: thumbnail,
				Timestamp: time.Unix(0, timestamp*int64(time.Millisecond)),
				TopicID:   topicID,
				OriginKey: originKey,
				Meta:      meta,
			},
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if ascend {
		// flip array
		flipped := make([]Publishing, len(ret))
		for i := range ret {
			flipped[i] = ret[len(ret)-i-1]
		}
		return flipped, nil
	}
	return ret, nil
}
