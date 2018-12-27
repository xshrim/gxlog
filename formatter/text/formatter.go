package text

import (
	"regexp"
	"strings"
	"sync"

	"github.com/gxlog/gxlog"
)

var headerRegexp = regexp.MustCompile("{{([^:%]*?)(?::([^%]*?))?(%.*?)?}}")

type Formatter struct {
	header      string
	minBufSize  int
	enableColor bool

	colorMgr  *colorMgr
	appenders []*headerAppender
	suffix    string

	lock sync.Mutex
}

func New(config *Config) *Formatter {
	if config.MinBufSize < 0 {
		panic("formatter/text.New: Config.MinBufSize must not be negative")
	}
	formatter := &Formatter{
		minBufSize:  config.MinBufSize,
		enableColor: config.EnableColor,
		colorMgr:    newColorMgr(),
	}
	formatter.SetHeader(config.Header)
	formatter.MapColors(config.ColorMap)
	return formatter
}

func (formatter *Formatter) Header() string {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	return formatter.header
}

func (formatter *Formatter) SetHeader(header string) {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	formatter.header = header
	formatter.appenders = formatter.appenders[:0]
	var staticText string
	for header != "" {
		indexes := headerRegexp.FindStringSubmatchIndex(header)
		if indexes == nil {
			break
		}
		begin, end := indexes[0], indexes[1]
		staticText += header[:begin]
		element, property, fmtspec := extractElement(indexes, header)
		if formatter.addAppender(element, property, fmtspec, staticText) {
			staticText = ""
		}
		header = header[end:]
	}
	formatter.suffix = staticText + header
}

func (formatter *Formatter) MinBufSize() int {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	return formatter.minBufSize
}

func (formatter *Formatter) SetMinBufSize(size int) {
	if size < 0 {
		panic("formatter/text.SetMinBufSize: size must not be negative")
	}

	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	formatter.minBufSize = size
}

func (formatter *Formatter) ColorEnabled() bool {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	return formatter.enableColor
}

func (formatter *Formatter) EnableColor() {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	formatter.enableColor = true
}

func (formatter *Formatter) DisableColor() {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	formatter.enableColor = false
}

func (formatter *Formatter) Color(level gxlog.Level) ColorID {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	return formatter.colorMgr.Color(level)
}

func (formatter *Formatter) SetColor(level gxlog.Level, color ColorID) {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	formatter.colorMgr.SetColor(level, color)
}

func (formatter *Formatter) MapColors(colorMap map[gxlog.Level]ColorID) {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	formatter.colorMgr.MapColors(colorMap)
}

func (formatter *Formatter) MarkedColor() ColorID {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	return formatter.colorMgr.MarkedColor()
}

func (formatter *Formatter) SetMarkedColor(color ColorID) {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	formatter.colorMgr.SetMarkedColor(color)
}

func (formatter *Formatter) Format(record *gxlog.Record) []byte {
	formatter.lock.Lock()
	defer formatter.lock.Unlock()

	var left, right []byte
	if formatter.enableColor {
		if record.Aux.Marked {
			left, right = formatter.colorMgr.MarkedColorEars()
		} else {
			left, right = formatter.colorMgr.ColorEars(record.Level)
		}
	}

	buf := make([]byte, 0, formatter.minBufSize)
	buf = append(buf, left...)
	for _, appender := range formatter.appenders {
		buf = appender.AppendHeader(buf, record)
	}
	buf = append(buf, formatter.suffix...)
	buf = append(buf, right...)

	return buf
}

func (formatter *Formatter) addAppender(element, property, fmtspec, staticText string) bool {
	appender := newHeaderAppender(element, property, fmtspec, staticText)
	if appender == nil {
		return false
	}
	formatter.appenders = append(formatter.appenders, appender)
	return true
}

func extractElement(indexes []int, header string) (element, property, fmtspec string) {
	element = strings.ToLower(getField(header, indexes[2], indexes[3]))
	property = getField(header, indexes[4], indexes[5])
	fmtspec = getField(header, indexes[6], indexes[7])
	if fmtspec == "%" {
		fmtspec = ""
	}
	return element, property, fmtspec
}

func getField(str string, begin, end int) string {
	if begin < end {
		return strings.TrimSpace(str[begin:end])
	}
	return ""
}
