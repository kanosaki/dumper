package pixiv

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/kanosaki/dumper/common"
	"github.com/kanosaki/gopixiv"
)

// PRIMARY KEY (Pixiv ID, FetchTime)
type WorksMapper struct {
	db *sql.DB
}

type WorkRecord struct {
	Timestamp time.Time // updated time
	Item      *pixiv.Item
}

func (wm *WorksMapper) initSQLite(ctx context.Context) error {
	_, err := wm.db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS pixiv_work (
			id INTEGER NOT NULL,
			timestamp INTEGER NOT NULL,
			body BLOB NOT NULL,
			PRIMARY KEY (id, timestamp)
		)`,
	)
	return err
}

func NewWorksMapper(ctx context.Context, db *sql.DB, dbtype common.DBType) (*WorksMapper, error) {
	wm := &WorksMapper{
		db: db,
	}
	switch dbtype {
	case common.SQLite:
		if err := wm.initSQLite(ctx); err != nil {
			return nil, err
		}
	default:
		panic("Unsupported DBType")
	}
	return wm, nil
}

func (wm *WorksMapper) InsertBulk(ctx context.Context, items []*pixiv.Item, timestamp time.Time) error {
	q := `INSERT INTO pixiv_work(id, timestamp, body) VALUES (?, ?, ?)`
	stmt, err := wm.db.PrepareContext(ctx, q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, item := range items {
		body, err := json.Marshal(item)
		if err != nil {
			return err
		}
		_, err = stmt.ExecContext(ctx, item.ID, common.Timestamp(timestamp), body)
		if err != nil {
			return err
		}
	}
	return nil
}

func (wm *WorksMapper) Get(ctx context.Context, id int64) (WorkRecord, error) {
	q := `SELECT (timestamp, body) FROM pixiv_work WHERE id = ? ORDER BY timestamp DESC LIMIT 1`
	row := wm.db.QueryRowContext(ctx, q, id)
	var ts int64
	var bodyBytes []byte
	if err := row.Scan(&ts, &bodyBytes); err != nil {
		return WorkRecord{}, err
	}
	item := &pixiv.Item{}
	if err := json.Unmarshal(bodyBytes, &item); err != nil {
		return WorkRecord{}, err
	}
	return WorkRecord{
		Timestamp: common.FromTimestamp(ts),
		Item:      item,
	}, nil
}
