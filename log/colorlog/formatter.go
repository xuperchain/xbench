package colorlog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	depth = 3
)

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

type Formatter interface {
	GetPrefix(lvl level) string
	Format(lvl level, v ...interface{}) []interface{}
	GetSuffix(lvl level) string
}

func header() string {
	_, fn, line, ok := runtime.Caller(depth)
	if !ok {
		fn = "???"
		line = 1
	}

	// return fmt.Sprintf("%s:%d ", filepath.Base(fn), line)
	return fmt.Sprintf("pid:%d gid:%d %s:%d ", os.Getpid(), getGID(), filepath.Base(fn), line)
}
