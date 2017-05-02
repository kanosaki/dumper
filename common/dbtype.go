package common

type DBType int

const (
	SQLite DBType = iota
	MySQL
)

func (d DBType) String() string {
	switch d {
	case SQLite:
		return "sqlite3"
	case MySQL:
		return "mysql"
	default:
		panic("Unknown dbtype")
	}
}
