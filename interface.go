package gxlog

import (
	"time"
)

type LogLevel int

const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelInfo
	LevelNotice
	LevelWarning
	LevelError
	LevelCritical
	LevelFatal
	LevelOff
)

type LinkSlot int

const (
	LinkSlot0 LinkSlot = iota
	LinkSlot1
	LinkSlot2
	LinkSlot3
	LinkSlot4
	LinkSlot5
	LinkSlot6
	LinkSlot7
	MaxLinkSlot
)

type Record struct {
	Time     time.Time
	Level    LogLevel
	Pathname string
	Line     int
	Func     string
	Msg      string
}

type Formatter interface {
	Format(record *Record) []byte
}

type Writer interface {
	Write(bs []byte, record *Record)
}
