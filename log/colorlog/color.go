package colorlog

import (
	"fmt"
)

const (
	resetSeq  = "\033[0m"
	colourSeq = "\033[0;%dm"
)

var colour = map[level]string{
	INFO:    fmt.Sprintf(colourSeq, 94), // blue
	WARNING: fmt.Sprintf(colourSeq, 95), // pink
	ERROR:   fmt.Sprintf(colourSeq, 91), // red
	FATAL:   fmt.Sprintf(colourSeq, 91), // red
}

type ColouredFormatter struct {
}

func (f *ColouredFormatter) GetPrefix(lvl level) string {
	return colour[lvl]
}

func (f *ColouredFormatter) GetSuffix(lvl level) string {
	return resetSeq
}

func (f *ColouredFormatter) Format(lvl level, v ...interface{}) []interface{} {
	return append([]interface{}{header()}, v...)
}
