package eslog

//go:generate stringer -type=Level levels.go
type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
)
