package eslog

import "fmt"

type LogEntry struct {
	logger *Logger
	meta   map[string]interface{}
}

func (e LogEntry) Set(key string, value interface{}) LogEntry {
	if e.meta == nil {
		e.meta = make(map[string]interface{})
	}
	e.meta[key] = value
	return e
}

func (e LogEntry) Meta(meta map[string]interface{}) LogEntry {
	if e.meta == nil {
		e.meta = make(map[string]interface{})
	}
	for k, v := range meta {
		e.meta[k] = v
	}
	return e
}

func (e LogEntry) Debugf(tag, s string, params ... interface{}) {
	e.logger.Append(e.meta, Debug, tag, fmt.Sprintf(s, params...))
}

func (e LogEntry) Infof(tag, s string, params ... interface{}) {
	e.logger.Append(e.meta, Info, tag, fmt.Sprintf(s, params...))
}

func (e LogEntry) Warnf(tag, s string, params ... interface{}) {
	e.logger.Append(e.meta, Warn, tag, fmt.Sprintf(s, params...))
}

func (e LogEntry) Errorf(tag, s string, params ... interface{}) {
	e.logger.Append(e.meta, Error, tag, fmt.Sprintf(s, params...))
}

func (e LogEntry) Debug(tag string, params ... interface{}) {
	e.logger.Append(e.meta, Debug, tag, fmt.Sprint(params...))
}

func (e LogEntry) Info(tag string, params ... interface{}) {
	e.logger.Append(e.meta, Info, tag, fmt.Sprint(params...))
}

func (e LogEntry) Warn(tag string, params ... interface{}) {
	e.logger.Append(e.meta, Warn, tag, fmt.Sprint(params...))
}

func (e LogEntry) Error(tag string, params ... interface{}) {
	e.logger.Append(e.meta, Error, tag, fmt.Sprint(params...))
}
